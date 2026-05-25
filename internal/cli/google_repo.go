package cli

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"yukti/internal/domain/project"
	"yukti/internal/infrastructure/google"
	"yukti/internal/infrastructure/keychain"
)

func authenticatedProjectRepo(ctx context.Context) (project.Repository, error) {
	logLevel := slog.LevelError
	if verbose {
		logLevel = slog.LevelDebug
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: logLevel,
	}))

	oauthClientID, oauthClientSecret, err := getOAuthCredentials()
	if err != nil {
		return nil, err
	}

	auth := google.NewAuthenticator(oauthClientID, oauthClientSecret, keychain.NewStore(), logger)
	tokenSource, err := auth.TokenSource(ctx)
	if err != nil {
		return nil, fmt.Errorf("not authenticated: %w; run 'yukti login' first", err)
	}

	return google.NewProjectRepository(google.NewClient(ctx, tokenSource, logger)), nil
}
