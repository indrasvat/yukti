package cli

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"yukti/internal/buildinfo"
	"yukti/internal/infrastructure/config"
	"yukti/internal/infrastructure/google"
	"yukti/internal/infrastructure/keychain"
)

// ANSI color codes.
const (
	colorReset  = "\033[0m"
	colorBold   = "\033[1m"
	colorDim    = "\033[2m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorCyan   = "\033[36m"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current authentication status",
	Long: `Display the current authentication and configuration status.

Shows whether you are logged in, your config file location,
and token expiration information.`,
	RunE: runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) error {
	// Print header
	printHeader()

	// Check config
	configPath := config.DefaultConfigPath()
	cfg, err := config.Load()

	printSectionHeader("Configuration")

	if err != nil {
		printStatusRow("Config", "Not found", statusError)
		printHint("Run 'yukti init' to set up credentials")
	} else {
		printStatusRow("Config", shortenPath(configPath), statusOK)
		if cfg.OAuth.ClientID != "" {
			printStatusRow("Client ID", maskString(cfg.OAuth.ClientID), statusOK)
		}
		if cfg.OAuth.ClientSecret != "" {
			printStatusRow("Secret", "Configured", statusOK)
		}
	}

	fmt.Println()
	printSectionHeader("Authentication")

	kc := keychain.NewStore()
	if !kc.HasToken() {
		printStatusRow("Status", "Not logged in", statusError)
		printHint("Run 'yukti login' to authenticate")
		fmt.Println()
		return nil
	}

	token, err := kc.LoadToken()
	if err != nil {
		printStatusRow("Status", "Error loading token", statusError)
		fmt.Println()
		return nil
	}

	if token == nil {
		printStatusRow("Status", "No token found", statusError)
		fmt.Println()
		return nil
	}

	// Check token validity
	if token.Valid() {
		printStatusRow("Status", "Logged in", statusOK)

		// Show token expiry with progress bar
		if !token.Expiry.IsZero() {
			remaining := time.Until(token.Expiry)
			if remaining > 0 {
				printTokenExpiry(remaining)
			} else {
				printStatusRow("Token", "Expired (will refresh)", statusWarning)
			}
		}
	} else {
		printStatusRow("Status", "Token expired", statusWarning)

		// Check if we can refresh
		if cfg != nil && cfg.OAuth.ClientID != "" {
			logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
				Level: slog.LevelError,
			}))
			auth := google.NewAuthenticator(cfg.OAuth.ClientID, cfg.OAuth.ClientSecret, kc, logger)
			ctx := context.Background()
			_, err := auth.GetToken(ctx)
			if err != nil {
				printStatusRow("Refresh", "Failed", statusError)
				printHint("Run 'yukti login' to re-authenticate")
			} else {
				printStatusRow("Refresh", "Success", statusOK)
			}
		}
	}

	fmt.Println()
	return nil
}

type statusType int

const (
	statusOK statusType = iota
	statusWarning
	statusError
)

func printHeader() {
	fmt.Println()
	fmt.Printf("%s%s", colorBold, colorCyan)
	fmt.Println("  ╭─────────────────────────────────────────╮")
	fmt.Println("  │                                         │")
	fmt.Printf("  │   %s⚡ YUKTI%s%s%s                              │\n", colorYellow, colorReset, colorBold, colorCyan)
	fmt.Printf("  │   %s%sGoogle Apps Script Manager%s%s%s            │\n", colorReset, colorDim, colorReset, colorBold, colorCyan)
	fmt.Println("  │                                         │")
	fmt.Println("  ╰─────────────────────────────────────────╯")
	fmt.Printf("%s", colorReset)

	// Version info
	fmt.Printf("  %s%s%s%s\n", colorDim, "v", buildinfo.Version, colorReset)
	fmt.Println()
}

func printSectionHeader(title string) {
	fmt.Printf("  %s%s━━ %s ━━%s\n", colorBold, colorBlue, title, colorReset)
	fmt.Println()
}

func printStatusRow(label, value string, status statusType) {
	var icon, valueColor string

	switch status {
	case statusOK:
		icon = colorGreen + "●" + colorReset
		valueColor = colorGreen
	case statusWarning:
		icon = colorYellow + "●" + colorReset
		valueColor = colorYellow
	case statusError:
		icon = colorRed + "●" + colorReset
		valueColor = colorRed
	}

	// Pad label to align values
	paddedLabel := fmt.Sprintf("%-12s", label)
	fmt.Printf("  %s  %s%s%s %s%s%s\n", icon, colorDim, paddedLabel, colorReset, valueColor, value, colorReset)
}

func printHint(hint string) {
	fmt.Printf("     %s%s↳ %s%s\n", colorDim, colorCyan, hint, colorReset)
}

func printTokenExpiry(remaining time.Duration) {
	// Calculate percentage (assuming 1 hour token lifetime)
	maxDuration := time.Hour
	percentage := min(1.0, float64(remaining)/float64(maxDuration))

	// Create progress bar
	barWidth := 20
	filledWidth := max(0, min(barWidth, int(percentage*float64(barWidth))))

	// Choose color based on time remaining
	var barColor string
	switch {
	case remaining < 10*time.Minute:
		barColor = colorRed
	case remaining < 30*time.Minute:
		barColor = colorYellow
	default:
		barColor = colorGreen
	}

	filled := strings.Repeat("█", filledWidth)
	empty := strings.Repeat("░", barWidth-filledWidth)

	fmt.Printf("  %s●%s  %sExpires in   %s %s%s%s%s%s %s\n",
		colorGreen, colorReset,
		colorDim, colorReset,
		barColor, filled, colorDim, empty, colorReset,
		formatDuration(remaining))
}

// maskString masks the middle of a string for privacy.
func maskString(s string) string {
	if len(s) <= 12 {
		return "••••"
	}
	return s[:8] + "••••" + s[len(s)-4:]
}

// shortenPath shortens a path for display.
func shortenPath(path string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	if strings.HasPrefix(path, home) {
		return "~" + path[len(home):]
	}
	return path
}

// formatDuration formats a duration in a human-readable way.
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	if minutes > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	return fmt.Sprintf("%dh", hours)
}
