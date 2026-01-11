package cli

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	// Flags
	clientID     string
	clientSecret string
	verbose      bool
)

// rootCmd is the base command when called without subcommands.
var rootCmd = &cobra.Command{
	Use:   "yukti",
	Short: "A beautiful TUI for Google Apps Script",
	Long: `Yukti (युक्ति) is a terminal user interface for managing
Google Apps Script projects. It provides a beautiful, keyboard-driven
interface for browsing, editing, and deploying your scripts.

Run without arguments to start the TUI, or use subcommands for
specific operations.`,
	// Run TUI when no subcommand is provided
	Run: func(cmd *cobra.Command, args []string) {
		runTUI()
	},
}

// Execute runs the root command.
func Execute() {
	// Silence usage on error - we'll show it manually where needed
	rootCmd.SilenceUsage = true

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVar(&clientID, "client-id", "", "OAuth client ID (overrides config)")
	rootCmd.PersistentFlags().StringVar(&clientSecret, "client-secret", "", "OAuth client secret (overrides config)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
}

// GetClientID returns the OAuth client ID from flags or empty string.
func GetClientID() string {
	return clientID
}

// GetClientSecret returns the OAuth client secret from flags or empty string.
func GetClientSecret() string {
	return clientSecret
}

// IsVerbose returns true if verbose mode is enabled.
func IsVerbose() bool {
	return verbose
}
