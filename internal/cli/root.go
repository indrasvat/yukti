package cli

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"yukti/internal/infrastructure/config"
	"yukti/internal/infrastructure/logger"
)

var (
	// Flags
	clientID     string
	clientSecret string
	tokenFile    string
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
	// Set up token file and logger before any command runs
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		setupTokenFile()
		// Initialize logger (ignore errors - logging is optional)
		_ = logger.Init()
		logger.Info("Starting yukti: %s", cmd.Name())
	},
	// Clean up logger when done
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		_ = logger.Close()
	},
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
	rootCmd.PersistentFlags().StringVar(&tokenFile, "token-file", "", "Store tokens in file instead of keychain (use 'default' for ~/.config/yukti/token.json)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
}

// setupTokenFile configures file-based token storage if requested.
// Priority: --token-file flag > config file > YUKTI_TOKEN_FILE env var
func setupTokenFile() {
	var effectivePath string

	// 1. Check flag first (highest priority)
	if tokenFile != "" {
		if tokenFile == "default" {
			effectivePath = config.DefaultTokenFilePath()
		} else {
			effectivePath = expandPath(tokenFile)
		}
	}

	// 2. Check config file
	if effectivePath == "" {
		if cfg, err := config.Load(); err == nil && cfg.TokenFile != "" {
			if cfg.TokenFile == "default" {
				effectivePath = config.DefaultTokenFilePath()
			} else {
				effectivePath = expandPath(cfg.TokenFile)
			}
		}
	}

	// 3. If we have a path, set the env var (keychain package checks this)
	if effectivePath != "" {
		_ = os.Setenv("YUKTI_TOKEN_FILE", effectivePath)
	}
	// If none set, env var from shell (if any) is used, otherwise keychain
}

// expandPath expands ~ to home directory.
func expandPath(path string) string {
	if path != "" && path[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[1:])
	}
	return path
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
