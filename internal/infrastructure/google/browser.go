package google

import (
	"context"
	"os/exec"
	"runtime"
)

// OpenBrowser opens the specified URL in the default browser.
func OpenBrowser(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.CommandContext(context.Background(), "open", url) //nolint:gosec // Opens the OAuth URL in the system browser.
	case "linux":
		cmd = exec.CommandContext(context.Background(), "xdg-open", url) //nolint:gosec // Opens the OAuth URL in the system browser.
	case "windows":
		cmd = exec.CommandContext(context.Background(), "rundll32", "url.dll,FileProtocolHandler", url) //nolint:gosec // Opens the OAuth URL in the system browser.
	default:
		cmd = exec.CommandContext(context.Background(), "xdg-open", url) //nolint:gosec // Opens the OAuth URL in the system browser.
	}

	return cmd.Start()
}
