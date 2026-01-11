package cli

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/spf13/cobra"

	"yukti/internal/infrastructure/config"
	"yukti/internal/infrastructure/google"
	"yukti/internal/infrastructure/keychain"
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
	fmt.Println("⚡ Yukti Status")
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println()

	// Check config
	configPath := config.DefaultConfigPath()
	cfg, err := config.Load()

	fmt.Println("Configuration:")
	if err != nil {
		fmt.Printf("  ✗ Config file: not found\n")
		fmt.Printf("    Run 'yukti init' to set up credentials\n")
	} else {
		fmt.Printf("  ✓ Config file: %s\n", configPath)
		if cfg.OAuth.ClientID != "" {
			// Mask the client ID for privacy
			maskedID := maskString(cfg.OAuth.ClientID)
			fmt.Printf("  ✓ Client ID: %s\n", maskedID)
		}
		if cfg.OAuth.ClientSecret != "" {
			fmt.Printf("  ✓ Client Secret: ****\n")
		}
	}
	fmt.Println()

	// Check authentication
	fmt.Println("Authentication:")

	kc := keychain.NewStore()
	if !kc.HasToken() {
		fmt.Printf("  ✗ Not logged in\n")
		fmt.Printf("    Run 'yukti login' to authenticate\n")
		return nil
	}

	token, err := kc.LoadToken()
	if err != nil {
		fmt.Printf("  ✗ Error loading token: %v\n", err)
		return nil
	}

	if token == nil {
		fmt.Printf("  ✗ No token found\n")
		return nil
	}

	// Check token validity
	if token.Valid() {
		fmt.Printf("  ✓ Logged in\n")

		// Show token expiry
		if !token.Expiry.IsZero() {
			remaining := time.Until(token.Expiry)
			if remaining > 0 {
				fmt.Printf("  ✓ Token expires in: %s\n", formatDuration(remaining))
			} else {
				fmt.Printf("  ⚠ Token expired (will refresh on next use)\n")
			}
		}

		// Show token type
		if token.TokenType != "" {
			fmt.Printf("  ✓ Token type: %s\n", token.TokenType)
		}
	} else {
		fmt.Printf("  ⚠ Token expired\n")

		// Check if we can refresh
		if cfg != nil && cfg.OAuth.ClientID != "" {
			logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
				Level: slog.LevelError,
			}))
			auth := google.NewAuthenticator(cfg.OAuth.ClientID, cfg.OAuth.ClientSecret, kc, logger)
			ctx := context.Background()
			_, err := auth.GetToken(ctx)
			if err != nil {
				fmt.Printf("  ✗ Cannot refresh token: %v\n", err)
				fmt.Printf("    Run 'yukti login' to re-authenticate\n")
			} else {
				fmt.Printf("  ✓ Token refreshed successfully\n")
			}
		}
	}

	fmt.Println()
	return nil
}

// maskString masks the middle of a string for privacy.
func maskString(s string) string {
	if len(s) <= 12 {
		return "****"
	}
	return s[:6] + "****" + s[len(s)-6:]
}

// formatDuration formats a duration in a human-readable way.
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%d seconds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%d minutes", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		hours := int(d.Hours())
		minutes := int(d.Minutes()) % 60
		if minutes > 0 {
			return fmt.Sprintf("%dh %dm", hours, minutes)
		}
		return fmt.Sprintf("%d hours", hours)
	}
	days := int(d.Hours()) / 24
	return fmt.Sprintf("%d days", days)
}
