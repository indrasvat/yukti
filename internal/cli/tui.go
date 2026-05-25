package cli

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/muesli/termenv"

	appprocess "yukti/internal/application/process"
	"yukti/internal/domain/project"
	"yukti/internal/infrastructure/cache"
	"yukti/internal/infrastructure/config"
	"yukti/internal/infrastructure/google"
	"yukti/internal/infrastructure/keychain"
	"yukti/internal/tui"
	"yukti/internal/tui/styles"
	"yukti/internal/tui/views"
	"yukti/internal/workspace"
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
		runWithViewAndOpts(newWelcomeView(), tui.AppOptions{
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

		// Create API client and repository with caching layer
		apiClient := google.NewClient(ctx, tokenSource, logger)
		googleRepo := google.NewProjectRepository(apiClient)
		projectRepo := cache.NewCachingRepository(googleRepo)

		// Create Cloud Logging service for console.log viewing
		loggingService := google.NewCloudLoggingService(ctx, tokenSource, logger)

		// Get GCP project number (try config override first, then derive from client ID)
		var gcpProjectOverride string
		if cfg, err := config.Load(); err == nil && cfg.GCPProject != "" {
			gcpProjectOverride = cfg.GCPProject
		}
		gcpProjectNum := auth.GetGCPProjectNumber(gcpProjectOverride)

		// Create script runner and process service for function execution
		scriptRunner := google.NewScriptRunner(apiClient)
		processService := appprocess.NewService(scriptRunner, loggingService, gcpProjectNum)

		// Show welcome view with repository and process service available
		runWithViewAndOpts(newWelcomeView(), tui.AppOptions{
			AuthState:   tui.AuthStateLoggedIn,
			ViewFactory: views.NewFactoryWithService(processService),
		}, projectRepo)
		return
	}

	// User needs to log in - show login view
	runWithViewAndOpts(views.NewLoginView(auth), tui.AppOptions{
		AuthState:   tui.AuthStateLoggedOut,
		ViewFactory: views.NewFactory(),
	}, nil)
}

func newWelcomeView() *views.WelcomeView {
	root, err := workspace.FindRoot(".")
	if err != nil {
		return views.NewWelcomeView()
	}
	manifest, err := workspace.LoadManifest(root)
	if err != nil {
		return views.NewWelcomeView()
	}
	return views.NewWelcomeViewWithWorkspace(&workspace.Result{
		ScriptID:   manifest.ScriptID,
		Title:      manifest.Title,
		Dir:        root,
		RemoteHash: manifest.LastRemoteHash,
	})
}

// runWithViewAndOpts runs the TUI application with the given initial view and options.
func runWithViewAndOpts(view tui.View, opts tui.AppOptions, projectRepo project.Repository) {
	// Set terminal background color to our app's background color.
	// This ensures empty cells (not explicitly styled) use our background.
	output := termenv.NewOutput(os.Stdout)
	output.SetBackgroundColor(output.Color(string(styles.Background)))

	app := tui.NewApp(view, opts, projectRepo)
	p := tea.NewProgram(app, tea.WithAltScreen())

	_, err := p.Run()

	// Reset terminal colors after TUI exits (before potential os.Exit)
	output.Reset()

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
