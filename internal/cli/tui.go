package cli

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"yukti/internal/domain/project"
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
			AuthState:   tui.AuthStateLoggedOut,
			ViewFactory: views.NewFactory(),
		}, nil)
		return
	}

	// Create keychain store
	kc := keychain.NewStore()

	// Create authenticator
	auth := google.NewAuthenticator(oauthClientID, oauthClientSecret, kc, logger)

	// Check if user is already authenticated
	ctx := context.Background()
	if auth.IsAuthenticated(ctx) {
		// User is authenticated - create API client and repository
		tokenSource, err := auth.TokenSource(ctx)
		if err != nil {
			// Token source error - show login view
			runWithViewAndOpts(views.NewLoginView(auth), tui.AppOptions{
				AuthState:   tui.AuthStateLoggedOut,
				ViewFactory: views.NewFactory(),
			}, nil)
			return
		}

		// Create API client and repository
		apiClient := google.NewClient(ctx, tokenSource, logger)
		projectRepo := google.NewProjectRepository(apiClient)

		// Show welcome view with repository available
		runWithViewAndOpts(views.NewWelcomeView(), tui.AppOptions{
			AuthState:   tui.AuthStateLoggedIn,
			ViewFactory: views.NewFactory(),
		}, projectRepo)
		return
	}

	// User needs to log in - show login view
	runWithViewAndOpts(views.NewLoginView(auth), tui.AppOptions{
		AuthState:   tui.AuthStateLoggedOut,
		ViewFactory: views.NewFactory(),
	}, nil)
}

// runWithViewAndOpts runs the TUI application with the given initial view and options.
func runWithViewAndOpts(view tui.View, opts tui.AppOptions, projectRepo project.Repository) {
	app := tui.NewApp(view, opts, projectRepo)
	p := tea.NewProgram(app, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
