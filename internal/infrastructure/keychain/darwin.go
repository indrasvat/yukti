//go:build darwin

package keychain

import (
	"os"
	"path/filepath"
)

// NewStore creates a token store for macOS.
//
// Yukti intentionally defaults to the same restricted file store used by the
// other platforms. Rebuilt development binaries otherwise trigger repeated
// macOS Keychain prompts because Keychain access is tied to the binary hash.
func NewStore() Store {
	if tokenFile := os.Getenv("YUKTI_TOKEN_FILE"); tokenFile != "" {
		return NewFileStore(expandTokenPath(tokenFile))
	}

	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = filepath.Join(os.Getenv("HOME"), ".config")
	}
	return NewFileStore(filepath.Join(configDir, "yukti", "token.json"))
}

func expandTokenPath(path string) string {
	if path == "" || path[0] != '~' {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	return filepath.Join(home, path[1:])
}
