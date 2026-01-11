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

(To be filled as bugs are encountered and fixed)

## Performance Notes

- Project list: Pagination required for >100 projects
- Code viewer: 5000+ lines causes noticeable lag
- Syntax highlighting: Cache highlighted output

## Testing Notes

- `teatest` requires explicit `tea.Quit` to finish
- iTerm2 driver automation scripts MUST have proper cleanup code
- Mock repositories should implement full interface
