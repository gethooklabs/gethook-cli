package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/gethook/gethook-cli/internal/output"
	"github.com/gethook/gethook-cli/internal/proxy"
	"github.com/gethook/gethook-cli/internal/tunnel"
)

func newListenCmd() *cobra.Command {
	var (
		forwardTo  string
		sourceID   string
		filter     string
		noTunnel   bool
	)

	cmd := &cobra.Command{
		Use:   "listen",
		Short: "Receive webhook events and forward them to a local server",
		Long: `Start a live tunnel to your GetHook source and stream incoming events.

Without --forward-to, events are printed to the terminal only.
With --forward-to, each event is also POSTed to your local server.

Examples:
  gethook listen
  gethook listen --forward-to http://localhost:3000/webhooks
  gethook listen --source src_abc123 --filter "stripe.*"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			c := requireAuth()

			ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
			defer stop()

			// Resolve or create source.
			var src interface{ GetPathToken() string }
			var ingestURL string

			if sourceID != "" {
				s, err := c.GetSource(ctx, sourceID)
				if err != nil {
					return fmt.Errorf("get source: %w", err)
				}
				ingestURL = fmt.Sprintf("%s/%s", cfg.Ingest, s.PathToken)
				sourceID = s.ID
			} else {
				// Create a temporary source for this session.
				name := fmt.Sprintf("cli-listen-%d", time.Now().Unix())
				s, err := c.CreateSource(ctx, name)
				if err != nil {
					return fmt.Errorf("create temporary source: %w", err)
				}
				ingestURL = fmt.Sprintf("%s/%s", cfg.Ingest, s.PathToken)
				sourceID = s.ID
			}
			_ = src // suppress unused warning

			// Print the session banner.
			printListenBanner(ingestURL, forwardTo)

			if noTunnel {
				output.Muted("  (print-only mode — not forwarding)")
				fmt.Fprintln(os.Stderr)
			}

			// Start event relay.
			eventCh := make(chan tunnel.Event, 32)
			relay := tunnel.New(c, sourceID)

			go func() {
				if err := relay.Run(ctx, eventCh); err != nil && ctx.Err() == nil {
					output.Error("relay error: " + err.Error())
				}
			}()

			// Print header row.
			fmt.Fprintf(os.Stdout, "\n  %s  %s  %s  %s\n",
				lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280")).Render("TIME    "),
				lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280")).Render("EVENT TYPE                    "),
				lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280")).Render("STATUS"),
				lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280")).Render("LATENCY"),
			)
			fmt.Fprintln(os.Stdout, "  "+lipgloss.NewStyle().Foreground(lipgloss.Color("#374151")).Render("─────────────────────────────────────────────────────────"))

			eventCount := 0

			for {
				select {
				case <-ctx.Done():
					fmt.Fprintln(os.Stderr)
					if eventCount == 0 {
						output.Muted("  No events received this session.")
					} else {
						output.Muted(fmt.Sprintf("  Session ended — %d event(s) received.", eventCount))
					}
					return nil

				case ev := <-eventCh:
					if filter != "" && !matchGlob(filter, ev.EventType) {
						continue
					}
					eventCount++

					statusStr := ev.Status
					durationMs := 0

					if !noTunnel && forwardTo != "" {
						body, _ := json.Marshal(ev.Payload)
						result := proxy.Forward(ctx, forwardTo, ev.Headers, body)
						if result.Err != nil {
							statusStr = "forward-error"
							output.Warn("  forward error: " + result.Err.Error())
						} else {
							statusStr = fmt.Sprintf("%d", result.StatusCode)
							durationMs = result.DurationMs
						}
					}

					output.EventLine(os.Stdout, ev.ReceivedAt, "POST", ev.EventType, statusStr, durationMs)
				}
			}
		},
	}

	cmd.Flags().StringVar(&forwardTo, "forward-to", "", "Local URL to forward events to (e.g. http://localhost:3000/webhooks)")
	cmd.Flags().StringVar(&sourceID, "source", "", "Use an existing source ID instead of creating a temporary one")
	cmd.Flags().StringVar(&filter, "filter", "", "Only show/forward events matching this glob pattern (e.g. 'stripe.*')")
	cmd.Flags().BoolVar(&noTunnel, "no-tunnel", false, "Print events only, don't forward")

	return cmd
}

func printListenBanner(ingestURL, forwardTo string) {
	lines := []string{
		output.StyleSuccess.Render("✓") + " Tunnel active",
		"  " + output.StyleURL.Render(ingestURL),
	}
	if forwardTo != "" {
		lines = append(lines,
			output.StylePrimary.Render("→") + " Forwarding to",
			"  "+forwardTo,
		)
	}
	lines = append(lines, "")
	lines = append(lines, output.StyleMuted.Render("  Paste the tunnel URL into your webhook provider."))
	lines = append(lines, output.StyleMuted.Render("  Press Ctrl+C to stop."))
	output.Banner(lines...)
}

// matchGlob is a simple glob matcher supporting '*' wildcard.
func matchGlob(pattern, s string) bool {
	if pattern == "*" || pattern == "" {
		return true
	}
	// Simple prefix/suffix/exact matching.
	if len(pattern) > 0 && pattern[len(pattern)-1] == '*' {
		prefix := pattern[:len(pattern)-1]
		return len(s) >= len(prefix) && s[:len(prefix)] == prefix
	}
	if len(pattern) > 0 && pattern[0] == '*' {
		suffix := pattern[1:]
		return len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix
	}
	return pattern == s
}

// ── spinner model (used by other commands) ─────────────────────────────────

type spinnerModel struct {
	spinner  spinner.Model
	message  string
	quitting bool
}

func newSpinnerModel(msg string) spinnerModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#7C3AED"))
	return spinnerModel{spinner: s, message: msg}
}

func (m spinnerModel) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m spinnerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case tea.KeyMsg:
		m.quitting = true
		return m, tea.Quit
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m spinnerModel) View() string {
	if m.quitting {
		return ""
	}
	return fmt.Sprintf("  %s %s\n", m.spinner.View(), m.message)
}
