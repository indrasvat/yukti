# CLAUDE.md — Yukti Development Learnings

## Overview
This document captures learnings, fixes, and patterns discovered during Yukti development.
Future AI sessions should read this file to avoid repeating mistakes.

## Git Conventions
- Always create frequent, atomic, relevant, one-liner conventional commits. Commit early, commit often.
- Never bulk-add files (`git add -A` etc). Always explicitly enumerate the files to be staged/committed.
- Use conventional commit format: `type(scope): message`
  - Types: feat, fix, docs, style, refactor, test, chore
  - Keep messages concise and descriptive

## Project Structure
```
yukti/
├── cmd/yukti/          # Application entry point
├── internal/
│   ├── buildinfo/      # Version info (injected via ldflags)
│   ├── cli/            # Cobra CLI commands (login, logout, init, status, version)
│   ├── domain/         # Domain entities and interfaces
│   ├── application/    # Use cases and services
│   ├── infrastructure/ # External services (Google APIs, keychain, config)
│   │   ├── config/     # Config file management
│   │   ├── google/     # OAuth authenticator, browser opener
│   │   └── keychain/   # Token storage (keychain + file-based)
│   └── tui/            # BubbleTea TUI components
├── pkg/                # Public packages (syntax, ascii charts)
└── plugins/            # Plugin implementations
```

## CLI Commands

Yukti uses Cobra for CLI management. Available commands:
- `yukti` - Launch TUI (default when no subcommand)
- `yukti init` - Interactive OAuth setup wizard (asks for credentials + token storage preference)
- `yukti login` - OAuth authentication flow (opens browser)
- `yukti logout` - Clear stored credentials
- `yukti status` - Show auth and config state (beautified with colors, progress bar)
- `yukti version` - Show version info

The `init` wizard prompts for:
1. Client ID and Client Secret
2. Token storage preference (file-based recommended, avoids keychain prompts)

## API Learnings

### Google OAuth Setup

**Critical: Users must create their own OAuth credentials.**
- clasp's OAuth credentials are blocked by Google for third-party use
- Attempting to use clasp's client ID results in "This app is blocked" error

**Google Cloud Console Setup:**
1. Create project at https://console.cloud.google.com/
2. Enable "Apps Script API" via search bar
3. Configure OAuth consent screen via "Google Auth Platform" (left sidebar)
   - Click "Get started" or "Configure"
   - Set app name, add email, choose "External" audience
   - Add yourself as test user
4. Create credentials via "Clients" in sidebar
   - Click "+ Create Client" → "Desktop app"
   - Download JSON to get both `client_id` and `client_secret`

**Important:** Client secret IS required even for desktop apps using PKCE.
Google returns `oauth2: "invalid_request" "client_secret is missing."` without it.

### Google Apps Script API

**Rate Limits:**
- Projects API: 5000 requests/day
- Deployments API: 1000 requests/day
- Content updates: 50/minute

**Quirks:**
- `getContent` returns files in arbitrary order
- Empty projects have one file: `appsscript.json`
- Bound scripts require `parentId` in creation request

## Code Patterns

### BubbleTea Best Practices

1. Always return `tea.Cmd` from Update, never block
2. Use channels for long-running operations
3. Handle `tea.WindowSizeMsg` early in Update
4. Propagate size changes to child components

### LipGloss Styling

1. Create styles once, not in View()
2. Use AdaptiveColor for light/dark themes
3. Calculate widths dynamically from terminal size

## Bug Fixes

### Linting Issues (Phase 1)
- Use `errors.Is()` for error comparison, not `==`
- Use modern octal literals: `0o700` not `0700`
- Pre-allocate slices when length is known: `make([]T, 0, len)`
- Avoid variable shadowing with imported package names
- Handle error return values from deferred Close() calls: `defer func() { _ = f.Close() }()`

### OAuth2 Token Refresh
- Always use `oauth2.TokenSource` wrapper, not raw token
- Store refreshed tokens back to keychain

### macOS Keychain Popup Issue

**Problem:** Repeated "yukti wants to use your confidential information" popups during development, even after clicking "Always Allow".

**Root Cause:** macOS keychain ties access permissions to the binary's code signature hash. Each rebuild produces a binary with different content (timestamps, etc.), so the hash changes and macOS sees it as a new application.

**What doesn't work:**
- Ad-hoc code signing with consistent identifier (`codesign -s - --identifier com.yukti.cli`) - the identifier is the same but the hash still changes

**Solution:** File-based token storage.

Three ways to enable (in priority order):
1. **Flag:** `yukti --token-file default status` (per-command)
2. **Config:** Add `"token_file": "default"` to config.json (persistent, recommended)
3. **Env var:** `YUKTI_TOKEN_FILE=~/.config/yukti/token.json` (per-session)

The value `default` uses the platform's config directory. Custom paths are also supported.

For development with Makefile:
- `make run`, `make dev-login`, `make dev-status` set the env var automatically

### Import Cycle Fix

**Problem:** Import cycle between `cli` and `views` packages when sharing version info.

**Solution:** Created `internal/buildinfo` package with version variables:
```go
package buildinfo
var (
    Version   = "dev"
    Commit    = "unknown"
    BuildDate = "unknown"
    GoVersion = "unknown"
)
```
Updated Makefile ldflags to use `yukti/internal/buildinfo.Version` etc.

### macOS Quarantine Attribute

**Problem:** Browser-downloaded binaries are quarantined by macOS Gatekeeper, causing "unverified developer" errors.

**Solution:** Remove the quarantine attribute before running:
```bash
xattr -d com.apple.quarantine yukti
```

