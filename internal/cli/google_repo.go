package cli

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"golang.org/x/oauth2"

	"yukti/internal/domain/deployment"
	"yukti/internal/domain/project"
	"yukti/internal/domain/version"
	"yukti/internal/infrastructure/google"
	"yukti/internal/infrastructure/keychain"
)

func authenticatedProjectRepo(ctx context.Context) (project.Repository, error) {
	client, err := authenticatedGoogleClient(ctx)
	if err != nil {
		return nil, err
	}
	return google.NewProjectRepository(client), nil
}

func authenticatedVersionRepo(ctx context.Context) (version.Repository, error) {
	client, err := authenticatedGoogleClient(ctx)
	if err != nil {
		return nil, err
	}
	return google.NewVersionRepository(client), nil
}

func authenticatedDeploymentRepo(ctx context.Context) (deployment.Repository, error) {
	client, err := authenticatedGoogleClient(ctx)
	if err != nil {
		return nil, err
	}
	return google.NewDeploymentRepository(client), nil
}

func authenticatedGoogleClient(ctx context.Context) (*google.Client, error) {
	tokenSource, logger, err := authenticatedTokenSource(ctx)
	if err != nil {
		return nil, err
	}
	return google.NewClient(ctx, tokenSource, logger), nil
}

func authenticatedTokenSource(ctx context.Context) (oauth2.TokenSource, *slog.Logger, error) {
	logLevel := slog.LevelError
	if verbose {
		logLevel = slog.LevelDebug
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: logLevel,
	}))

	oauthClientID, oauthClientSecret, err := getOAuthCredentials()
	if err != nil {
		return nil, nil, err
	}

	auth := google.NewAuthenticator(oauthClientID, oauthClientSecret, keychain.NewStore(), logger)
	tokenSource, err := auth.TokenSource(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("not authenticated: %w; run 'yukti login' first", err)
	}

	return tokenSource, logger, nil
}
