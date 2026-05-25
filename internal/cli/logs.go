package cli

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/spf13/cobra"

	"yukti/internal/infrastructure/logger"
)

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Show log file location and recent entries",
	Long: `Display the location of Yukti's log files and optionally
tail recent log entries.

Log files are stored in:
  macOS:   ~/Library/Application Support/yukti/logs/
  Linux:   ~/.config/yukti/logs/
  Windows: %APPDATA%\yukti\logs\

Use --tail to show recent log entries.
Use --open to open the log directory in your file manager.`,
	Run: func(cmd *cobra.Command, args []string) {
		logPath := logger.DefaultLogPath()
		fmt.Printf("Log file: %s\n", logPath)

		openDir, _ := cmd.Flags().GetBool("open")
		tail, _ := cmd.Flags().GetInt("tail")

		if openDir {
			openLogDirectory(logPath)
			return
		}

		if tail > 0 {
			tailLogFile(logPath, tail)
			return
		}

		// Check if log file exists
		if _, err := os.Stat(logPath); os.IsNotExist(err) {
			fmt.Println("\nNo log file exists yet. Run yukti to create logs.")
		} else {
			info, _ := os.Stat(logPath)
			fmt.Printf("Size: %d bytes\n", info.Size())
			fmt.Printf("\nUse 'yukti logs --tail 50' to see recent entries\n")
			fmt.Printf("Use 'yukti logs --open' to open log directory\n")
		}
	},
}

func init() {
	logsCmd.Flags().IntP("tail", "n", 0, "Show last N lines of the log file")
	logsCmd.Flags().BoolP("open", "o", false, "Open log directory in file manager")
	rootCmd.AddCommand(logsCmd)
}

// openLogDirectory opens the log directory in the system file manager.
func openLogDirectory(logPath string) {
	dir := logPath[:len(logPath)-len("/yukti-2006-01-02.log")] // Remove filename pattern
	if dir == "" {
		dir = "."
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.CommandContext(context.Background(), "open", dir) //nolint:gosec // Opening known log directory
	case "linux":
		cmd = exec.CommandContext(context.Background(), "xdg-open", dir) //nolint:gosec // Opening known log directory
	case "windows":
		cmd = exec.CommandContext(context.Background(), "explorer", dir) //nolint:gosec // Opening known log directory
	default:
		fmt.Printf("Cannot open directory on %s. Path: %s\n", runtime.GOOS, dir)
		return
	}

	if err := cmd.Start(); err != nil {
		fmt.Printf("Failed to open directory: %v\n", err)
	}
}

// tailLogFile shows the last N lines of the log file.
func tailLogFile(logPath string, lines int) {
	data, err := os.ReadFile(logPath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("No log file exists yet.")
		} else {
			fmt.Printf("Error reading log file: %v\n", err)
		}
		return
	}

	// Split into lines and get last N
	content := string(data)
	allLines := splitLines(content)

	start := 0
	if len(allLines) > lines {
		start = len(allLines) - lines
	}

	for _, line := range allLines[start:] {
		fmt.Println(line)
	}
}

// splitLines splits a string into lines.
func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}
