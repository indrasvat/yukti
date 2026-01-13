package components

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	appprocess "yukti/internal/application/process"
	"yukti/internal/domain/process"
	"yukti/internal/tui/styles"
)

// ExecutionLog displays a list of script executions with expandable details.
type ExecutionLog struct {
	entries  []appprocess.ExecutionEntry
	selected int  // Currently selected entry index
	expanded bool // Panel expanded/collapsed state
	focused  bool // Whether this component has focus
	width    int
	height   int

	// Spinner animation state for running entries
	spinnerFrame int
}

// Spinner frames for the running animation.
var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// NewExecutionLog creates a new execution log component.
func NewExecutionLog() *ExecutionLog {
	return &ExecutionLog{
		entries:  make([]appprocess.ExecutionEntry, 0),
		expanded: false,
		selected: 0,
	}
}

// SetSize sets the component dimensions.
func (e *ExecutionLog) SetSize(width, height int) {
	e.width = width
	e.height = height
}

// SetFocused sets whether the component has focus.
func (e *ExecutionLog) SetFocused(focused bool) {
	e.focused = focused
}

// SetExpanded sets the panel expanded state.
func (e *ExecutionLog) SetExpanded(expanded bool) {
	e.expanded = expanded
}

// IsExpanded returns whether the panel is expanded.
func (e *ExecutionLog) IsExpanded() bool {
	return e.expanded
}

// Toggle toggles the panel expanded state.
func (e *ExecutionLog) Toggle() {
	e.expanded = !e.expanded
}

// AddEntry adds a new entry to the log.
func (e *ExecutionLog) AddEntry(entry appprocess.ExecutionEntry) {
	e.entries = append([]appprocess.ExecutionEntry{entry}, e.entries...)
	// Auto-expand when we get a new entry
	e.expanded = true
}

// UpdateEntry updates an existing entry by ID.
func (e *ExecutionLog) UpdateEntry(entry appprocess.ExecutionEntry) {
	for i := range e.entries {
		if e.entries[i].ID == entry.ID {
			e.entries[i] = entry
			return
		}
	}
}

// HasRunningEntry returns true if there's at least one running entry.
func (e *ExecutionLog) HasRunningEntry() bool {
	for i := range e.entries {
		if e.entries[i].Status == process.StatusRunning {
			return true
		}
	}
	return false
}

// SetEntries replaces all entries.
func (e *ExecutionLog) SetEntries(entries []appprocess.ExecutionEntry) {
	e.entries = entries
	if e.selected >= len(entries) {
		e.selected = max(0, len(entries)-1)
	}
}

// ExecutionLogKeyMap defines key bindings for the execution log.
type ExecutionLogKeyMap struct {
	Up     key.Binding
	Down   key.Binding
	Expand key.Binding
}

// DefaultExecutionLogKeyMap returns the default key bindings.
func DefaultExecutionLogKeyMap() ExecutionLogKeyMap {
	return ExecutionLogKeyMap{
		Up:     key.NewBinding(key.WithKeys("k", "up"), key.WithHelp("k", "up")),
		Down:   key.NewBinding(key.WithKeys("j", "down"), key.WithHelp("j", "down")),
		Expand: key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "expand")),
	}
}

// Update handles input events.
func (e *ExecutionLog) Update(msg tea.Msg) (*ExecutionLog, tea.Cmd) {
	keys := DefaultExecutionLogKeyMap()

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if !e.focused {
			return e, nil
		}

		switch {
		case key.Matches(msg, keys.Up):
			if e.selected > 0 {
				e.selected--
			}
		case key.Matches(msg, keys.Down):
			if e.selected < len(e.entries)-1 {
				e.selected++
			}
		case key.Matches(msg, keys.Expand):
			// Toggle entry expansion (placeholder - entry detail view not yet implemented)
			_ = e.selected < len(e.entries)
		}

	case SpinnerTickMsg:
		e.spinnerFrame = (e.spinnerFrame + 1) % len(spinnerFrames)
		if e.HasRunningEntry() {
			return e, tickSpinner()
		}
	}

	return e, nil
}

// SpinnerTickMsg is used to animate the spinner.
type SpinnerTickMsg time.Time

// tickSpinner creates a command that ticks the spinner.
func tickSpinner() tea.Cmd {
	return tea.Tick(80*time.Millisecond, func(t time.Time) tea.Msg {
		return SpinnerTickMsg(t)
	})
}

