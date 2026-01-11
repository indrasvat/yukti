//go:build windows

package keychain

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/oauth2"
)

// WindowsStore implements Store using the filesystem.
// On Windows, we use a file in the user's AppData directory.
// For production use, consider integrating with Windows Credential Manager.
type WindowsStore struct {
	path string
}

// NewStore creates a new keychain store for the current platform.
func NewStore() Store {
	appData := os.Getenv("APPDATA")
	if appData == "" {
		appData = filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Roaming")
	}

	return &WindowsStore{
		path: filepath.Join(appData, "yukti", "token.json"),
	}
}

// StoreToken saves the OAuth token to a file.
func (s *WindowsStore) StoreToken(token *oauth2.Token) error {
	// Ensure directory exists
	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	// Serialize token
	data, err := json.Marshal(token)
	if err != nil {
		return fmt.Errorf("marshaling token: %w", err)
	}

	// Write with restrictive permissions
	if err := os.WriteFile(s.path, data, 0o600); err != nil {
		return fmt.Errorf("writing token file: %w", err)
	}

	return nil
}

// LoadToken retrieves the OAuth token from the file.
func (s *WindowsStore) LoadToken() (*oauth2.Token, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading token file: %w", err)
	}

	var token oauth2.Token
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, fmt.Errorf("unmarshaling token: %w", err)
	}

	return &token, nil
}

// DeleteToken removes the token file.
func (s *WindowsStore) DeleteToken() error {
	err := os.Remove(s.path)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing token file: %w", err)
	}
	return nil
}

// HasToken checks if a token file exists.
func (s *WindowsStore) HasToken() bool {
	_, err := os.Stat(s.path)
	return err == nil
}
