// Package keychain provides secure token storage across platforms.
package keychain

import (
	"errors"

	"golang.org/x/oauth2"
)

// Common errors.
var (
	ErrTokenNotFound = errors.New("token not found")
	ErrAccessDenied  = errors.New("keychain access denied")
)

// ServiceName is the keychain service identifier for Yukti.
const ServiceName = "yukti-gas-cli"

// AccountName is the account identifier for the OAuth token.
const AccountName = "oauth-token"

// Store defines the interface for secure token storage.
type Store interface {
	// StoreToken saves the OAuth token securely.
	StoreToken(token *oauth2.Token) error

	// LoadToken retrieves the stored OAuth token.
	// Returns nil, nil if no token is stored.
	LoadToken() (*oauth2.Token, error)

	// DeleteToken removes the stored token.
	DeleteToken() error

	// HasToken checks if a token is stored.
	HasToken() bool
}
