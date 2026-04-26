package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/gethooklabs/gethook-cli/internal/output"
)

func newDestinationsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "destinations",
		Short: "Manage webhook destinations",
		Long:  "List, create, and delete webhook destinations.",
	}

	cmd.AddCommand(
		newDestinationsListCmd(),
		newDestinationsCreateCmd(),
		newDestinationsDeleteCmd(),
	)
	return cmd
}

func newDestinationsListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all destinations",
		RunE: func(cmd *cobra.Command, args []string) error {
			c := requireAuth()
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()

			dests, err := c.ListDestinations(ctx)
			if err != nil {
				return err
			}

			if jsonOut {
				b, _ := json.MarshalIndent(dests, "", "  ")
				fmt.Println(string(b))
				return nil
			}

			if len(dests) == 0 {
				output.Muted("No destinations found.")
				return nil
			}

			rows := make([]output.TableRow, len(dests))
			for i, d := range dests {
				rows[i] = output.TableRow{
					d.ID, d.Name, d.URL,
					fmt.Sprintf("%ds", d.TimeoutSeconds),
					d.CreatedAt.Format("2006-01-02"),
				}
			}
			output.PrintTable([]string{"ID", "NAME", "URL", "TIMEOUT", "CREATED"}, rows)
			return nil
		},
	}
}

func newDestinationsCreateCmd() *cobra.Command {
	var timeout int

	cmd := &cobra.Command{
		Use:   "create <name> <url>",
		Short: "Create a new destination",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			c := requireAuth()
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()

			d, err := c.CreateDestination(ctx, args[0], args[1], timeout)
			if err != nil {
				return err
			}

			if jsonOut {
				b, _ := json.MarshalIndent(d, "", "  ")
				fmt.Println(string(b))
				return nil
			}

			output.Success(fmt.Sprintf("Created destination %s", d.ID))
			fmt.Printf("  Name: %s\n", d.Name)
			fmt.Printf("  URL:  %s\n", d.URL)
			return nil
		},
	}

	cmd.Flags().IntVar(&timeout, "timeout", 30, "Request timeout in seconds")
	return cmd
}

func newDestinationsDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a destination",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c := requireAuth()
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()

			if err := c.DeleteDestination(ctx, args[0]); err != nil {
				return err
			}
			output.Success(fmt.Sprintf("Deleted destination %s", args[0]))
			return nil
		},
	}
}
