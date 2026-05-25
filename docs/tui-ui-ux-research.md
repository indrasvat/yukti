# TUI UI/UX Research & Best Practices

> **Purpose:** Comprehensive research on terminal user interface design patterns, gathered from leading TUI frameworks and exemplary open-source applications.
> **Last Updated:** January 11, 2026
> **Audience:** Yukti developers building an exceptional Google Apps Script management TUI

---

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [Framework Analysis](#framework-analysis)
   - [BubbleTea & LipGloss (Go)](#bubbletea--lipgloss-go)
   - [Ratatui (Rust)](#ratatui-rust)
   - [Textual (Python)](#textual-python)
3. [Exemplary TUI Applications](#exemplary-tui-applications)
4. [Layout Patterns](#layout-patterns)
5. [Navigation & Focus Management](#navigation--focus-management)
6. [Keyboard Interaction Design](#keyboard-interaction-design)
7. [Visual Design & Theming](#visual-design--theming)
8. [Loading States & Async Operations](#loading-states--async-operations)
9. [Notifications & Feedback](#notifications--feedback)
10. [Command Palette & Fuzzy Finding](#command-palette--fuzzy-finding)
11. [Recommendations for Yukti](#recommendations-for-yukti)
12. [Sources](#sources)

---

## Executive Summary

Modern TUI applications have evolved far beyond simple text interfaces. The best TUIs rival desktop applications in functionality while maintaining the efficiency of keyboard-driven workflows. Key findings:

1. **Split-pane layouts** are the standard for IDE-like experiences (lazygit, k9s)
2. **Vim-style keybindings** (hjkl) are expected by power users but should coexist with arrow keys
3. **CSS-like styling** (Textual) and **declarative layouts** (Ratatui) improve maintainability
4. **Focus management** is critical - visual cues must clearly indicate the active pane
5. **Progressive disclosure** - command palettes (Ctrl+P/Cmd+K) for discoverability
6. **Consistent theming** - Catppuccin, Dracula, Tokyo Night provide ecosystem-wide consistency

---

## Framework Analysis

### BubbleTea & LipGloss (Go)

**Architecture:** The Elm Architecture (TEA) - Model, Update, View pattern.

#### Key Strengths
- Functional, composable design
- Rich ecosystem of components (Bubbles library)
- Excellent for building complex, stateful applications

#### Layout Best Practices

```go
// Always account for borders in calculations
contentHeight := totalHeight - headerHeight - footerHeight - 2 // -2 for borders

// Use lipgloss for joining panes
view := lipgloss.JoinHorizontal(
    lipgloss.Top,
    leftPane,
    rightPane,
)

// Dynamic width calculation
leftWidth := int(float64(totalWidth) * 0.3)
rightWidth := totalWidth - leftWidth - 1 // -1 for separator
```

#### Golden Rules
1. **Always subtract borders** - Subtract 2 from height/width calculations BEFORE rendering bordered panels
2. **Never auto-wrap in bordered panels** - Truncate text explicitly
3. **Use lipgloss.Height()/Width()** for dynamic measurements, not hardcoded values
4. **Handle WindowSizeMsg early** in Update() and propagate to child components

#### Focus Management Pattern

```go
type Model struct {
    focusedPane int // 0 = left, 1 = right
    leftPane    tea.Model
    rightPane   tea.Model
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        if msg.String() == "tab" {
            m.focusedPane = (m.focusedPane + 1) % 2
            return m, nil
        }
    }

    // Route messages to focused pane only
    if m.focusedPane == 0 {
        m.leftPane, cmd = m.leftPane.Update(msg)
    } else {
        m.rightPane, cmd = m.rightPane.Update(msg)
    }
    return m, cmd
}
```

#### Recommended Libraries
- **[BubbleLayout](https://github.com/winder/bubblelayout)** - Declarative layout manager
- **[Bubbles](https://github.com/charmbracelet/bubbles)** - Ready-to-use components (viewport, list, table, spinner)
- **[bubbletea-overlay](https://pkg.go.dev/github.com/quickphosphat/bubbletea-overlay)** - Modal/overlay support

### Ratatui (Rust)

**Architecture:** Immediate-mode rendering with constraint-based layouts.

#### Layout System

Ratatui's layout system is inspired by CSS Flexbox:

```rust
// Constraint types
Constraint::Length(20)      // Fixed 20 cells
Constraint::Percentage(50)  // 50% of parent
Constraint::Ratio(1, 3)     // 1/3 of parent
Constraint::Min(10)         // At least 10, can grow
Constraint::Max(50)         // At most 50
Constraint::Fill(1)         // Fill remaining space

// Flex modes (like CSS flexbox)
Flex::Start         // Pack to start
Flex::Center        // Center items
Flex::End           // Pack to end
Flex::SpaceAround   // Equal space around items
Flex::SpaceBetween  // Equal space between items
```

#### Nested Layouts Example

```rust
let outer = Layout::vertical([
    Constraint::Length(3),      // Header
    Constraint::Min(0),         // Content (fills remaining)
    Constraint::Length(1),      // Footer
]).split(frame.area());

// Nested horizontal split for content area
let content = Layout::horizontal([
    Constraint::Percentage(30), // Sidebar
    Constraint::Percentage(70), // Main content
]).split(outer[1]);
```

#### Key Insights
- **60+ FPS** achievable with complex layouts
- **Fraction-based math** avoids floating-point rounding errors
- **Widget builder pattern** for fluent configuration
- Support for **spacing between items**: `Layout::horizontal([...]).spacing(2)`

### Textual (Python)

**Architecture:** CSS-like styling with reactive programming model.

#### CSS-Based Layout

```python
# Python code
class MyApp(App):
    CSS = """
    #sidebar {
        width: 30%;
        dock: left;
        background: $surface;
        border-right: solid $primary;
    }

    #main {
        width: 1fr;
    }

    #footer {
        dock: bottom;
        height: 3;
    }
    """
```

#### Key Features
- **Live CSS editing** with `textual run --dev` for instant iteration
- **Docking system** - widgets can dock to edges and stay fixed during scroll
- **FR units** - `1fr` divides space equally among siblings
- **Grid layout** with row-span and column-span support
- **Responsive design** with terminal size awareness

#### Lessons for Go TUIs
1. **Separation of styling from logic** improves maintainability
2. **Hot-reload capability** dramatically speeds up design iteration
3. **Docking** is valuable for sticky headers/footers/sidebars

---

## Exemplary TUI Applications

### Lazygit

**The gold standard for Git TUI applications.**

#### Panel Layout
Six interconnected panels:
1. **Status** - Repository state, current branch
2. **Files** - Modified/staged files with diff preview
3. **Branches** - Local and remote branches
4. **Commits** - Commit history with details
5. **Stash** - Stash entries
6. **Preview** - Context-sensitive preview pane

#### Design Patterns
- **Panel-based layout** with clear visual separation
- **Context-sensitive preview** - right pane shows relevant content for selected item
- **Single-key shortcuts** - `s` for stage, `c` for commit, `p` for push
- **Confirmation dialogs** for destructive operations
- **Hot-reload configuration** - changes apply without restart

#### Configuration Options
```yaml
gui:
  windowSize: 'normal'  # normal, half, full
  border: 'rounded'     # rounded, single, double, hidden
  portraitMode: 'auto'  # auto, always, never (vertical stacking)
```

### K9s (Kubernetes TUI)

#### Design Patterns
- **Real-time monitoring** with automatic refresh
- **Resource-type navigation** - switch between pods, services, deployments
- **Contextual commands** - available actions depend on selected resource
- **Search/filter** always accessible
- **Namespace switching** - quick context change

### Soft-Serve (Git Server TUI)

#### Design Patterns
- **SSH-accessible interface** - browse repos over SSH
- **Repository browsing** with file tree and preview
- **Minimal, focused interface**

---

## Visual Analysis: Screenshots from Reference TUIs

> **Note:** This section captures detailed observations from actual TUI applications (lazygit, yazi, envx) that inspire Yukti's UI/UX design.

### Lazygit - Panel Numbering & Layout

Lazygit uses a distinctive **numbered panel system** that provides both visual hierarchy and keyboard shortcuts.

#### Panel Structure (from screenshots)
```
╭[1]─Status────────────────────╮  ╭[0]─Log/Unstaged changes─────────────────────╮
│ ✓ yukti → create-yukti       │  │  commit S9ce5c1 (HEAD → create-yukti, ...)  │
╰──────────────────────────────╯  │  Author: indrasvat <koffee2kode@gmail.com>  │
╭[2]─Files - Worktrees - ...───╮  │  Date:   5 minutes ago                      │
│ ▼ internal/tui               │  │                                             │
│   ▼ components               │  │      docs: update design guide - Phase 2    │
│     ?? filetree_test.go      │  │                                             │
│     ?? fuzzy_test.go         │  ╰─────────────────────────────────────────────╯
│   ▼ views                    │
│     ?? projects_test.go      │
╰──────────────────────────────╯
╭[3]─Local branches - Remotes──╮
│   * create-yukti ✓           │
│   17h main ✓                 │
╰──────────────────────────────╯
╭[4]─Commits - Reflog──────────╮
│ S9ce5c12 in ○ docs: update...│
│ 08f240ff in ○ docs: update...│
╰──────────────────────────────╯
╭[5]─Stash────────────────────╮
│                        0 of 0│
╰──────────────────────────────╯
```

#### Key Design Patterns Observed

1. **Numbered Panel Titles**: `[1]`, `[2]`, `[3]`, etc. in panel headers
   - Numbers allow quick navigation via keyboard
   - Combined with descriptive title: `[2]─Files - Worktrees - Submodules`

2. **Dash-Based Title Decoration**: `[N]─Title─────────`
   - Number in brackets
   - En-dash (`─`) connects number to title
   - Remaining dashes fill to border edge

3. **Item Counters**: `4 of 6`, `1 of 2`, `1 of 55`, `0 of 0`
   - Always positioned at bottom-right of panel
   - Format: `current of total`
   - Even empty panels show: `0 of 0`

4. **Panel-Local Tabs**: `Files - Worktrees - Submodules`
   - Multiple views within single panel
   - Dash-separated labels in title

5. **Focus Indication**:
   - Focused panel: Bright cyan/green border
   - Unfocused panel: Muted gray border
   - Title color matches border color

6. **Context-Sensitive Status Bar**:
   ```
   Stage: <space> │ Commit: c │ Edit: e │ Stash: s │ Discard: d │ Reset: D │ Keybindings: ?
   ```
   - Changes based on focused panel
   - Format: `Action: key │ Action: key │ ...`
   - Pipe (`│`) separators between items

7. **Modal Overlays**:
   - White/light background for modals
   - Clear "Press <enter> to get started" or "Press esc to close"
   - Keybindings menu shows categorized shortcuts:
     ```
     ─── Local ───
     <c-o> Copy branch name to clipboard
     i     Show git-flow options...
     <space> Checkout
     ─── Global ───
     <c-r> Switch to a recent repo
     ```

8. **Selection Indicators**:
   - `▸` or `>` prefix for selected item
   - Highlighted background on selected row
   - `??` prefix for untracked files (git status)
   - `✓` suffix for clean branches

9. **Tree Expansion**: `▼` for expanded, `▸` for collapsed directories

### Yazi - Miller Columns File Manager

Yazi uses a **three-column Miller layout** typical of file managers.

#### Layout Structure
```
╭─────────────────╮╭─────────────────────────────╮╭─────────────────────────╮
│ corp152         ││ ▸ add-ai-neovim             ││ docs                    │
│ corp152admin    ││   apple-scripts             ││ github-copilot          │
│ globalit        ││   Applications              ││ nvim                    │
│ ▸ robinsharma   ││   argo-example-workflows    ││ scripts                 │
│ ✓ Shared        ││   automator-scripts         ││ AI-NEOVIM-PLAN.md       │
│                 ││   browser-automations       ││                         │
│                 ││   bulk-mrs-config           ││                         │
╰─────────────────╯╰─────────────────────────────╯╰─────────────────────────╯
```

#### Key Design Patterns Observed

1. **Three-Column Miller View**:
   - Left: Parent directory (narrow)
   - Center: Current directory (widest)
   - Right: Preview of selected item

2. **Item Position Counter**: `1/55` at bottom-right
   - Format: `position/total` (slash-separated, no "of")

3. **Visual File Type Indicators**:
   - Folder icons by type (documents, pictures, music, etc.)
   - Different colors for different file types:
     - Green: Shell scripts (`.sh`)
     - Yellow: JSON/config files
     - Cyan: Markdown
     - Red: PDFs

4. **Status Line at Bottom**:
   ```
   NOR  mode  add-ai-neovim                          drwxr xr x  type  1/55
   ```
   - Mode indicator (NOR = normal)
   - Current selection name
   - Permissions
   - Position counter

5. **Minimal Borders**: Just subtle lines, no heavy boxes

### envx - Environment Variable Manager

#### Layout Structure (TUI mode)
```
╭─────────────────────────────────────────────────────────────────────────────────╮
│ envx │ Environment Variable Manager                               v0.1.0        │
├─────────────────────────────────────────────────────────────────────────────────┤
│ ┌Environment Variables (4/84)──────────────────────────────────────────────────┐│
│ │▸CARGO          C:\Users\Mikko\.rustup\toolchains\stable-x86_64...  Process  ││
│ │ CARGO_HOME     C:\Users\Mikko\.cargo                               Process  ││
│ │ CARGO_MANIFEST C:\Users\Mikko\dev\envx\crates\envx                 Process  ││
│ └──────────────────────────────────────────────────────────────────────────────┘│
├─────────────────────────────────────────────────────────────────────────────────┤
│ ↑↓/jk Navigate  Enter/v View  / Search  a Add  e Edit  d Delete  r Refresh  q Quit │
│                                                                    Item 4 of 84 │
╰─────────────────────────────────────────────────────────────────────────────────╯
```

#### Key Design Patterns Observed

1. **Header Bar**: `appname │ Description` with version at right

2. **Title with Count**: `Environment Variables (4/84)`
   - Format: `Title (current/total)`
   - In parentheses, slash-separated

3. **Three-Column Data View**:
   - Column 1: Variable name (left-aligned)
   - Column 2: Value (truncated with `...`)
   - Column 3: Source tag (right-aligned, colored)

4. **Source Tags**: `Process`, `System`, `User`, `Shell`
   - Color-coded by type
   - Right-aligned in their column

5. **Status Bar Format**:
   ```
   ↑↓/jk Navigate  Enter/v View  / Search  a Add  e Edit  d Delete  r Refresh  q Quit
   ```
   - Key combo first, then action
   - Space-separated (not pipe-separated)
   - Both arrows AND vim keys shown: `↑↓/jk`

6. **Item Counter in Footer**: `Item 4 of 84`
   - Format: `Item N of M`
   - Right-aligned in status bar area

7. **CLI Table Output** (non-TUI mode):
   ```
   ┼───────────┼────────────────────────┼─────────────────────────────┼
   │  Source   │         Name           │           Value             │
   ╪═══════════╪════════════════════════╪═════════════════════════════╪
   │ Process   │ CARGO                  │ C:\Users\Mikko\.rustup\...  │
   ├───────────┼────────────────────────┼─────────────────────────────┤
   ```
   - Double-line header separator (`═`)
   - Single-line row separators (`─`)

8. **Bar Charts for Summary**:
   ```
   ▸ By Source:
     System    ████████           17 vars
     User      ████               11 vars
     Process   ████████████████   56 vars
     Shell     ██                  0 vars
   ```
   - Colored bars proportional to count
   - Right-aligned counts

### Implementation Recommendations for Yukti

Based on this visual analysis, here are concrete implementation patterns:

#### 1. Numbered Panel Titles
```go
// Format: [N]─Title───────────────────────────
func buildPanelTitle(num int, title string, width int, focused bool) string {
    prefix := fmt.Sprintf("[%d]─%s", num, title)
    remaining := width - len(prefix) - 2 // -2 for corners
    dashes := strings.Repeat("─", max(0, remaining))
    return prefix + dashes
}
```

#### 2. Item Counter Formats
Choose ONE format consistently:
- **Lazygit style**: `4 of 6` (with "of")
- **Yazi style**: `1/55` (slash, compact)
- **envx style**: `Item 4 of 84` (with "Item" prefix)

**Recommendation for Yukti**: Use `3 of 15` format (matches lazygit, more readable)

#### 3. Context-Sensitive Status Bar
```go
type StatusBar struct {
    items []StatusItem  // Left-aligned keybinding hints
    info  string        // Right-aligned: "3 of 15"
}

// Render: "j/k nav │ Enter open │ ? help                      3 of 15"
```

#### 4. Focus Border Colors
```go
var (
    FocusedBorder   = lipgloss.Color("#89B4FA")  // Bright blue/cyan
    UnfocusedBorder = lipgloss.Color("#45475A")  // Muted gray
)
```

#### 5. Panel Title Structure
```
╭[1]─Files─────────────────────────3 of 15╮
│ ...content...                           │
╰─────────────────────────────────────────╯
```
- Number in brackets at start
- Title after dash
- Info (counter) at end before corner

---

## Layout Patterns

### Master-Detail (Split Pane)

The most common pattern for data browsing applications.

```
┌─────────────────┬───────────────────────────────────────────┐
│   MASTER        │              DETAIL                       │
│   (List/Tree)   │              (Content)                    │
│                 │                                           │
│ > Item 1        │   Selected item details...                │
│   Item 2        │                                           │
│   Item 3        │                                           │
│   Item 4        │                                           │
│                 │                                           │
├─────────────────┴───────────────────────────────────────────┤
│ [shortcuts]                                                  │
└─────────────────────────────────────────────────────────────┘
```

**Implementation Guidelines:**
1. Master pane: 25-35% width (configurable)
2. Clear visual separator (border or space)
3. Highlighted selection in master
4. Detail updates instantly on selection change
5. Tab or h/l to switch focus between panes

### Triple-Column Layout

For deeper hierarchies (files in projects in workspace).

```
┌────────────┬────────────┬──────────────────────────────────┐
│  PROJECTS  │   FILES    │           CODE                   │
│            │            │                                  │
│ > Proj A   │ > main.gs  │  1 │ function doGet() {          │
│   Proj B   │   utils.gs │  2 │   return ContentService     │
│   Proj C   │   api.gs   │  3 │     .createTextOutput()     │
│            │            │  4 │ }                            │
└────────────┴────────────┴──────────────────────────────────┘
```

### Responsive Breakpoints

Adapt layout based on terminal size:

| Width      | Layout                    |
|------------|---------------------------|
| < 80 cols  | Single column, stacked    |
| 80-120     | Two-column split          |
| > 120      | Three-column or wider master |

```go
func (m Model) getLayout() LayoutMode {
    switch {
    case m.width < 80:
        return LayoutStacked
    case m.width < 120:
        return LayoutTwoColumn
    default:
        return LayoutThreeColumn
    }
}
```

---

## Navigation & Focus Management

### Focus Indication

**Critical:** Users must always know which pane is active.

```go
// Focused pane styling
var FocusedPaneStyle = lipgloss.NewStyle().
    BorderStyle(lipgloss.RoundedBorder()).
    BorderForeground(styles.Primary) // Bright color

// Unfocused pane styling
var UnfocusedPaneStyle = lipgloss.NewStyle().
    BorderStyle(lipgloss.RoundedBorder()).
    BorderForeground(styles.Border) // Muted color
```

### Model Stack Architecture

For complex navigation flows:

```go
type Router struct {
    stack []View
}

func (r *Router) Push(v View) tea.Cmd {
    r.stack = append(r.stack, v)
    return v.Init()
}

func (r *Router) Pop() {
    if len(r.stack) > 1 {
        r.stack = r.stack[:len(r.stack)-1]
    }
}

func (r *Router) Current() View {
    return r.stack[len(r.stack)-1]
}
```

### Focus Ring Pattern

Cycle through focusable elements:

```go
type FocusManager struct {
    elements []Focusable
    current  int
}

func (f *FocusManager) Next() {
    f.current = (f.current + 1) % len(f.elements)
}

func (f *FocusManager) Prev() {
    f.current = (f.current - 1 + len(f.elements)) % len(f.elements)
}
```

---

## Keyboard Interaction Design

### Vim-Style Navigation (hjkl)

Ergonomic benefits:
- **Reduces wrist strain** - no reaching for arrow keys
- **Keeps hands on home row** - faster interaction
- **Expected by power users** - standard in developer tools

```go
var NavigationKeys = key.NewBinding(
    key.WithKeys("h", "left"),
    key.WithHelp("h/←", "left"),
)

// Support both vim keys AND arrows
switch msg.String() {
case "h", "left":
    // Move left
case "j", "down":
    // Move down
case "k", "up":
    // Move up
case "l", "right":
    // Move right
}
```

### Key Binding Layers

| Context       | Key    | Action              |
|---------------|--------|---------------------|
| Global        | `q`    | Quit                |
| Global        | `?`    | Help                |
| Global        | `Esc`  | Back / Cancel       |
| Global        | `Tab`  | Switch pane         |
| Global        | `Ctrl+P` | Command palette   |
| List view     | `j/k`  | Navigate up/down    |
| List view     | `Enter`| Select item         |
| List view     | `/`    | Filter/search       |
| Code viewer   | `g`    | Go to top           |
| Code viewer   | `G`    | Go to bottom        |
| Code viewer   | `Ctrl+d` | Page down         |
| Code viewer   | `Ctrl+u` | Page up           |

### Discoverability

Always show available keys in footer:

```
┌─────────────────────────────────────────────────────────────┐
│ Enter:open │ /:filter │ Tab:pane │ ?:help │ q:quit         │
└─────────────────────────────────────────────────────────────┘
```

---

## Visual Design & Theming

### Color Theme Selection

Popular themes with broad ecosystem support:

| Theme       | Character              | Best For                |
|-------------|------------------------|-------------------------|
| Catppuccin  | Warm, pastel, calming  | Long coding sessions    |
| Dracula     | Purple/orange, bold    | High contrast           |
| Tokyo Night | Neon, cyberpunk        | Dark theme enthusiasts  |
| Nord        | Cool, arctic blues     | Minimalist preference   |

Yukti currently uses **Catppuccin Mocha** - good choice for consistency.

### Contrast Guidelines

```go
var (
    // High contrast for important elements
    Primary       = lipgloss.Color("#89B4FA") // Bright blue

    // Medium contrast for secondary info
    TextSecondary = lipgloss.Color("#A6ADC8")

    // Low contrast for decorative elements
    TextMuted     = lipgloss.Color("#6C7086")

    // Status colors - universally understood
    Success       = lipgloss.Color("#A6E3A1") // Green
    Warning       = lipgloss.Color("#F9E2AF") // Yellow
    Error         = lipgloss.Color("#F38BA8") // Red
)
```

### Typography Hierarchy

```go
// Title - bold, primary color, larger visual weight
var TitleStyle = lipgloss.NewStyle().
    Bold(true).
    Foreground(Primary)

// Subtitle - regular weight, secondary color
var SubtitleStyle = lipgloss.NewStyle().
    Foreground(TextSecondary)

// Body - default styling
var BodyStyle = lipgloss.NewStyle().
    Foreground(TextPrimary)

// Muted - for decorative/meta information
var MutedStyle = lipgloss.NewStyle().
    Foreground(TextMuted).
    Italic(true)
```

---

## Loading States & Async Operations

### When to Use What

| Wait Time        | Indicator Type         | Notes                    |
|------------------|------------------------|--------------------------|
| < 200ms          | None                   | Imperceptible delay      |
| 200ms - 1s       | Spinner                | Brief loading            |
| 1s - 4s          | Spinner + message      | Explain what's loading   |
| 4s - 10s         | Progress bar           | Show determinate progress|
| > 10s            | Progress bar + ETA     | Users need to plan       |

### Implementation Pattern

```go
type LoadingState int

const (
    LoadingStateIdle LoadingState = iota
    LoadingStateLoading
    LoadingStateSuccess
    LoadingStateError
)

type Model struct {
    state    LoadingState
    spinner  spinner.Model
    progress float64
    errMsg   string
}

func (m Model) View() string {
    switch m.state {
    case LoadingStateLoading:
        return m.spinner.View() + " Loading projects..."
    case LoadingStateError:
        return styles.ErrorBadge(m.errMsg)
    case LoadingStateSuccess:
        return m.renderContent()
    default:
        return ""
    }
}
```

### Best Practices

1. **Position spinners contextually** - near the content being loaded
2. **Explain the wait** - "Fetching project files..." not just spinning
3. **Don't block everything** - allow navigation while background loads
4. **Consider skeleton screens** - perceived as faster than spinners

---

## Notifications & Feedback

### Toast Notifications

Short-lived, non-blocking messages for:
- Success confirmations ("Project saved")
- Warnings ("API rate limit approaching")
- Errors ("Failed to connect")

```go
type ToastMsg struct {
    Message  string
    Level    ToastLevel  // info, success, warning, error
    Duration time.Duration
}

// Auto-dismiss after 3-5 seconds
func clearToastAfterDelay(d time.Duration) tea.Cmd {
    return tea.Tick(d, func(_ time.Time) tea.Msg {
        return clearToastMsg{}
    })
}
```

### Modal Dialogs

For confirmations and critical decisions:

```go
// Use modals for:
// - Destructive actions ("Delete project?")
// - Multi-step wizards (OAuth setup)
// - Important settings changes

type ConfirmModal struct {
    title    string
    message  string
    onConfirm func() tea.Cmd
    onCancel  func() tea.Cmd
}
```

### Inline Feedback

For immediate validation:

```
┌─────────────────────────────────────┐
│ Project Name: [My Project____]      │
│ ✓ Valid name                        │
└─────────────────────────────────────┘
```

---

## Command Palette & Fuzzy Finding

### Design Pattern (VSCode/Sublime style)

```
┌─────────────────────────────────────────────────────────────┐
│ > deploy                                                     │
├─────────────────────────────────────────────────────────────┤
│ > Deploy Project          Deploy to production              │
│   Deploy to Test          Deploy to test environment        │
│   View Deployments        List all deployments              │
└─────────────────────────────────────────────────────────────┘
```

### Fuzzy Matching Rules (fzf-style)

- `^term` - starts with "term"
- `term$` - ends with "term"
- `!term` - does not include "term"
- `term1 term2` - AND (both must match)
- `term1 | term2` - OR (either matches)

### Implementation Tips

```go
type CommandPalette struct {
    input    textinput.Model
    commands []Command
    filtered []Command
    selected int
}

type Command struct {
    ID          string
    Title       string
    Description string
    Shortcut    string
    Execute     func() tea.Cmd
}

func (p *CommandPalette) filter() {
    query := strings.ToLower(p.input.Value())
    p.filtered = make([]Command, 0)

    for _, cmd := range p.commands {
        // Fuzzy match on title and description
        if fuzzyMatch(cmd.Title, query) || fuzzyMatch(cmd.Description, query) {
            p.filtered = append(p.filtered, cmd)
        }
    }
}
```

---

## Recommendations for Yukti

Based on this research, here are specific recommendations for Yukti:

### 1. Split-Pane Layout Implementation

```
┌─────────────────────────────────────────────────────────────┐
│ ⚡ Yukti - Sample Scripts                    user@gmail.com │
├─────────────────┬───────────────────────────────────────────┤
│ FILES           │  Code.gs                         [1/3] ⚡ │
│ ─────────────── │ ─────────────────────────────────────────│
│ ▸ Code.gs      ◄│   1 │ function doGet(e) {                │
│   appsscript.json│   2 │   // Handle GET requests          │
│   sidebar.html  │   3 │   return ContentService            │
│                 │   4 │     .createTextOutput('Hello')     │
│ LIBRARIES       │   5 │     .setMimeType(MimeType.TEXT);   │
│ + Add library   │   6 │ }                                  │
│                 │   7 │                                    │
│ SERVICES        │   8 │ function doPost(e) {               │
│ + Add service   │   9 │   // Handle POST requests          │
│                 │  10 │   const data = JSON.parse(e.post   │
├─────────────────┴───────────────────────────────────────────┤
│ Tab:pane │ j/k:nav │ Enter:open │ d:deploy │ ?:help │ q:quit│
└─────────────────────────────────────────────────────────────┘
```

### 2. Keyboard Shortcuts

| Key          | Action                    |
|--------------|---------------------------|
| `Tab`/`h`/`l`| Switch pane focus         |
| `j`/`k`      | Navigate up/down          |
| `Enter`      | Open/select               |
| `/`          | Filter files              |
| `Ctrl+P`     | Command palette           |
| `d`          | Deploy                    |
| `p`          | Pull latest               |
| `s`          | Push changes              |
| `r`          | Refresh                   |
| `?`          | Help                      |
| `q`          | Quit                      |

### 3. Component Architecture

```go
type SplitPaneModel struct {
    // Layout
    leftPane     tea.Model  // File tree
    rightPane    tea.Model  // Code viewer
    focusedPane  int        // 0 = left, 1 = right
    splitRatio   float64    // 0.3 = 30% left

    // State
    width        int
    height       int

    // Project context
    currentProject project.Project
    currentFile    project.File
}
```

### 4. Progressive Enhancement Path

**Phase 1: Basic Split Pane**
- Two-column layout (file tree | code viewer)
- Tab to switch focus
- Basic keyboard navigation

**Phase 2: Enhanced Interaction**
- Command palette (Ctrl+P)
- Fuzzy file search
- Resizable panes

**Phase 3: Advanced Features**
- Multiple file tabs
- Side-by-side diff view
- Integrated deployment panel

### 5. Technical Considerations

1. **Border handling**: Always subtract border width from content calculations
2. **Minimum sizes**: Set floor values (e.g., min 20 cols for file tree)
3. **Responsive**: Stack vertically on narrow terminals (< 80 cols)
4. **Performance**: Lazy-load file contents, cache syntax highlighting

---

## Sources

### BubbleTea & LipGloss
- [BubbleTea GitHub](https://github.com/charmbracelet/bubbletea)
- [LipGloss GitHub](https://github.com/charmbracelet/lipgloss)
- [Tips for building Bubble Tea programs](https://leg100.github.io/en/posts/building-bubbletea-programs/)
- [Layout handling discussion](https://github.com/charmbracelet/bubbletea/discussions/307)
- [BubbleLayout - Declarative layouts](https://github.com/winder/bubblelayout)
- [Managing nested models](https://donderom.com/posts/managing-nested-models-with-bubble-tea/)
- [BubbleTea State Machine pattern](https://zackproser.com/blog/bubbletea-state-machine)
- [Multi-model tutorial](https://blog.sometimestech.com/posts/bubbletea-multimodel)

### Ratatui (Rust)
- [Ratatui Official Site](https://ratatui.rs/)
- [Ratatui GitHub](https://github.com/ratatui/ratatui)
- [Layout Documentation](https://ratatui.rs/concepts/layout/)
- [Constraint Types](https://docs.rs/ratatui/latest/ratatui/layout/enum.Constraint.html)
- [Flex Layout](https://ratatui.rs/examples/layout/flex/)
- [awesome-ratatui](https://github.com/ratatui/awesome-ratatui)

### Textual (Python)
- [Textual Official Site](https://textual.textualize.io/)
- [7 Things learned building a TUI Framework](https://www.textualize.io/blog/7-things-ive-learned-building-a-modern-tui-framework/)
- [Layout Guide](https://textual.textualize.io/guide/layout/)
- [Design a Layout How-To](https://textual.textualize.io/how-to/design-a-layout/)
- [Docking](https://textual.textualize.io/styles/dock/)
- [Real Python Textual Tutorial](https://realpython.com/python-textual/)

### Exemplary Applications
- [Lazygit](https://github.com/jesseduffield/lazygit)
- [Lazygit Architecture (DeepWiki)](https://deepwiki.com/jesseduffield/lazygit)
- [Soft-Serve](https://github.com/charmbracelet/soft-serve)
- [awesome-tuis](https://github.com/rothgar/awesome-tuis)

### Design Patterns
- [Vim keybindings everywhere](https://github.com/erikw/vim-keybindings-everywhere-the-ultimate-list)
- [fzf - Fuzzy Finder](https://github.com/junegunn/fzf)
- [Toast Notifications UX](https://blog.logrocket.com/ux-design/toast-notifications/)
- [Progress Bars vs Spinners](https://uxmovement.com/navigation/progress-bars-vs-spinners-when-to-use-which/)
- [Loading Feedback Patterns](https://www.pencilandpaper.io/articles/ux-pattern-analysis-loading-feedback)

### Theming
- [Catppuccin](https://github.com/catppuccin/catppuccin)
- [Dracula Theme](https://draculatheme.com)
- [iTerm2 Color Schemes](https://iterm2colorschemes.com/)

---

*This document should be updated as new patterns emerge and as Yukti's TUI evolves.*
