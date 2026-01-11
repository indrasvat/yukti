// Package main is the entry point for the Yukti TUI application.
package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"yukti/internal/tui"
	"yukti/internal/tui/views"
)

func main() {
	// Create the initial welcome view
	welcomeView := views.NewWelcomeView()

	// Create the application with the welcome view
	app := tui.NewApp(welcomeView)

	// Run the program
	p := tea.NewProgram(app, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
