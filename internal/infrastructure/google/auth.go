// Package google provides the Google Apps Script API client.
package google

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"yukti/internal/infrastructure/keychain"
)

// OAuth2 scopes required for Google Apps Script API.
var scopes = []string{
	"https://www.googleapis.com/auth/script.projects",
	"https://www.googleapis.com/auth/script.deployments",
	"https://www.googleapis.com/auth/script.metrics",
	"https://www.googleapis.com/auth/script.processes",
	// Drive API scope needed for listing Apps Script projects
	// (Apps Script API doesn't have a list endpoint)
	"https://www.googleapis.com/auth/drive.readonly",
}

// Common errors.
var (
	ErrNotAuthenticated = errors.New("not authenticated")
	ErrAuthCancelled    = errors.New("authentication cancelled")
	ErrInvalidState     = errors.New("invalid OAuth state")
)

// Authenticator handles OAuth2 authentication with PKCE.
type Authenticator struct {
	config   *oauth2.Config
	keychain keychain.Store
	logger   *slog.Logger
}

// NewAuthenticator creates a new authenticator instance.
func NewAuthenticator(clientID, clientSecret string, kc keychain.Store, logger *slog.Logger) *Authenticator {
	return &Authenticator{
		config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			Scopes:       scopes,
			Endpoint:     google.Endpoint,
			RedirectURL:  "http://localhost:0/callback", // Will be set dynamically
		},
		keychain: kc,
		logger:   logger,
	}
}

// Login initiates the OAuth2 login flow with PKCE.
// It opens a browser for user authentication and waits for the callback.
func (a *Authenticator) Login(ctx context.Context) (*oauth2.Token, error) {
	// Generate PKCE verifier and challenge
	verifier := generateVerifier()
	challenge := generateChallenge(verifier)

	// Start local callback server on a random port
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return nil, fmt.Errorf("starting callback server: %w", err)
	}
	defer func() { _ = listener.Close() }()

	port := listener.Addr().(*net.TCPAddr).Port
	a.config.RedirectURL = fmt.Sprintf("http://localhost:%d/callback", port)

	// Generate state for CSRF protection
	state := generateState()

	// Build auth URL with PKCE parameters
	authURL := a.config.AuthCodeURL(state,
		oauth2.SetAuthURLParam("code_challenge", challenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
	)

	// Channels for callback communication
	codeChan := make(chan string, 1)
	errChan := make(chan error, 1)

	// HTTP server to handle OAuth callback
	server := &http.Server{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/callback" {
				http.NotFound(w, r)
				return
			}

			// Verify state to prevent CSRF
			if r.URL.Query().Get("state") != state {
				errChan <- ErrInvalidState
				http.Error(w, "Invalid state", http.StatusBadRequest)
				return
			}

			// Check for errors from OAuth provider
			if errMsg := r.URL.Query().Get("error"); errMsg != "" {
				errChan <- fmt.Errorf("OAuth error: %s", errMsg)
				http.Error(w, errMsg, http.StatusBadRequest)
				return
			}

			// Get authorization code
			code := r.URL.Query().Get("code")
			if code == "" {
				errChan <- errors.New("no code in callback")
				http.Error(w, "Missing code", http.StatusBadRequest)
				return
			}

			// Send success response to browser
			w.Header().Set("Content-Type", "text/html")
			_, _ = w.Write([]byte(successHTML))
			codeChan <- code
		}),
	}

	// Start server in background
	go func() {
		_ = server.Serve(listener)
	}()

	// Open browser for authentication
	a.logger.Info("Opening browser for authentication",
		slog.String("url", authURL),
		slog.Int("port", port),
	)

	if err := OpenBrowser(authURL); err != nil {
		return nil, fmt.Errorf("opening browser: %w", err)
	}

	a.logger.Info("Waiting for authentication...")

	// Wait for callback, error, or timeout
	select {
	case code := <-codeChan:
		// Shut down server
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)

		// Exchange authorization code for token using PKCE verifier
		token, err := a.config.Exchange(ctx, code,
			oauth2.SetAuthURLParam("code_verifier", verifier),
		)
		if err != nil {
			return nil, fmt.Errorf("exchanging code: %w", err)
		}

		// Store token in keychain
		if err := a.keychain.StoreToken(token); err != nil {
			a.logger.Warn("Failed to store token in keychain", slog.Any("error", err))
		}

		a.logger.Info("Authentication successful")
		return token, nil

	case err := <-errChan:
		return nil, err

	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// GetToken retrieves the stored token, refreshing if necessary.
func (a *Authenticator) GetToken(ctx context.Context) (*oauth2.Token, error) {
	// Try to load from keychain
	token, err := a.keychain.LoadToken()
	if err != nil {
		return nil, fmt.Errorf("loading token: %w", err)
	}

	if token == nil {
		return nil, ErrNotAuthenticated
	}

	// If token is still valid, return it
	if token.Valid() {
		return token, nil
	}

	// Token expired, try to refresh
	a.logger.Debug("Token expired, attempting refresh")

	tokenSource := a.config.TokenSource(ctx, token)
	newToken, err := tokenSource.Token()
	if err != nil {
		// Refresh failed, user needs to re-authenticate
		return nil, fmt.Errorf("refreshing token: %w", err)
	}

	// Store refreshed token
	if err := a.keychain.StoreToken(newToken); err != nil {
		a.logger.Warn("Failed to store refreshed token", slog.Any("error", err))
	}

	return newToken, nil
}

// TokenSource returns an oauth2.TokenSource that automatically refreshes tokens.
func (a *Authenticator) TokenSource(ctx context.Context) (oauth2.TokenSource, error) {
	token, err := a.GetToken(ctx)
	if err != nil {
		return nil, err
	}
	return a.config.TokenSource(ctx, token), nil
}

// Logout removes the stored token.
func (a *Authenticator) Logout() error {
	return a.keychain.DeleteToken()
}

// IsAuthenticated checks if a valid token exists.
func (a *Authenticator) IsAuthenticated(ctx context.Context) bool {
	token, err := a.GetToken(ctx)
	return err == nil && token != nil && token.Valid()
}

// generateVerifier creates a PKCE code verifier.
func generateVerifier() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

// generateChallenge creates a PKCE code challenge from the verifier.
func generateChallenge(verifier string) string {
	h := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(h[:])
}

// generateState creates a random state value for CSRF protection.
func generateState() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

// successHTML is the HTML shown to users after successful authentication.
const successHTML = `<!DOCTYPE html>
<html>
<head>
    <title>Yukti - Authentication Successful</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
            display: flex;
            justify-content: center;
            align-items: center;
            min-height: 100vh;
            margin: 0;
            background: linear-gradient(135deg, #1A1B26 0%, #24283B 100%);
            color: #C0CAF5;
        }
        .container {
            text-align: center;
            padding: 2rem;
            background: rgba(36, 40, 59, 0.8);
            border-radius: 12px;
            box-shadow: 0 4px 24px rgba(0, 0, 0, 0.3);
        }
        .icon {
            font-size: 4rem;
            margin-bottom: 1rem;
        }
        h1 {
            color: #9ECE6A;
            margin: 0 0 0.5rem 0;
        }
        p {
            color: #9AA5CE;
            margin: 0;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="icon">⚡</div>
        <h1>Authentication Successful!</h1>
        <p>You can close this window and return to Yukti.</p>
    </div>
</body>
</html>`
