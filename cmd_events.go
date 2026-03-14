package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/gethooklabs/gethook-cli/internal/api"
	"github.com/gethooklabs/gethook-cli/internal/output"
)

func newEventsCmd() *cobra.Command {
	var (
		tail      bool
		sourceID  string
		status    string
		direction string
		limit     int
	)

	cmd := &cobra.Command{
		Use:   "events",
		Short: "List or stream recent events",
		Long: `List recent webhook events from your GetHook account.

Use --tail to stream new events in real time (like tail -f).
Use --json for pipe-friendly raw JSON output.

Examples:
  gethook events
  gethook events --tail
  gethook events --status dead_letter --limit 10
  gethook events --source src_abc123 --tail`,
		RunE: func(cmd *cobra.Command, args []string) error {
			c := requireAuth()

			if tail {
				return streamEvents(c, sourceID, status, direction)
			}
			return listEvents(c, sourceID, status, direction, limit)
		},
	}

	cmd.Flags().BoolVarP(&tail, "tail", "f", false, "Stream new events in real time")
	cmd.Flags().StringVar(&sourceID, "source", "", "Filter by source ID")
	cmd.Flags().StringVar(&status, "status", "", "Filter by status (received, queued, delivered, dead_letter, …)")
	cmd.Flags().StringVar(&direction, "direction", "", "Filter by direction (inbound, outbound)")
	cmd.Flags().IntVar(&limit, "limit", 20, "Number of events to show")

	return cmd
}

func listEvents(c *api.Client, sourceID, status, direction string, limit int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	evts, err := c.ListEvents(ctx, api.ListEventsParams{
		SourceID: sourceID,
		Status:   status,
		Limit:    limit,
	})
	if err != nil {
		return fmt.Errorf("list events: %w", err)
	}

	if jsonOut {
		b, _ := json.MarshalIndent(evts, "", "  ")
		fmt.Println(string(b))
		return nil
	}

	if len(evts) == 0 {
		output.Muted("No events found.")
		return nil
	}

	rows := make([]output.TableRow, len(evts))
	for i, e := range evts {
		srcID := ""
		if e.SourceID != nil {
			srcID = *e.SourceID
		}
		rows[i] = output.TableRow{
			e.ID,
			e.EventTypeStr(),
			e.Status,
			e.Direction,
			srcID,
			fmt.Sprintf("%d", e.AttemptsCount),
			e.ReceivedAt.Format("2006-01-02 15:04:05"),
		}
	}

	output.PrintTable(
		[]string{"ID", "EVENT TYPE", "STATUS", "DIRECTION", "SOURCE", "ATTEMPTS", "RECEIVED AT"},
		rows,
	)
	return nil
}

func streamEvents(c *api.Client, sourceID, status, direction string) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	output.Info("Streaming events (Ctrl+C to stop)...")
	fmt.Fprintln(os.Stderr)

	seen := map[string]bool{}
	initial, err := c.ListEvents(ctx, api.ListEventsParams{
		SourceID: sourceID,
		Status:   status,
		Limit:    50,
	})
	if err != nil {
		return fmt.Errorf("initial list: %w", err)
	}
	for _, e := range initial {
		seen[e.ID] = true
	}

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			output.Muted("\nStream stopped.")
			return nil
		case <-ticker.C:
			evts, err := c.ListEvents(ctx, api.ListEventsParams{
				SourceID: sourceID,
				Status:   status,
				Limit:    20,
			})
			if err != nil {
				output.Warn("poll error: " + err.Error())
				continue
			}
			for i := len(evts) - 1; i >= 0; i-- {
				e := evts[i]
				if seen[e.ID] {
					continue
				}
				seen[e.ID] = true

				if jsonOut {
					b, _ := json.MarshalIndent(e, "", "  ")
					fmt.Println(string(b))
				} else {
					output.EventLine(os.Stdout, e.ReceivedAt, "POST", e.EventTypeStr(), e.Status, 0)
				}
			}
		}
	}
}
