// Package main is the entry point for the Yukti TUI application.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"yukti/internal/infrastructure/config"
	"yukti/internal/infrastructure/google"
	"yukti/internal/infrastructure/keychain"
	"yukti/internal/tui"
	"yukti/internal/tui/views"
)

func main() {
	// Set up logger (disabled by default for TUI)
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelError,
	}))

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		if errors.Is(err, config.ErrConfigNotFound) {
			// No config found - show welcome view
			// User needs to set up OAuth credentials
			runWithView(views.NewWelcomeView())
			return
		}
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		// Invalid config - show welcome view
		runWithView(views.NewWelcomeView())
		return
	}

	// Create keychain store
	kc := keychain.NewStore()

	// Create authenticator
	auth := google.NewAuthenticator(
		cfg.OAuth.ClientID,
		cfg.OAuth.ClientSecret,
		kc,
		logger,
	)

	// Check if user is already authenticated
	ctx := context.Background()
	if auth.IsAuthenticated(ctx) {
		// User is authenticated - show welcome view (later: project list)
		runWithView(views.NewWelcomeView())
		return
	}

	// User needs to log in
	runWithView(views.NewLoginView(auth))
}

// runWithView runs the TUI application with the given initial view.
func runWithView(view tui.View) {
	app := tui.NewApp(view)
	p := tea.NewProgram(app, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
