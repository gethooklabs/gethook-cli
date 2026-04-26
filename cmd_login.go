package main

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/gethooklabs/gethook-cli/internal/api"
	"github.com/gethooklabs/gethook-cli/internal/config"
	"github.com/gethooklabs/gethook-cli/internal/output"
)

func newLoginCmd() *cobra.Command {
	var apiKeyFlag string

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Authenticate with your GetHook account",
		Long: `Authenticate the CLI with your GetHook account.

You can log in via the browser (recommended) or paste an API key directly:

  gethook login            — opens browser for authentication
  gethook login --key hk_… — use an existing API key`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if apiKeyFlag != "" {
				return loginWithKey(apiKeyFlag)
			}
			return loginWithBrowser()
		},
	}

	cmd.Flags().StringVar(&apiKeyFlag, "key", "", "API key to use directly (skips browser)")
	return cmd
}

func loginWithKey(key string) error {
	if !strings.HasPrefix(key, "hk_") {
		output.Warn("API key should start with 'hk_'")
	}

	c := api.New(cfg.APIBase, key)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if _, err := c.ListSources(ctx); err != nil {
		return fmt.Errorf("invalid API key: %w", err)
	}

	if err := config.SaveAPIKey(key); err != nil {
		return fmt.Errorf("saving key: %w", err)
	}

	output.Success("Logged in successfully")
	return nil
}

func loginWithBrowser() error {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("start callback server: %w", err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	callbackURL := fmt.Sprintf("http://localhost:%d/callback", port)

	appBase := strings.Replace(cfg.APIBase, "api.", "app.", 1)
	authURL := fmt.Sprintf("%s/cli-auth?callback=%s", appBase, callbackURL)

	keyCh := make(chan string, 1)
	errCh := make(chan error, 1)

	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		token := r.URL.Query().Get("token")
		if token == "" {
			http.Error(w, "missing token", http.StatusBadRequest)
			errCh <- fmt.Errorf("no token in callback")
			return
		}
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `<!DOCTYPE html><html><body style="font-family:sans-serif;text-align:center;padding:60px">
<h2>&#10003; Authenticated!</h2><p>You can close this tab and return to your terminal.</p></body></html>`)
		keyCh <- token
	})

	srv := &http.Server{Handler: mux}
	go func() {
		if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	output.Info("Opening browser for authentication...")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "  If the browser doesn't open, visit:")
	fmt.Fprintln(os.Stderr, "  "+output.StyleURL.Render(authURL))
	fmt.Fprintln(os.Stderr)

	openBrowser(authURL)

	fmt.Fprint(os.Stderr, "  Or paste your API key and press Enter: ")
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan() {
			text := strings.TrimSpace(scanner.Text())
			if text != "" {
				keyCh <- text
			}
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	select {
	case key := <-keyCh:
		_ = srv.Shutdown(context.Background())
		if err := config.SaveAPIKey(key); err != nil {
			return fmt.Errorf("saving key: %w", err)
		}
		fmt.Fprintln(os.Stderr)
		output.Success("Logged in successfully")
		return nil
	case err := <-errCh:
		return err
	case <-ctx.Done():
		return fmt.Errorf("login timed out after 5 minutes")
	}
}

// ── logout ────────────────────────────────────────────────────────────────────

func newLogoutCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Log out and remove stored credentials",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := config.ClearAPIKey(); err != nil {
				return err
			}
			output.Success("Logged out")
			return nil
		},
	}
}
