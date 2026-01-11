package keychain

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/oauth2"
)

// FileStore implements Store using a local file.
// This is useful for development to avoid keychain prompts.
type FileStore struct {
	path string
}

// NewFileStore creates a file-based token store at the specified path.
func NewFileStore(path string) Store {
	return &FileStore{path: path}
}

// StoreToken saves the OAuth token to a file.
func (s *FileStore) StoreToken(token *oauth2.Token) error {
	data, err := json.MarshalIndent(token, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling token: %w", err)
	}

	// Ensure directory exists
	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("creating token directory: %w", err)
	}

	// Write with restricted permissions
	if err := os.WriteFile(s.path, data, 0o600); err != nil {
		return fmt.Errorf("writing token file: %w", err)
	}

	return nil
}

// LoadToken retrieves the OAuth token from the file.
func (s *FileStore) LoadToken() (*oauth2.Token, error) {
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
func (s *FileStore) DeleteToken() error {
	err := os.Remove(s.path)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing token file: %w", err)
	}
	return nil
}

// HasToken checks if the token file exists.
func (s *FileStore) HasToken() bool {
	_, err := os.Stat(s.path)
	return err == nil
}
