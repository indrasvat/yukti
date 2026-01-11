package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"yukti/internal/buildinfo"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Long:  `Print the version, commit hash, and build date of yukti.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("yukti %s\n", buildinfo.Version)
		fmt.Printf("  commit: %s\n", buildinfo.Commit)
		fmt.Printf("  built:  %s\n", buildinfo.BuildDate)
		fmt.Printf("  go:     %s\n", buildinfo.GoVersion)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
