package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/gethook/gethook-cli/internal/output"
)

func newSourcesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sources",
		Short: "Manage webhook sources",
		Long:  "List, create, and delete webhook sources.",
	}

	cmd.AddCommand(
		newSourcesListCmd(),
		newSourcesCreateCmd(),
		newSourcesDeleteCmd(),
	)
	return cmd
}

func newSourcesListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all sources",
		RunE: func(cmd *cobra.Command, args []string) error {
			c := requireAuth()
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()

			sources, err := c.ListSources(ctx)
			if err != nil {
				return err
			}

			if jsonOut {
				b, _ := json.MarshalIndent(sources, "", "  ")
				fmt.Println(string(b))
				return nil
			}

			if len(sources) == 0 {
				output.Muted("No sources found. Create one with: gethook sources create <name>")
				return nil
			}

			rows := make([]output.TableRow, len(sources))
			for i, s := range sources {
				rows[i] = output.TableRow{s.ID, s.Name, s.PathToken, s.Status, s.CreatedAt.Format("2006-01-02")}
			}
			output.PrintTable([]string{"ID", "NAME", "TOKEN", "STATUS", "CREATED"}, rows)
			return nil
		},
	}
}

func newSourcesCreateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "create <name>",
		Short: "Create a new source",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c := requireAuth()
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()

			s, err := c.CreateSource(ctx, args[0])
			if err != nil {
				return err
			}

			if jsonOut {
				b, _ := json.MarshalIndent(s, "", "  ")
				fmt.Println(string(b))
				return nil
			}

			output.Success(fmt.Sprintf("Created source %s", s.ID))
			fmt.Printf("  Name:       %s\n", s.Name)
			fmt.Printf("  Ingest URL: %s/%s\n", cfg.Ingest, s.PathToken)
			return nil
		},
	}
}

func newSourcesDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a source",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c := requireAuth()
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()

			if err := c.DeleteSource(ctx, args[0]); err != nil {
				return err
			}
			output.Success(fmt.Sprintf("Deleted source %s", args[0]))
			return nil
		},
	}
}
