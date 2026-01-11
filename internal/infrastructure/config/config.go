// Package config manages application configuration.
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// Common errors.
var (
	ErrConfigNotFound = errors.New("configuration not found")
	ErrInvalidConfig  = errors.New("invalid configuration")
)

// Config holds all application configuration.
type Config struct {
	// OAuth2 credentials
	OAuth OAuthConfig `json:"oauth"`
}

// OAuthConfig holds OAuth2 credentials.
type OAuthConfig struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

// DefaultConfigPath returns the default configuration file path.
func DefaultConfigPath() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = os.ExpandEnv("$HOME/.config")
	}
	return filepath.Join(configDir, "yukti", "config.json")
}

// Load reads the configuration from the default path.
func Load() (*Config, error) {
	return LoadFrom(DefaultConfigPath())
}

// LoadFrom reads the configuration from the specified path.
func LoadFrom(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrConfigNotFound
		}
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidConfig, err)
	}

	return &cfg, nil
}

// Save writes the configuration to the default path.
func (c *Config) Save() error {
	return c.SaveTo(DefaultConfigPath())
}

// SaveTo writes the configuration to the specified path.
func (c *Config) SaveTo(path string) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("writing config file: %w", err)
	}

	return nil
}

// Validate checks if the configuration is valid.
func (c *Config) Validate() error {
	if c.OAuth.ClientID == "" {
		return fmt.Errorf("%w: missing OAuth client_id", ErrInvalidConfig)
	}
	// Client secret is optional for desktop apps using PKCE
	return nil
}
