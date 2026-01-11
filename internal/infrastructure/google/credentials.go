package google

// Default OAuth2 credentials for Yukti.
// These are embedded in the binary for convenience.
// Users can override with their own credentials via config file.
//
// These credentials are from clasp (Google's official Apps Script CLI).
// They are public and safe to use for Apps Script API access.
// See: https://github.com/google/clasp
//
// To use your own credentials:
// 1. Go to https://console.cloud.google.com/
// 2. Create a new project or select existing
// 3. Enable the Apps Script API
// 4. Create OAuth 2.0 credentials (Desktop application)
// 5. Add credentials to ~/.config/yukti/config.json

const (
	// DefaultClientID - Yukti needs its own registered OAuth credentials.
	// Users must create their own via Google Cloud Console.
	// Run `yukti init` for guided setup.
	DefaultClientID = ""

	// DefaultClientSecret - see above.
	DefaultClientSecret = ""
)

// HasDefaultCredentials returns true if default credentials are available.
func HasDefaultCredentials() bool {
	return DefaultClientID != "" && DefaultClientSecret != ""
}
