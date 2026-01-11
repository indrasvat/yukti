//go:build darwin

package keychain

import (
	"encoding/json"
	"errors"
	"fmt"

	gokeychain "github.com/keybase/go-keychain"
	"golang.org/x/oauth2"
)

// DarwinStore implements Store using macOS Keychain.
type DarwinStore struct{}

// NewStore creates a new keychain store for the current platform.
func NewStore() Store {
	return &DarwinStore{}
}

// StoreToken saves the OAuth token to macOS Keychain.
func (s *DarwinStore) StoreToken(token *oauth2.Token) error {
	// Serialize token to JSON
	data, err := json.Marshal(token)
	if err != nil {
		return fmt.Errorf("marshaling token: %w", err)
	}

	// Delete existing token first (keychain doesn't have upsert)
	_ = s.DeleteToken()

	// Create keychain item
	item := gokeychain.NewItem()
	item.SetSecClass(gokeychain.SecClassGenericPassword)
	item.SetService(ServiceName)
	item.SetAccount(AccountName)
	item.SetData(data)
	item.SetSynchronizable(gokeychain.SynchronizableNo)
	item.SetAccessible(gokeychain.AccessibleWhenUnlockedThisDeviceOnly)

	if err := gokeychain.AddItem(item); err != nil {
		return fmt.Errorf("adding keychain item: %w", err)
	}

	return nil
}

// LoadToken retrieves the OAuth token from macOS Keychain.
func (s *DarwinStore) LoadToken() (*oauth2.Token, error) {
	query := gokeychain.NewItem()
	query.SetSecClass(gokeychain.SecClassGenericPassword)
	query.SetService(ServiceName)
	query.SetAccount(AccountName)
	query.SetMatchLimit(gokeychain.MatchLimitOne)
	query.SetReturnData(true)

	results, err := gokeychain.QueryItem(query)
	if err != nil {
		if errors.Is(err, gokeychain.ErrorItemNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("querying keychain: %w", err)
	}

	if len(results) == 0 {
		return nil, nil
	}

	var token oauth2.Token
	if err := json.Unmarshal(results[0].Data, &token); err != nil {
		return nil, fmt.Errorf("unmarshaling token: %w", err)
	}

	return &token, nil
}

// DeleteToken removes the token from macOS Keychain.
func (s *DarwinStore) DeleteToken() error {
	item := gokeychain.NewItem()
	item.SetSecClass(gokeychain.SecClassGenericPassword)
	item.SetService(ServiceName)
	item.SetAccount(AccountName)

	err := gokeychain.DeleteItem(item)
	if err != nil && !errors.Is(err, gokeychain.ErrorItemNotFound) {
		return fmt.Errorf("deleting keychain item: %w", err)
	}

	return nil
}

// HasToken checks if a token exists in macOS Keychain.
func (s *DarwinStore) HasToken() bool {
	token, err := s.LoadToken()
	return err == nil && token != nil
}
