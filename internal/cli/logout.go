package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"yukti/internal/infrastructure/keychain"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Log out of Google Apps Script",
	Long: `Remove stored credentials and log out of your Google account.

This will delete your authentication token from the system keychain.
You will need to run 'yukti login' again to access your projects.`,
	RunE: runLogout,
}

func init() {
	rootCmd.AddCommand(logoutCmd)
}

func runLogout(cmd *cobra.Command, args []string) error {
	// Create keychain store
	kc := keychain.NewStore()

	// Check if logged in
	if !kc.HasToken() {
		fmt.Println("Not logged in.")
		return nil
	}

	// Delete token
	if err := kc.DeleteToken(); err != nil {
		return fmt.Errorf("failed to delete credentials: %w", err)
	}

	fmt.Println("Deleted credentials.")
	return nil
}
