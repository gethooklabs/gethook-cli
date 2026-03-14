package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/gethook/gethook-cli/internal/api"
	"github.com/gethook/gethook-cli/internal/config"
	"github.com/gethook/gethook-cli/internal/output"
)

var (
	cfg     *config.Config
	client  *api.Client
	jsonOut bool
)

var rootCmd = &cobra.Command{
	Use:   "gethook",
	Short: "GetHook CLI — receive, debug, and replay webhooks locally",
	Long: lipgloss.NewStyle().Foreground(lipgloss.Color("#7C3AED")).Bold(true).Render(`
   ██████╗ ███████╗████████╗██╗  ██╗ ██████╗  ██████╗ ██╗  ██╗
  ██╔════╝ ██╔════╝╚══██╔══╝██║  ██║██╔═══██╗██╔═══██╗██║ ██╔╝
  ██║  ███╗█████╗     ██║   ███████║██║   ██║██║   ██║█████╔╝
  ██║   ██║██╔══╝     ██║   ██╔══██║██║   ██║██║   ██║██╔═██╗
  ╚██████╔╝███████╗   ██║   ██║  ██║╚██████╔╝╚██████╔╝██║  ██╗
   ╚═════╝ ╚══════╝   ╚═╝   ╚═╝  ╚═╝ ╚═════╝  ╚═════╝ ╚═╝  ╚═╝
`) + `
Receive, debug, and replay webhooks locally — in seconds.

  gethook listen --forward-to http://localhost:3000/webhooks
  gethook events --tail
  gethook trigger stripe payment_intent.succeeded

Docs: https://docs.gethook.dev/cli`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().BoolVar(&jsonOut, "json", false, "Output raw JSON")

	rootCmd.AddCommand(
		newListenCmd(),
		newEventsCmd(),
		newReplayCmd(),
		newInspectCmd(),
		newTriggerCmd(),
		newLoginCmd(),   // flags defined inside newLoginCmd
		newLogoutCmd(),
		newSourcesCmd(),
		newDestinationsCmd(),
		newVersionCmd(),
	)
}

func initConfig() {
	var err error
	cfg, err = config.Load()
	if err != nil {
		output.Warn("could not load config: " + err.Error())
		cfg = &config.Config{
			APIBase: config.DefaultAPIBase,
			Ingest:  config.DefaultIngest,
		}
	}
	if cfg.APIKey != "" {
		client = api.New(cfg.APIBase, cfg.APIKey)
	}
}

// requireAuth returns a non-nil client or prints an error and exits.
func requireAuth() *api.Client {
	if client == nil {
		output.Error("Not logged in. Run: gethook login")
		os.Exit(1)
	}
	return client
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print CLI version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("gethook version " + version)
		},
	}
}

// version is set at build time via ldflags.
var version = "dev"
