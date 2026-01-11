package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"yukti/internal/infrastructure/config"
	"yukti/internal/infrastructure/google"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Set up Yukti with your Google OAuth credentials",
	Long: `Initialize Yukti by configuring your Google OAuth credentials.

This command will guide you through:
1. Creating a Google Cloud project (if needed)
2. Enabling the Apps Script API
3. Creating OAuth credentials
4. Saving them to your config file

Your credentials are stored locally at ~/.config/yukti/config.json`,
	RunE: runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	fmt.Println("⚡ Yukti Setup")
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println()

	// Check if already configured
	cfg, err := config.Load()
	if err == nil && cfg.OAuth.ClientID != "" {
		fmt.Println("✓ Yukti is already configured!")
		fmt.Printf("  Config file: %s\n", config.DefaultConfigPath())
		fmt.Println()
		fmt.Print("Do you want to reconfigure? [y/N]: ")

		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))

		if response != "y" && response != "yes" {
			fmt.Println("Setup cancelled.")
			return nil
		}
		fmt.Println()
	}

	// Instructions
	fmt.Println("To use Yukti, you need Google OAuth credentials.")
	fmt.Println("Follow these steps to create them:")
	fmt.Println()
	fmt.Println("┌─────────────────────────────────────────────────────────────┐")
	fmt.Println("│ Step 1: Open Google Cloud Console                          │")
	fmt.Println("│         https://console.cloud.google.com/                  │")
	fmt.Println("│                                                             │")
	fmt.Println("│ Step 2: Create a new project (or select existing)          │")
	fmt.Println("│         - Click the project dropdown at the top            │")
	fmt.Println("│         - Click 'New Project'                              │")
	fmt.Println("│         - Name it 'Yukti' or similar                       │")
	fmt.Println("│                                                             │")
	fmt.Println("│ Step 3: Enable the Apps Script API                         │")
	fmt.Println("│         - Go to 'APIs & Services' → 'Library'              │")
	fmt.Println("│         - Search for 'Apps Script API'                     │")
	fmt.Println("│         - Click 'Enable'                                   │")
	fmt.Println("│                                                             │")
	fmt.Println("│ Step 4: Configure OAuth consent screen                     │")
	fmt.Println("│         - Go to 'APIs & Services' → 'OAuth consent screen' │")
	fmt.Println("│         - Choose 'External' (unless you have Workspace)    │")
	fmt.Println("│         - Fill in app name: 'Yukti'                        │")
	fmt.Println("│         - Add your email as support email                  │")
	fmt.Println("│         - Add scopes: .../auth/script.projects (and others)│")
	fmt.Println("│         - Add yourself as a test user                      │")
	fmt.Println("│                                                             │")
	fmt.Println("│ Step 5: Create OAuth credentials                           │")
	fmt.Println("│         - Go to 'APIs & Services' → 'Credentials'          │")
	fmt.Println("│         - Click 'Create Credentials' → 'OAuth client ID'   │")
	fmt.Println("│         - Choose 'Desktop application'                     │")
	fmt.Println("│         - Name it 'Yukti CLI'                              │")
	fmt.Println("│         - Copy the Client ID and Client Secret             │")
	fmt.Println("└─────────────────────────────────────────────────────────────┘")
	fmt.Println()

	// Offer to open browser
	fmt.Print("Press Enter to open Google Cloud Console (or 's' to skip): ")
	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))

	if response != "s" && response != "skip" {
		_ = google.OpenBrowser("https://console.cloud.google.com/apis/credentials")
		fmt.Println()
		fmt.Println("Browser opened. Complete the steps above, then return here.")
		fmt.Println()
	}

	// Collect credentials
	fmt.Println("Enter your OAuth credentials:")
	fmt.Println()

	fmt.Print("Client ID: ")
	clientID, _ := reader.ReadString('\n')
	clientID = strings.TrimSpace(clientID)

	if clientID == "" {
		return fmt.Errorf("client ID is required")
	}

	fmt.Print("Client Secret: ")
	clientSecret, _ := reader.ReadString('\n')
	clientSecret = strings.TrimSpace(clientSecret)

	if clientSecret == "" {
		return fmt.Errorf("client secret is required")
	}

	// Save config
	newCfg := &config.Config{
		OAuth: config.OAuthConfig{
			ClientID:     clientID,
			ClientSecret: clientSecret,
		},
	}

	if err := newCfg.Save(); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	fmt.Println()
	fmt.Println("✓ Configuration saved!")
	fmt.Printf("  Config file: %s\n", config.DefaultConfigPath())
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  1. Run 'yukti login' to authenticate with Google")
	fmt.Println("  2. Run 'yukti' to start the TUI")
	fmt.Println()

	return nil
}
