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
│   ├── domain/         # Domain entities and interfaces
│   ├── application/    # Use cases and services
│   ├── infrastructure/ # External services (Google APIs, keychain, etc.)
│   └── tui/            # BubbleTea TUI components
├── pkg/                # Public packages (syntax, ascii charts)
└── plugins/            # Plugin implementations
```

## API Learnings

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

## Performance Notes

- Project list: Pagination required for >100 projects
- Code viewer: 5000+ lines causes noticeable lag
- Syntax highlighting: Cache highlighted output

## Testing Notes

- `teatest` requires explicit `tea.Quit` to finish
- iTerm2 driver automation scripts MUST have proper cleanup code (try/finally)
- Mock repositories should implement full interface
- Use `clasp` CLI as reference for API behavior verification

## Build System

- Makefile targets: `build`, `test`, `lint`, `fmt`, `ci`, `hooks`
- Always run `make ci` before committing
- golangci-lint v2 config in `.golangci.yml`
- lefthook pre-push hook runs `make ci`
- Binary output: `bin/yukti`

## Configuration

- Config file: `~/.config/yukti/config.json`
- Required OAuth fields: `client_id`, `client_secret`
- Keychain service name: `yukti-gas-cli`
- Token account name: `oauth-token`

## Dependencies

Key dependencies added in Phase 1:
- `github.com/charmbracelet/bubbletea` - TUI framework
- `github.com/charmbracelet/lipgloss` - Styling
- `github.com/charmbracelet/bubbles` - TUI components
- `golang.org/x/oauth2` - OAuth2 with PKCE
- `github.com/keybase/go-keychain` - macOS Keychain (darwin only)
