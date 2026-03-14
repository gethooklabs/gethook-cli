package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/gethooklabs/gethook-cli/internal/output"
)

func newInspectCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "inspect <event-id>",
		Short: "Show full details of an event",
		Long: `Show the complete details of a webhook event including:
  - Full payload body
  - Original headers
  - All delivery attempts with status codes

Examples:
  gethook inspect evt_abc123
  gethook inspect evt_abc123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c := requireAuth()

			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()

			detail, err := c.GetEvent(ctx, args[0])
			if err != nil {
				return fmt.Errorf("fetch event: %w", err)
			}

			if jsonOut {
				b, _ := json.MarshalIndent(detail, "", "  ")
				fmt.Println(string(b))
				return nil
			}

			ev := detail.Event
			if ev == nil {
				return fmt.Errorf("empty event response")
			}

			// ── Header ────────────────────────────────────────────────────
			statusColor := lipgloss.Color("#22C55E")
			if ev.Status == "dead_letter" {
				statusColor = lipgloss.Color("#EF4444")
			} else if ev.Status != "delivered" {
				statusColor = lipgloss.Color("#F59E0B")
			}
			statusStyle := lipgloss.NewStyle().Foreground(statusColor).Bold(true)

			fmt.Println()
			fmt.Printf("  %s  %s\n", output.StyleBold.Render("Event "), output.StyleMuted.Render(ev.ID))
			fmt.Printf("  %s  %s\n", output.StyleBold.Render("Type  "), ev.EventTypeStr())
			fmt.Printf("  %s  %s\n",
				output.StyleBold.Render("Status"),
				statusStyle.Render(fmt.Sprintf("%s (%d attempt(s))", ev.Status, ev.AttemptsCount)),
			)
			fmt.Printf("  %s  %s\n", output.StyleBold.Render("Time  "), ev.ReceivedAt.Format("2006-01-02 15:04:05 UTC"))
			if ev.SourceID != nil {
				fmt.Printf("  %s  %s\n", output.StyleBold.Render("Source"), *ev.SourceID)
			}

			// ── Headers ───────────────────────────────────────────────────
			if len(ev.Headers) > 0 {
				output.Section("Headers")
				for k, v := range ev.Headers {
					fmt.Printf("  %s: %v\n", output.StyleMuted.Render(k), v)
				}
			}

			// ── Payload ───────────────────────────────────────────────────
			output.Section("Payload")
			if ev.Body != "" {
				// Pretty-print if valid JSON, otherwise raw.
				var pretty interface{}
				if json.Unmarshal([]byte(ev.Body), &pretty) == nil {
					fmt.Println(output.PrettyJSON(pretty))
				} else {
					fmt.Println(ev.Body)
				}
			} else {
				output.Muted("  (no body)")
			}

			// ── Delivery Attempts ─────────────────────────────────────────
			if len(detail.Attempts) > 0 {
				output.Section("Delivery Attempts")
				for _, a := range detail.Attempts {
					outcomeStyle := output.StyleSuccess
					if a.Outcome != "success" {
						outcomeStyle = output.StyleError
					}

					statusCode := ""
					if a.ResponseStatus != nil {
						statusCode = fmt.Sprintf("  HTTP %d", *a.ResponseStatus)
					}

					latency := ""
					if a.LatencyMS != nil {
						latency = output.StyleMuted.Render(fmt.Sprintf("(%dms)", *a.LatencyMS))
					}

					fmt.Printf("  #%d  %s  %s%s  %s\n",
						a.AttemptNumber,
						output.StyleMuted.Render(a.StartedAt.Format("15:04:05")),
						outcomeStyle.Render(a.Outcome),
						statusCode,
						latency,
					)

					if a.ResponseBody != nil && len(*a.ResponseBody) > 0 && len(*a.ResponseBody) < 500 {
						truncated := *a.ResponseBody
						if len(truncated) > 200 {
							truncated = truncated[:200] + "…"
						}
						fmt.Printf("       %s\n", output.StyleMuted.Render(strings.TrimSpace(truncated)))
					}
					if a.ErrorMessage != nil {
						fmt.Printf("       %s\n", output.StyleError.Render(*a.ErrorMessage))
					}
				}
			}

			fmt.Println()
			return nil
		},
	}
}
