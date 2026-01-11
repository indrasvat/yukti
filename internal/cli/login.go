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

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Log in to Google Apps Script",
	Long: `Authenticate with Google to access your Apps Script projects.

This will open a browser window for you to sign in with your Google account
and grant Yukti access to your Apps Script projects.

Your credentials are stored securely in your system keychain.`,
	RunE: runLogin,
}

func init() {
	rootCmd.AddCommand(loginCmd)
}

func runLogin(cmd *cobra.Command, args []string) error {
	// Set up logger
	logLevel := slog.LevelError
	if verbose {
		logLevel = slog.LevelDebug
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: logLevel,
	}))

	// Get OAuth credentials
	oauthClientID, oauthClientSecret, err := getOAuthCredentials()
	if err != nil {
		return err
	}

	// Create keychain store
	kc := keychain.NewStore()

	// Create authenticator
	auth := google.NewAuthenticator(oauthClientID, oauthClientSecret, kc, logger)

	// Check if already logged in
	ctx := context.Background()
	if auth.IsAuthenticated(ctx) {
		fmt.Println("Already logged in. Use 'yukti logout' to sign out first.")
		return nil
	}

	// Start login flow
	fmt.Println("🔑 Authorize yukti by visiting the URL that opens in your browser...")
	fmt.Println()

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	token, err := auth.Login(ctx)
	if err != nil {
		return fmt.Errorf("login failed: %w", err)
	}

	// Try to get user info (email)
	email := getUserEmail(ctx, token, oauthClientID, oauthClientSecret, logger)
	if email != "" {
		fmt.Printf("✅ You are logged in as %s\n", email)
	} else {
		fmt.Println("✅ Login successful!")
	}

	return nil
}

// getOAuthCredentials returns OAuth credentials from flags, config, or defaults.
// Note: Client secret is optional for desktop apps using PKCE.
func getOAuthCredentials() (clientID, clientSecret string, err error) {
	// Priority: flags > config > defaults

	// Check flags first
	if GetClientID() != "" {
		return GetClientID(), GetClientSecret(), nil
	}

	// Try to load from config
	cfg, err := config.Load()
	if err == nil && cfg.OAuth.ClientID != "" {
		return cfg.OAuth.ClientID, cfg.OAuth.ClientSecret, nil
	}

	// Check for defaults
	if google.HasDefaultCredentials() {
		return google.DefaultClientID, google.DefaultClientSecret, nil
	}

	// No credentials available
	return "", "", fmt.Errorf(`no OAuth credentials found

To use Yukti, you need to provide a Google OAuth Client ID.

Option 1: Run 'yukti init' for guided setup

Option 2: Create a config file at ~/.config/yukti/config.json:
{
  "oauth": {
    "client_id": "YOUR_CLIENT_ID.apps.googleusercontent.com"
  }
}

Option 3: Use command-line flag:
  yukti login --client-id=YOUR_ID

To get a Client ID:
1. Go to https://console.cloud.google.com/
2. Create or select a project
3. Enable the "Apps Script API"
4. Configure OAuth consent screen
5. Go to "Credentials" → "Create Credentials" → "OAuth client ID"
6. Choose "Desktop application"
7. Copy the Client ID`)
}

// getUserEmail attempts to get the user's email from the token.
// Returns empty string if unable to get email.
func getUserEmail(_ context.Context, _ any, _, _ string, _ *slog.Logger) string {
	// The token contains the user's email in the ID token claims
	// For now, we'll return empty - we can enhance this later
	// by decoding the ID token or making a userinfo API call
	return ""
}