// StartSpinner returns a command to start the spinner animation.
func (e *ExecutionLog) StartSpinner() tea.Cmd {
	return tickSpinner()
}

// View renders the execution log panel.
func (e *ExecutionLog) View() string {
	if !e.expanded || e.height < 4 {
		return ""
	}

	contentWidth := e.width - 4

	// Build content
	var content strings.Builder

	if len(e.entries) == 0 {
		emptyStyle := lipgloss.NewStyle().
			Foreground(styles.TextMuted).
			Italic(true)
		content.WriteString(emptyStyle.Render("No executions yet. Press Ctrl+R to run a function."))
	} else {
		for i := range e.entries {
			if i > 0 {
				content.WriteString("\n")
			}
			content.WriteString(e.renderEntry(e.entries[i], i == e.selected, contentWidth))
		}
	}

	// Build the panel with custom title
	panel := e.renderPanel(content.String(), contentWidth)
	return panel
}

// renderPanel renders the panel with bordered frame.
func (e *ExecutionLog) renderPanel(content string, contentWidth int) string {
	borderColor := styles.Border
	titleColor := styles.TextMuted
	if e.focused {
		borderColor = styles.Primary
		titleColor = styles.Primary
	}

	borderStyle := lipgloss.NewStyle().Foreground(borderColor)
	titleStyle := lipgloss.NewStyle().Foreground(titleColor).Bold(true)

	// Build title parts separately to avoid UTF-8 slicing issues
	titlePrefix := titleStyle.Render("[3]─Execution Log")

	// Add run count badge
	badgeStr := ""
	if len(e.entries) > 0 {
		badgeStyle := lipgloss.NewStyle().Foreground(styles.TextMuted)
		badgeStr = badgeStyle.Render(fmt.Sprintf(" %d runs", len(e.entries)))
	}

	// Build indicator (▼ for expanded)
	indicatorStyle := lipgloss.NewStyle().Foreground(titleColor)
	indicator := indicatorStyle.Render(" ▼ ")

	// Calculate spacing - use lipgloss.Width for accurate ANSI-aware width
	// Total = 1 (╭) + titleWidth + remainingWidth + indicatorWidth + 1 (╮) = e.width
	titleWidth := lipgloss.Width(titlePrefix) + lipgloss.Width(badgeStr)
	indicatorWidth := lipgloss.Width(indicator)
	remainingWidth := max(0, e.width-2-titleWidth-indicatorWidth)

	// Build top border - each segment styled separately
	topBorder := borderStyle.Render("╭") +
		titlePrefix +
		badgeStr +
		borderStyle.Render(strings.Repeat("─", remainingWidth)) +
		indicator +
		borderStyle.Render("╮")

	// Content lines with explicit background
	bgStyle := lipgloss.NewStyle().Background(styles.Background)
	lines := strings.Split(content, "\n")
	var result strings.Builder
	result.WriteString(topBorder + "\n")

	contentHeight := e.height - 2
	for i := 0; i < contentHeight; i++ {
		var line string
		if i < len(lines) {
			line = lines[i]
		}

		lineWidth := lipgloss.Width(line)
		if lineWidth < contentWidth {
			padding := strings.Repeat(" ", contentWidth-lineWidth)
			line += bgStyle.Render(padding)
		}

		result.WriteString(borderStyle.Render("│") + " " + line + " " + borderStyle.Render("│") + "\n")
	}

	// Bottom border with scroll indicator
	scrollHint := ""
	if e.focused && len(e.entries) > contentHeight {
		scrollHint = lipgloss.NewStyle().Foreground(styles.TextMuted).Render(" j/k ↕ ")
	}

	bottomWidth := max(0, e.width-2-lipgloss.Width(scrollHint))
	var bottomDashes strings.Builder
	for range bottomWidth {
		bottomDashes.WriteString("─")
	}
	bottomBorder := borderStyle.Render("╰") +
		bottomDashes.String() +
		scrollHint +
		borderStyle.Render("╯")

	result.WriteString(bottomBorder)

	return result.String()
}

// getStatusIcon returns the icon and color for a given status.
func (e *ExecutionLog) getStatusIcon(status process.Status) (string, lipgloss.Color) {
	switch status {
	case process.StatusRunning:
		return spinnerFrames[e.spinnerFrame], styles.Info
	case process.StatusCompleted:
		return "✓", styles.Success
	case process.StatusFailed:
		return "✗", styles.Error
	case process.StatusTimedOut:
		return "⏱", styles.Warning
	default:
		return "?", styles.TextMuted
	}
}

