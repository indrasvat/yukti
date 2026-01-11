package cli

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"yukti/internal/infrastructure/google"
	"yukti/internal/infrastructure/keychain"
	"yukti/internal/tui"
	"yukti/internal/tui/views"
)

// runTUI starts the terminal user interface.
func runTUI() {
	// Set up logger (disabled by default for TUI)
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
		// No credentials - show welcome view with setup instructions
		runWithViewAndOpts(views.NewWelcomeView(), tui.AppOptions{
			AuthState: tui.AuthStateLoggedOut,
		})
		return
	}

	// Create keychain store
	kc := keychain.NewStore()

	// Create authenticator
	auth := google.NewAuthenticator(oauthClientID, oauthClientSecret, kc, logger)

	// Check if user is already authenticated
	ctx := context.Background()
	if auth.IsAuthenticated(ctx) {
		// User is authenticated - show welcome view (later: project list)
		runWithViewAndOpts(views.NewWelcomeView(), tui.AppOptions{
			AuthState: tui.AuthStateLoggedIn,
		})
		return
	}

	// User needs to log in - show login view
	runWithViewAndOpts(views.NewLoginView(auth), tui.AppOptions{
		AuthState: tui.AuthStateLoggedOut,
	})
}

// runWithViewAndOpts runs the TUI application with the given initial view and options.
func runWithViewAndOpts(view tui.View, opts tui.AppOptions) {
	app := tui.NewApp(view, opts)
	p := tea.NewProgram(app, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
