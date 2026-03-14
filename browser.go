package main

import (
	"os/exec"
	"runtime"
)

// openBrowser attempts to open the URL in the user's default browser.
// Silently ignores errors — the user can always copy-paste the URL.
func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	_ = cmd.Start()
}