// getStatusText returns the display text for a given status.
func getStatusText(status process.Status) string {
	switch status {
	case process.StatusRunning:
		return "Running..."
	case process.StatusCompleted:
		return "Completed"
	case process.StatusFailed:
		return "Failed"
	case process.StatusTimedOut:
		return "Timed out"
	default:
		return string(status)
	}
}

// getSecondLine returns the detail line for an entry (result/error/running message).
func getSecondLine(entry appprocess.ExecutionEntry) string {
	mutedStyle := lipgloss.NewStyle().Foreground(styles.TextMuted)

	switch {
	case entry.Status == process.StatusCompleted && entry.Result != nil:
		resultStyle := lipgloss.NewStyle().Foreground(styles.Success)
		result := appprocess.FormatResultCompact(entry.Result)
		return "\n    " + mutedStyle.Render("└─ Returned: ") + resultStyle.Render(result)

	case entry.Status == process.StatusFailed && entry.Error != "":
		errStyle := lipgloss.NewStyle().Foreground(styles.Error)
		errMsg := entry.Error
		if len(errMsg) > 60 {
			errMsg = errMsg[:57] + "..."
		}
		return "\n    " + mutedStyle.Render("└─ Error: ") + errStyle.Render(errMsg)

	case entry.Status == process.StatusRunning:
		return "\n    " + mutedStyle.Render("└─ Executing function in Google Apps Script...")

	default:
		return ""
	}
}

// renderEntry renders a single execution entry.
func (e *ExecutionLog) renderEntry(entry appprocess.ExecutionEntry, selected bool, width int) string {
	icon, iconColor := e.getStatusIcon(entry.Status)
	iconStyle := lipgloss.NewStyle().Foreground(iconColor)

	// Function name style
	fnStyle := lipgloss.NewStyle().Foreground(styles.Primary)
	if selected && e.focused {
		fnStyle = fnStyle.Bold(true)
	}

	// Truncate function name if too long
	fnName := entry.FunctionName
	if len(fnName) > 20 {
		fnName = fnName[:17] + "..."
	}

	// Build line: icon + function name
	line := fmt.Sprintf("  %s %s", iconStyle.Render(icon), fnStyle.Render(fnName))

	// Build right-aligned status info
	statusStyle := lipgloss.NewStyle().Foreground(styles.TextSecondary)
	mutedStyle := lipgloss.NewStyle().Foreground(styles.TextMuted)

	statusPart := statusStyle.Render(getStatusText(entry.Status))
	durationPart := mutedStyle.Render(formatDuration(entry.Duration))
	timePart := mutedStyle.Render(formatTime(entry.StartTime))

	rightPart := statusPart + "    " + durationPart + "    " + timePart
	spacing := max(1, width-lipgloss.Width(line)-lipgloss.Width(rightPart)-2)
	fullLine := line + strings.Repeat(" ", spacing) + rightPart

	// Add detail line (result/error/running message)
	secondLine := getSecondLine(entry)

	if selected && e.focused {
		highlightStyle := lipgloss.NewStyle().Background(styles.Surface)
		return highlightStyle.Render(fullLine + secondLine)
	}

	return fullLine + secondLine
}

// formatDuration formats a duration for display.
func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%.0fms", float64(d.Milliseconds()))
	}
	if d < time.Minute {
		return fmt.Sprintf("%.3fs", d.Seconds())
	}
	return fmt.Sprintf("%.1fm", d.Minutes())
}

// formatTime formats a time for display.
func formatTime(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	if diff < time.Minute {
		return "Now"
	}
	if diff < time.Hour {
		return fmt.Sprintf("%dm ago", int(diff.Minutes()))
	}
	if t.Day() == now.Day() && t.Month() == now.Month() && t.Year() == now.Year() {
		return t.Format("3:04 PM")
	}
	return t.Format("Jan 2")
}

// ShortHelp returns key bindings for help.
func (e *ExecutionLog) ShortHelp() []key.Binding {
	if !e.focused {
		return nil
	}
	keys := DefaultExecutionLogKeyMap()
	return []key.Binding{keys.Up, keys.Down}
}
