package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/gethook/gethook-cli/internal/output"
	"github.com/gethook/gethook-cli/internal/proxy"
)

func newReplayCmd() *cobra.Command {
	var (
		forwardTo string
		dryRun    bool
	)

	cmd := &cobra.Command{
		Use:   "replay <event-id>",
		Short: "Replay a past event",
		Long: `Replay a specific event from your GetHook account.

By default, the event is replayed to its original destination.
Use --forward-to to override the destination URL (e.g. your local server).
Use --dry-run to inspect what would be sent without sending anything.

Examples:
  gethook replay evt_abc123
  gethook replay evt_abc123 --forward-to http://localhost:3000/webhooks
  gethook replay evt_abc123 --dry-run`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c := requireAuth()
			eventID := args[0]

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			// Fetch the full event for display and local forwarding.
			ev, err := c.GetEvent(ctx, eventID)
			if err != nil {
				return fmt.Errorf("fetch event: %w", err)
			}

			output.Info(fmt.Sprintf("Replaying %s", ev.ID))
			output.Muted(fmt.Sprintf("  Type:     %s", ev.EventType))
			output.Muted(fmt.Sprintf("  Status:   %s", ev.Status))
			output.Muted(fmt.Sprintf("  Received: %s", ev.ReceivedAt.Format("2006-01-02 15:04:05")))
			fmt.Println()

			if dryRun {
				output.Warn("Dry run — no request sent.")
				if ev.Payload != nil {
					output.Section("Payload")
					fmt.Println(output.PrettyJSON(ev.Payload))
				}
				return nil
			}

			if forwardTo != "" {
				// Forward locally without going through GetHook.
				body, _ := json.Marshal(ev.Payload)
				output.Info(fmt.Sprintf("→ POST %s", forwardTo))
				result := proxy.Forward(ctx, forwardTo, ev.Headers, body)
				if result.Err != nil {
					return fmt.Errorf("forward error: %w", result.Err)
				}
				statusLine := fmt.Sprintf("%d", result.StatusCode)
				if result.StatusCode >= 200 && result.StatusCode < 300 {
					output.Success(fmt.Sprintf("← %s (%dms)", statusLine, result.DurationMs))
				} else {
					output.Error(fmt.Sprintf("← %s (%dms)", statusLine, result.DurationMs))
				}
				return nil
			}

			// Use the GetHook replay API (re-delivers to original destination).
			if err := c.ReplayEvent(ctx, eventID); err != nil {
				return fmt.Errorf("replay: %w", err)
			}
			output.Success("Replay queued — event will be delivered to its original destination.")
			return nil
		},
	}

	cmd.Flags().StringVar(&forwardTo, "forward-to", "", "Override destination URL (e.g. http://localhost:3000/webhooks)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be sent without sending")

	return cmd
}
