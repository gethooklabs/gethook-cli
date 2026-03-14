package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/spf13/cobra"

	fixtures "github.com/gethook/gethook-cli/internal/fixtures"
	"github.com/gethook/gethook-cli/internal/output"
	"github.com/gethook/gethook-cli/internal/proxy"
)

func newTriggerCmd() *cobra.Command {
	var (
		forwardTo string
		dataJSON  string
		listAll   bool
	)

	cmd := &cobra.Command{
		Use:   "trigger <provider> <event-type>",
		Short: "Send a realistic test webhook payload",
		Long: `Send a realistic fake webhook payload for a known provider event type.
No real API account needed — payloads are bundled in the CLI.

If --forward-to is not set, the event is sent to your GetHook account's
ingest URL (requires login). Use --forward-to for pure local testing.

Examples:
  gethook trigger stripe payment_intent.succeeded
  gethook trigger stripe charge.failed --forward-to http://localhost:3000/webhooks
  gethook trigger github push --data '{"ref":"refs/heads/my-branch"}'
  gethook trigger --list`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if listAll {
				return printFixtureList()
			}

			if len(args) < 2 {
				return fmt.Errorf("usage: gethook trigger <provider> <event-type>\nRun `gethook trigger --list` to see available providers and event types")
			}

			provider := strings.ToLower(args[0])
			eventType := args[1]

			// Parse --data overrides.
			var overrides map[string]interface{}
			if dataJSON != "" {
				if err := json.Unmarshal([]byte(dataJSON), &overrides); err != nil {
					return fmt.Errorf("parse --data JSON: %w", err)
				}
			}

			payload, err := fixtures.Load(provider, eventType, overrides)
			if err != nil {
				return err
			}

			target := forwardTo
			if target == "" {
				// Fall back to the GetHook ingest endpoint if logged in.
				if client == nil {
					return fmt.Errorf("provide --forward-to <url> or run `gethook login` to use your ingest endpoint")
				}
				// Get the first source.
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()
				sources, err := client.ListSources(ctx)
				if err != nil || len(sources) == 0 {
					return fmt.Errorf("no sources found — provide --forward-to <url> or create a source first")
				}
				target = fmt.Sprintf("%s/%s", cfg.Ingest, sources[0].PathToken)
			}

			output.Info(fmt.Sprintf("Triggering %s/%s", provider, eventType))
			output.Muted(fmt.Sprintf("  → POST %s", target))

			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()

			result := sendPayload(ctx, target, payload)
			if result.Err != nil {
				return fmt.Errorf("trigger failed: %w", result.Err)
			}

			statusStr := fmt.Sprintf("%d", result.StatusCode)
			if result.StatusCode >= 200 && result.StatusCode < 300 {
				output.Success(fmt.Sprintf("← %s (%dms)", statusStr, result.DurationMs))
			} else {
				output.Error(fmt.Sprintf("← %s (%dms)", statusStr, result.DurationMs))
			}

			if jsonOut {
				var pretty interface{}
				json.Unmarshal(payload, &pretty)
				b, _ := json.MarshalIndent(pretty, "", "  ")
				fmt.Println(string(b))
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&forwardTo, "forward-to", "", "Target URL (local server or any HTTP endpoint)")
	cmd.Flags().StringVar(&dataJSON, "data", "", "JSON object with fields to override in the fixture payload")
	cmd.Flags().BoolVar(&listAll, "list", false, "List all available providers and event types")

	return cmd
}

func sendPayload(ctx context.Context, targetURL string, payload []byte) proxy.ForwardResult {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, targetURL, bytes.NewReader(payload))
	if err != nil {
		return proxy.ForwardResult{Err: fmt.Errorf("build request: %w", err)}
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-GetHook-Trigger", "1")

	httpClient := &http.Client{Timeout: 15 * time.Second}
	start := time.Now()
	resp, err := httpClient.Do(req)
	durationMs := int(time.Since(start).Milliseconds())
	if err != nil {
		return proxy.ForwardResult{DurationMs: durationMs, Err: err}
	}
	defer resp.Body.Close()

	return proxy.ForwardResult{StatusCode: resp.StatusCode, DurationMs: durationMs}
}

func printFixtureList() error {
	providers := fixtures.Providers()
	if len(providers) == 0 {
		output.Muted("No fixtures available.")
		return nil
	}

	fmt.Println()
	fmt.Println(output.StyleBold.Render("  Available providers and event types:"))
	fmt.Println()

	for _, p := range providers {
		fmt.Println("  " + output.StylePrimary.Render(p))
		types, err := fixtures.EventTypes(p)
		if err != nil {
			continue
		}
		for _, t := range types {
			fmt.Printf("    %s gethook trigger %s %s\n",
				output.StyleMuted.Render("→"),
				p,
				t,
			)
		}
		fmt.Println()
	}
	return nil
}
