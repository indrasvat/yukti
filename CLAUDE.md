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

## Automated TUI Testing With iTerm2-driver Skill

It is VERY IMPORTANT to leverage the `iterm2-driver` skill for automated testing/visulization of the Yukti TUI!

If not already installed (check `~/.claude/skills/iterm2-driver/SKILL.md` exists), it can be installed following the instructions in https://github.com/indrasvat/claude-code-skills/blob/main/README.md.

Automation scripts MUST be created under the local `./.claude/automations/` (create it if not there). They MUST have checks to ensure proper cleanup of the opened iTerm2 tab, the Yukti TUI app etc. All resources should be properly cleaned up after each automation run!

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

### Scripts Run API (scripts.run)

**Critical Requirements for Function Execution:**

The `scripts.run` API has strict requirements:

1. **Shared GCP Project**: The Apps Script project and the OAuth client (Yukti credentials) MUST be linked to the **same** GCP project. Without this, you get "resource not found: Requested entity was not found".

2. **API Executable Deployment**: The script must be deployed as "API Executable" at least once, even when using `devMode: true`.

3. **Matching OAuth Scopes**: The OAuth token must include all scopes that the script uses. For example, if the script uses GmailApp, you need `gmail.modify` scope.

**Setup Steps for Users:**
1. Link their Apps Script project to their GCP project via Apps Script → Project Settings → Change GCP project
2. Deploy as API Executable: Deploy → New deployment → API Executable
3. Re-authenticate if scope errors occur: `yukti logout && yukti login`