Note: Downloads via `curl` or `wget` don't have this issue.

## Performance Notes

- Project list: Pagination required for >100 projects
- Code viewer: 5000+ lines causes noticeable lag
- Syntax highlighting: Cache highlighted output

## Testing Notes

- `teatest` requires explicit `tea.Quit` to finish
- iTerm2 driver automation scripts MUST have proper cleanup code (try/finally)
- Mock repositories should implement full interface
- Use `clasp` CLI as reference for API behavior verification

### Apps Script Test Scripts

When creating Apps Script test files for testing Yukti features:

1. **Always include proper logging** - Use `console.log()` to output results. Functions that only return values won't show any output in the Apps Script editor's Execution log.

2. **First run requires permissions** - Running a new script for the first time will prompt the user to grant permissions (Gmail, Drive, Calendar, etc.). This is expected behavior.

3. **Test scripts location** - Sample test scripts are documented in `docs/apps-script-ideas.md` with verified, working code examples.

4. **Script ID for testing** - `1XawIjT8_t7YrgT4uB8wmxXqnJfdHPfPwthmcoED7jc9Sr0rv7hV1Hq6D` (Yukti Test Scripts project)

## Build System

**Main Targets:**
- `make build` - Build binary (includes ad-hoc code signing on macOS)
- `make test` - Run tests
- `make lint` - Run linter
- `make fmt` - Format code
- `make ci` - Full CI pipeline (always run before committing)
- `make hooks` - Setup lefthook git hooks

**Development Targets (file-based token storage, no keychain popups):**
- `make run` - Build and run TUI
- `make dev-login` - Login using file-based token
- `make dev-status` - Show status
- `make dev-logout` - Logout

**Configuration:**
- golangci-lint v2 config in `.golangci.yml`
- lefthook pre-push hook runs `make ci`
- Binary output: `bin/yukti`
- Version info injected via ldflags to `internal/buildinfo` package

## Configuration

**Config File:** `~/.config/yukti/config.json` (macOS: `~/Library/Application Support/yukti/config.json`)
```json
{
  "oauth": {
    "client_id": "YOUR_CLIENT_ID.apps.googleusercontent.com",
    "client_secret": "YOUR_CLIENT_SECRET"
  },
  "token_file": "default"
}
```

**Config Options:**
- `oauth.client_id` - Google OAuth client ID (required)
- `oauth.client_secret` - Google OAuth client secret (required)
- `token_file` - Path to store tokens in file instead of keychain; use `"default"` for platform config dir

**Token Storage Priority:**
1. `--token-file` flag
2. `token_file` in config.json
3. `YUKTI_TOKEN_FILE` environment variable
4. System keychain (default)

**CLI Flags:**
- `--token-file <path>` - Use file-based token storage (use `default` for config dir)
- `--client-id` - Override OAuth client ID
- `--client-secret` - Override OAuth client secret
- `-v, --verbose` - Enable verbose output

## Documentation

**User-facing docs:**
- `README.md` - Installation, setup steps, commands, troubleshooting
- `docs/google-oauth-setup.md` - Google Cloud Console setup only (5 steps)

**Developer docs:**
- `CLAUDE.md` - This file, development learnings for AI sessions
- `docs/Yukti-PRD.md` - Product requirements
- `docs/Yukti-DESIGN-IMPLEMENTATION-GUIDE.md` - Architecture and design

Keep README focused on "how to use". Keep OAuth guide focused on Google Cloud Console only (no yukti commands). Avoid duplication between docs.

## Clasp Reference (Competitor)

Yukti aims to provide a better TUI experience for clasp's functionality. Here's what clasp supports:

**Auth Commands:**
- `login` - Log in to script.google.com
- `logout` - Logout
- `show-authorized-user` - Show current auth state

**Project Commands:**
- `create-script` / `create` - Create a script
- `clone-script` / `clone [scriptId] [versionNumber]` - Clone existing script
- `delete-script` / `delete [scriptId]` - Delete a project
- `list-scripts` / `list` - List App Scripts projects
- `push` - Update remote project
- `pull` - Fetch remote project
- `show-file-status` / `status` - List files to be pushed

**Version & Deployment:**
- `create-version` / `version [description]` - Create immutable version
- `list-versions` / `versions [scriptId]` - List versions
- `create-deployment` / `deploy` - Deploy a project
- `delete-deployment` / `undeploy [deploymentId]` - Delete deployment
- `list-deployments` / `deployments [scriptId]` - List deployments
- `update-deployment` / `redeploy <deploymentId>` - Update deployment

**API Management:**
- `enable-api <api>` - Enable a service
- `disable-api <api>` - Disable a service
- `list-apis` / `apis` - List enabled APIs

**Utilities:**
- `open-script [scriptId]` - Open in Apps Script IDE
- `open-container` - Open container doc
- `open-web-app [deploymentId]` - Open deployed web app
- `open-logs` - Open logs in developer console
- `setup-logs` - Setup Cloud Logging
- `tail-logs` / `logs` - Print recent log entries
- `run-function` / `run [functionName]` - Run a function
- `open-api-console` - Open API console
- `open-credentials-setup` - Open credentials page
- `start-mcp-server` / `mcp` - Start MCP server

## Dependencies

Key dependencies:
- `github.com/charmbracelet/bubbletea` - TUI framework
- `github.com/charmbracelet/lipgloss` - Styling
- `github.com/charmbracelet/bubbles` - TUI components
- `github.com/spf13/cobra` - CLI framework
- `golang.org/x/oauth2` - OAuth2 with PKCE
- `github.com/keybase/go-keychain` - macOS Keychain (darwin only)