**Reference:** [Execute Functions with the Apps Script API](https://developers.google.com/apps-script/api/how-tos/execute)

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

### Terminal Background Color (Critical for Custom Themes)

**Problem:** When using a custom background color (e.g., Catppuccin Mocha's `#1E1E2E`), empty terminal cells show the terminal's default background instead of the app's background. This creates a "two-tone" appearance with content areas having the correct background and empty areas (trailing space, lines below content) having the terminal's darker default.

**Why lipgloss Background() doesn't work:**
- `lipgloss.Background()` only applies to **explicitly rendered characters**
- Empty terminal cells have NO characters, so they use the terminal's default background
- Padding with styled spaces doesn't reliably fill all empty cells
- `lipgloss.Place()` with `WithWhitespaceBackground()` has the same limitation

**The Solution: termenv.SetBackgroundColor()**

Use the `termenv` library (already a BubbleTea dependency) to set the terminal's default background color via OSC 11 escape sequence:

```go
import (
    tea "github.com/charmbracelet/bubbletea"
    "github.com/muesli/termenv"
)

func runTUI() {
    // Set terminal background BEFORE starting BubbleTea
    output := termenv.NewOutput(os.Stdout)
    output.SetBackgroundColor(output.Color("#1E1E2E"))  // Your app's background

    app := NewApp()
    p := tea.NewProgram(app, tea.WithAltScreen())

    _, err := p.Run()

    // Reset terminal colors AFTER TUI exits (before any os.Exit)
    output.Reset()

    if err != nil {
        os.Exit(1)
    }
}
```

**Why this works:**
- OSC 11 (`\033]11;#RRGGBB\007`) changes the terminal's **default** background
- ALL cells (including empty ones) now use your app's background
- `output.Reset()` restores original colors when app exits
- BubbleTea's alternate screen mode keeps main terminal unaffected

**Important notes:**
- Call `output.Reset()` BEFORE `os.Exit()` (defer won't run after os.Exit)
- Works with iTerm2, Terminal.app, Alacritty, Kitty, and most modern terminals
- The color string format should match lipgloss.Color (e.g., `"#1E1E2E"`)

**References:**
- [BubbleTea Issue #207](https://github.com/charmbracelet/bubbletea/issues/207)
- [termenv docs](https://pkg.go.dev/github.com/muesli/termenv)

### Custom Panel Borders with Embedded Titles

**Problem:** Building custom borders with embedded title text (like `╭─Title────╮`) causes ANSI escape code conflicts when the title is pre-styled.

**Root cause:** Each styled text segment ends with `\e[0m` (reset), which kills subsequent border styling:
```
\e[36m╭\e[0m\e[1;35mTitle\e[0m\e[36m────╮\e[0m
           ↑ This reset breaks the following border chars
```

**Solutions (in order of preference):**

1. **Use lipgloss's built-in borders** - Let lipgloss handle border rendering:
   ```go
   style := lipgloss.NewStyle().
       BorderStyle(lipgloss.RoundedBorder()).
       BorderForeground(borderColor)
   ```

2. **Separate title from border** - Render title as a row above the bordered panel

3. **Build border with plain text first** - Apply styling only at the final step:
   ```go
   border := "╭" + title + strings.Repeat("─", padding) + "╮"
   return borderStyle.Render(border)  // Style entire string at once
   ```

### Modal Overlays on Styled Background Content

**Problem:** When rendering modals that overlay on styled background content (panels with borders, colored text), the modal area causes background bleed - the rounded borders and styling in the background get "overshadowed" by plain spaces or conflicting ANSI codes.

**Root cause:** Two issues combine:
1. lipgloss's `Border()`, `Padding()`, and `Width()` don't compose well with overlay operations
2. Simple overlay implementations replace background content with plain spaces, which show the terminal background instead of preserving the styled content

**The Solution: ANSI-aware string slicing with `ansi.Cut`**

Use `github.com/charmbracelet/x/ansi` (already a transitive dependency) to extract and preserve background content:

```go
import "github.com/charmbracelet/x/ansi"

// composeModalLine overlays a modal line onto a background line.
// Background is visible on sides; modal replaces the center portion.
func composeModalLine(bgLine, modalLine string, leftOffset, modalWidth, totalWidth int) string {
    // Use ansi.Cut to extract background content while preserving ANSI codes
    leftPart := ansi.Cut(bgLine, 0, leftOffset)
    rightStart := leftOffset + modalWidth
    rightPart := ansi.Cut(bgLine, rightStart, totalWidth)

    // Compose: left bg + reset + modal + reset + right bg
    return leftPart + "\033[0m" + modalLine + "\033[0m" + rightPart
}
```

**Key insights:**
- `ansi.Cut(s, left, right)` extracts characters from position `left` to `right`, preserving ANSI escape codes
- Add `\033[0m` resets between segments to prevent style bleeding
- Don't use `Background()` on modal styles - the border provides visual separation
- For panels (like Execution Log), use manual border rendering with ANSI resets between elements

**Pattern for manual panel borders (same technique works for modals):**
```go
// Content lines - use plain spaces for padding (terminal bg is set via termenv)
verticalBorder := borderStyle.Render("│")
for i := 0; i < contentHeight; i++ {
    result.WriteString("\033[0m")        // Reset before border
    result.WriteString(verticalBorder)
    result.WriteString(" ")
    result.WriteString(line)
    result.WriteString("\033[0m")        // Reset after content
    result.WriteString(padding)          // Plain spaces
    result.WriteString(" ")
    result.WriteString("\033[0m")
    result.WriteString(verticalBorder)
    result.WriteString("\n")
}
```

### Global Key Handling with Modals

**Problem:** When a view has a modal open, global key handlers (like Back/Esc) intercept keys before the modal can handle them, causing the app to navigate away instead of closing the modal.

**Solution:** Add a `ModalHandler` interface and check before handling global keys:

```go
// In router.go - optional interface for views with modals
type ModalHandler interface {
    HasModal() bool
}

// In app.go - check before handling Back key
case key.Matches(msg, a.keys.Back):
    // Don't intercept Back if current view has a modal open
    if mh, ok := a.router.Current().(ModalHandler); ok && mh.HasModal() {
        return nil, false  // Let view handle it
    }
    if a.router.CanGoBack() {
        a.router.Pop()
        return nil, true
    }

// In view - implement the interface
func (v *WorkspaceView) HasModal() bool {
    return v.showLogPath || v.help.IsVisible() || v.fuzzy.IsVisible()
}
```

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

### Help Modal Display Bugs (Critical Lesson)

**Context:** This bug took dozens of commits to fix. TWO separate bugs caused a mutually exclusive failure mode where either the header was pushed off screen OR the modal was cut off at the bottom.

#### Bug 1: Go Switch Type Assertion Variable Shadowing

**Symptoms:**
- Views received `v.height = 70` when terminal was 70 lines
- Views SHOULD have received `v.height = 64` (70 - 6 for header/footer)
- Header got pushed off screen because views rendered to full terminal height

**Root cause:** Go's `switch msg := msg.(type)` creates a **LOCAL** variable that shadows the outer `msg`:

```go
// BROKEN CODE - DO NOT USE:
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {  // ← Creates LOCAL 'msg' that shadows outer!
    case tea.WindowSizeMsg:
        msg.Height = max(1, msg.Height-6)  // Only modifies local copy!
    }
    // Views receive the ORIGINAL msg with full height
    return a.router.Current().Update(msg)
}
```

**Fix:** Use a different variable name and reassign to outer variable:

```go
// CORRECT CODE:
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch typedMsg := msg.(type) {
    case tea.WindowSizeMsg:
        typedMsg.Height = max(1, typedMsg.Height-6)
        msg = typedMsg  // ← CRITICAL: Reassign to outer variable!
    }
    return a.router.Current().Update(msg)
}
```

**Warning signs:** If views are rendering too tall or content is pushed off screen, check if `WindowSizeMsg` modifications are being properly propagated.

**Linter protection:** The `govet` shadow analyzer catches this. Enabled in `.golangci.yml`:
```yaml
settings:
  govet:
    enable:
      - shadow  # Catches switch msg := msg.(type) bugs
```

#### Bug 2: Empty String Padding Breaks Modal Overlay

**Symptoms:**
- Modal appeared to start at column 0 on certain lines instead of being centered
- Bottom portion of modal was misaligned with top portion
- Debug screen dumps showed modal content appearing at wrong horizontal position

**Root cause:** `ensureExactHeight()` padded with empty strings `""`. When `ansi.Cut("", 0, leftOffset)` operates on an empty string, it returns an empty string for the left portion:

```go
// BROKEN CODE - DO NOT USE:
func ensureExactHeight(content string, height int) string {
    lines := strings.Split(content, "\n")
    for len(lines) < height {
        lines = append(lines, "")  // Empty strings break ansi.Cut!
    }
    return strings.Join(lines, "\n")
}

// When composing modal overlay:
leftPart := ansi.Cut("", 0, 30)  // Returns "" (empty), not 30 spaces!
// Result: modal starts at column 0 instead of column 30
```

**Fix:** Pad with full-width lines of spaces:

```go
// CORRECT CODE:
func ensureExactHeight(content string, height, width int) string {
    lines := strings.Split(content, "\n")
    emptyLine := strings.Repeat(" ", width)  // Full-width for overlay!
    for len(lines) < height {
        lines = append(lines, emptyLine)
    }
    return strings.Join(lines, "\n")
}
```

**Warning signs:** If modal overlays appear misaligned or shifted to the left on certain lines, check if the background content has proper full-width padding.

#### TUI Testing: Avoid False Positives

**Problem encountered:** Multiple times the fix was declared "complete" when it wasn't. Visual inspection via screenshots wasn't catching the issues.

**Solution: Automated screen content verification**

1. **Always dump screen contents, not just screenshots:**
   ```python
   async def dump_screen(session, label: str):
       screen = await session.async_get_screen_contents()
       print(f"\n--- {label} ---")
       for i in range(screen.number_of_lines):
           line = screen.line(i).string
           print(f"{i:03d}: {line}")
   ```

2. **Define explicit pass/fail criteria:**
   ```python
   # Check for critical elements by LINE POSITION
   found_header = False
   found_modal_top = -1
   found_modal_bottom = -1

   for i in range(screen.number_of_lines):
       line = screen.line(i).string
       if "⚡" in line and "Yukti" in line:
           found_header = True
           print(f"✓ Header found on line {i}")
       if "╭" in line and "Keybindings" in line:
           found_modal_top = i
       if "╰" in line and "──────" in line:
           found_modal_bottom = i

   # Verify modal is COMPLETE (has both top and bottom)
   if found_modal_top >= 0 and found_modal_bottom > found_modal_top:
       print(f"✓ Modal complete: top={found_modal_top}, bottom={found_modal_bottom}")
   else:
       print(f"✗ Modal INCOMPLETE or missing!")
   ```

3. **Test both states: before and after modal opens:**
   - Screenshot/dump the base view
   - Open modal with `await session.async_send_text("?")`
   - Screenshot/dump with modal
   - Verify header is STILL visible (critical for this bug)

4. **Use debug logging in the Go code:**
   ```go
   // Add temporary debug logging to understand what values views receive
   f, _ := os.OpenFile("/tmp/yukti_debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
   fmt.Fprintf(f, "renderList: vheight=%d vwidth=%d\n", v.height, v.width)
   f.Close()
   ```

## BubbleTea/LipGloss API Reference

### APIs to USE

| API | Purpose | Notes |
|-----|---------|-------|
| `lipgloss.MaxHeight(n)` | Cap content height | Use for scrollable areas; truncates excess |
| `lipgloss.Width(n)` | Set exact width | Reliable for fixed-width panels |
| `ansi.Cut(s, left, right)` | ANSI-aware substring | Essential for modal overlays |
| `ansi.StringWidth(s)` | Get display width | Accounts for ANSI codes |
| `termenv.SetBackgroundColor()` | Terminal default bg | Fills ALL cells including empty ones |
| `strings.Repeat(" ", width)` | Full-width padding | Required for overlay compositing |

### APIs to AVOID or Use Carefully

| API | Problem | Alternative |
|-----|---------|-------------|
| `lipgloss.Height(n)` | Sets MINIMUM, not exact | Use `MaxHeight` + manual padding |
| `lipgloss.Place()` | Unpredictable with modals | Manual composition with `ansi.Cut` |
| `switch msg := msg.(type)` | Shadows outer variable | Use `switch typedMsg := msg.(type)` then reassign |
| Empty string `""` for padding | Breaks `ansi.Cut` overlay | Use `strings.Repeat(" ", width)` |
| `lipgloss.Background()` | Only styled chars | Use `termenv.SetBackgroundColor()` |

### Height Management Pattern

```go
// For views that need exact height (e.g., panels with modals):
func (v *View) renderContent() string {
    content := v.buildContent()

    // 1. Cap content to available height
    style := lipgloss.NewStyle().MaxHeight(v.height)
    content = style.Render(content)

    // 2. Ensure exact height with full-width padding
    content = ensureExactHeight(content, v.height, v.width)

    return content
}

func ensureExactHeight(content string, height, width int) string {
    lines := strings.Split(content, "\n")

    // Truncate if too long
    if len(lines) > height {
        lines = lines[:height]
    }

    // Pad with full-width lines if too short
    emptyLine := strings.Repeat(" ", width)
    for len(lines) < height {
        lines = append(lines, emptyLine)
    }

    return strings.Join(lines, "\n")
}
```

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
- `github.com/charmbracelet/x/ansi` - ANSI-aware string manipulation (used for modal overlay compositing)
- `github.com/muesli/termenv` - Terminal environment detection and manipulation (used for setting terminal background color)
- `github.com/spf13/cobra` - CLI framework
- `golang.org/x/oauth2` - OAuth2 with PKCE
- `github.com/keybase/go-keychain` - macOS Keychain (darwin only)
