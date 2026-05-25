package components

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	appprocess "yukti/internal/application/process"
	"yukti/internal/domain/process"
	"yukti/internal/infrastructure/google"
	"yukti/internal/infrastructure/logger"
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

// ReplacePlaceholder replaces the placeholder entry with the actual result.
// Placeholder entries have ID "placeholder".
func (e *ExecutionLog) ReplacePlaceholder(entry appprocess.ExecutionEntry) {
	for i := range e.entries {
		if e.entries[i].ID == "placeholder" {
			e.entries[i] = entry
			return
		}
	}
	// If no placeholder found, add as new entry
	e.entries = append([]appprocess.ExecutionEntry{entry}, e.entries...)
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
	Up      key.Binding
	Down    key.Binding
	Expand  key.Binding
	Refresh key.Binding
}

// DefaultExecutionLogKeyMap returns the default key bindings.
func DefaultExecutionLogKeyMap() ExecutionLogKeyMap {
	return ExecutionLogKeyMap{
		Up:      key.NewBinding(key.WithKeys("k", "up"), key.WithHelp("k", "up")),
		Down:    key.NewBinding(key.WithKeys("j", "down"), key.WithHelp("j", "down")),
		Expand:  key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "toggle logs/open modal")),
		Refresh: key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "refresh logs")),
	}
}

// FetchLogsMsg is sent to request log fetching for an entry.
type FetchLogsMsg struct {
	EntryID  string
	ScriptID string
}

// LogsFetchedMsg is sent when logs have been fetched.
type LogsFetchedMsg struct {
	EntryID string
	Logs    []google.LogEntry
	Error   string
}

// OpenLogModalMsg is sent to open the full log modal.
type OpenLogModalMsg struct {
	Entry appprocess.ExecutionEntry
}

// Update handles input events.
func (e *ExecutionLog) Update(msg tea.Msg) (*ExecutionLog, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return e.handleKeyMsg(msg)
	case LogsFetchedMsg:
		e.handleLogsFetched(msg)
	case SpinnerTickMsg:
		return e.handleSpinnerTick()
	}
	return e, nil
}

// handleKeyMsg processes keyboard input.
func (e *ExecutionLog) handleKeyMsg(msg tea.KeyMsg) (*ExecutionLog, tea.Cmd) {
	if !e.focused {
		return e, nil
	}

	keys := DefaultExecutionLogKeyMap()

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
		return e.handleExpandKey()
	case key.Matches(msg, keys.Refresh):
		return e.handleRefreshKey()
	}
	return e, nil
}

// handleExpandKey handles the expand/modal key press.
func (e *ExecutionLog) handleExpandKey() (*ExecutionLog, tea.Cmd) {
	if e.selected >= len(e.entries) {
		return e, nil
	}

	entry := &e.entries[e.selected]
	if entry.LogsExpanded {
		// Already expanded - open full modal
		return e, func() tea.Msg {
			return OpenLogModalMsg{Entry: *entry}
		}
	}

	// Toggle expansion and fetch logs if needed
	entry.LogsExpanded = true
	if !entry.LogsLoaded {
		return e, func() tea.Msg {
			return FetchLogsMsg{EntryID: entry.ID, ScriptID: entry.ScriptID}
		}
	}
	return e, nil
}

// handleRefreshKey handles the refresh logs key press.
func (e *ExecutionLog) handleRefreshKey() (*ExecutionLog, tea.Cmd) {
	if e.selected >= len(e.entries) {
		return e, nil
	}

	entry := &e.entries[e.selected]
	entry.LogsLoaded = false
	entry.LogsError = ""
	return e, func() tea.Msg {
		return FetchLogsMsg{EntryID: entry.ID, ScriptID: entry.ScriptID}
	}
}

// handleLogsFetched updates an entry with fetched logs.
func (e *ExecutionLog) handleLogsFetched(msg LogsFetchedMsg) {
	for i := range e.entries {
		if e.entries[i].ID == msg.EntryID {
			e.entries[i].Logs = msg.Logs
			e.entries[i].LogsLoaded = true
			e.entries[i].LogsError = msg.Error
			return
		}
	}
}

// handleSpinnerTick advances the spinner animation.
func (e *ExecutionLog) handleSpinnerTick() (*ExecutionLog, tea.Cmd) {
	e.spinnerFrame = (e.spinnerFrame + 1) % len(spinnerFrames)
	if e.HasRunningEntry() {
		return e, tickSpinner()
	}
	return e, nil
}

// GetSelectedEntry returns the currently selected entry, or nil if none.
func (e *ExecutionLog) GetSelectedEntry() *appprocess.ExecutionEntry {
	if e.selected < len(e.entries) {
		return &e.entries[e.selected]
	}
	return nil
}

// CollapseSelected collapses the currently selected entry's logs.
func (e *ExecutionLog) CollapseSelected() {
	if e.selected < len(e.entries) {
		e.entries[e.selected].LogsExpanded = false
	}
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

	// Content lines - use plain spaces for padding (terminal bg is set via termenv)
	lines := strings.Split(content, "\n")
	var result strings.Builder
	result.WriteString(topBorder)
	result.WriteString("\033[0m\n") // Reset after top border

	verticalBorder := borderStyle.Render("│")
	contentHeight := e.height - 2
	for i := 0; i < contentHeight; i++ {
		var line string
		if i < len(lines) {
			line = lines[i]
		}

		// Use plain spaces for padding to avoid ANSI background issues
		lineWidth := lipgloss.Width(line)
		padding := ""
		if lineWidth < contentWidth {
			padding = strings.Repeat(" ", contentWidth-lineWidth)
		}

		// Add ANSI resets to ensure clean state between styled elements
		result.WriteString("\033[0m")
		result.WriteString(verticalBorder)
		result.WriteString(" ")
		result.WriteString(line)
		result.WriteString("\033[0m") // Reset after content before padding
		result.WriteString(padding)
		result.WriteString(" ")
		result.WriteString("\033[0m")
		result.WriteString(verticalBorder)
		result.WriteString("\n")
	}

	// Bottom border with log file path and scroll indicator
	mutedStyle := lipgloss.NewStyle().Foreground(styles.TextMuted)

	// Show log file name (dimmed) on the left
	logPath := logger.Path()
	logFileName := filepath.Base(logPath)
	logHint := mutedStyle.Render(" " + logFileName + " ")

	// Show scroll hint on the right when focused and scrollable
	scrollHint := ""
	if e.focused && len(e.entries) > contentHeight {
		scrollHint = mutedStyle.Render(" j/k ↕ ")
	}

	bottomWidth := max(0, e.width-2-lipgloss.Width(logHint)-lipgloss.Width(scrollHint))
	bottomBorder := borderStyle.Render("╰") +
		logHint +
		borderStyle.Render(strings.Repeat("─", bottomWidth)) +
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
// It pads the line to fill the full width with plain spaces (terminal bg is set via termenv).
func (e *ExecutionLog) getSecondLine(entry appprocess.ExecutionEntry, width int) string {
	mutedStyle := lipgloss.NewStyle().Foreground(styles.TextMuted)

	// Helper to pad line to full width with plain spaces
	padLine := func(line string) string {
		lineWidth := lipgloss.Width(line)
		if lineWidth < width {
			line += strings.Repeat(" ", width-lineWidth)
		}
		return line
	}

	switch {
	case entry.Status == process.StatusCompleted && entry.Result != nil:
		resultStyle := lipgloss.NewStyle().Foreground(styles.Success)
		result := appprocess.FormatResultCompact(entry.Result)
		line := "    " + mutedStyle.Render("└─ Returned: ") + resultStyle.Render(result)
		return "\n" + padLine(line)

	case entry.Status == process.StatusFailed && entry.Error != "":
		errStyle := lipgloss.NewStyle().Foreground(styles.Error)
		errMsg := entry.Error
		if len(errMsg) > 60 {
			errMsg = errMsg[:57] + "..."
		}
		line := "    " + mutedStyle.Render("└─ Error: ") + errStyle.Render(errMsg)
		return "\n" + padLine(line)

	case entry.Status == process.StatusRunning:
		line := "    " + mutedStyle.Render("└─ Executing function in Google Apps Script...")
		return "\n" + padLine(line)

	default:
		return ""
	}
}

// renderEntry renders a single execution entry.
func (e *ExecutionLog) renderEntry(entry appprocess.ExecutionEntry, selected bool, width int) string {
	icon, iconColor := e.getStatusIcon(entry.Status)
	iconStyle := lipgloss.NewStyle().Foreground(iconColor)

	// Expansion indicator
	expandIcon := "▸"
	if entry.LogsExpanded {
		expandIcon = "▾"
	}
	expandStyle := lipgloss.NewStyle().Foreground(styles.TextMuted)

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

	// Build line: expand indicator + icon + function name
	line := fmt.Sprintf("%s %s %s", expandStyle.Render(expandIcon), iconStyle.Render(icon), fnStyle.Render(fnName))

	// Build right-aligned status info
	statusStyle := lipgloss.NewStyle().Foreground(styles.TextSecondary)
	mutedStyle := lipgloss.NewStyle().Foreground(styles.TextMuted)

	statusPart := statusStyle.Render(getStatusText(entry.Status))
	durationPart := mutedStyle.Render(formatDuration(entry.Duration))
	timePart := mutedStyle.Render(formatTime(entry.StartTime))

	rightPart := statusPart + "    " + durationPart + "    " + timePart
	spacing := max(1, width-lipgloss.Width(line)-lipgloss.Width(rightPart)-2)
	fullLine := line + strings.Repeat(" ", spacing) + rightPart

	// Add detail line (result/error/running message), padded to full width
	secondLine := e.getSecondLine(entry, width)

	// Add inline log preview if expanded
	logLines := e.renderInlineLogs(entry, width, 10)

	// No Background() wrapping - it causes bleed. Selection is indicated by bold fnStyle.
	return fullLine + secondLine + logLines
}

// renderInlineLogs renders inline log preview (up to maxLines).
func (e *ExecutionLog) renderInlineLogs(entry appprocess.ExecutionEntry, width, maxLines int) string {
	if !entry.LogsExpanded {
		return ""
	}

	mutedStyle := lipgloss.NewStyle().Foreground(styles.TextMuted)

	// Helper to pad line to full width with plain spaces
	padLine := func(line string) string {
		lineWidth := lipgloss.Width(line)
		if lineWidth < width {
			line += strings.Repeat(" ", width-lineWidth)
		}
		return line
	}

	// If logs not loaded, show loading state
	if !entry.LogsLoaded {
		line := "    " + mutedStyle.Render("│ Loading logs...")
		return "\n" + padLine(line)
	}

	// If error fetching logs
	if entry.LogsError != "" {
		errStyle := lipgloss.NewStyle().Foreground(styles.Warning)
		line := "    " + mutedStyle.Render("│ ") + errStyle.Render(entry.LogsError)
		return "\n" + padLine(line)
	}

	// If no logs
	if len(entry.Logs) == 0 {
		line := "    " + mutedStyle.Render("│ No console output")
		return "\n" + padLine(line)
	}

	// Render up to maxLines logs
	var result strings.Builder
	logsToShow := min(maxLines, len(entry.Logs))

	for i := 0; i < logsToShow; i++ {
		log := entry.Logs[i]
		logLine := e.formatLogEntry(log, width-6)
		result.WriteString("\n")
		result.WriteString(padLine("    " + mutedStyle.Render("│ ") + logLine))
	}

	// If there are more logs, show "Enter for full logs" hint
	if len(entry.Logs) > maxLines {
		hint := fmt.Sprintf("... +%d more [Enter for full logs]", len(entry.Logs)-maxLines)
		hintLine := "    " + mutedStyle.Render("│ "+hint)
		result.WriteString("\n")
		result.WriteString(padLine(hintLine))
	}

	return result.String()
}

// formatLogEntry formats a single log entry for display.
func (e *ExecutionLog) formatLogEntry(log google.LogEntry, maxWidth int) string {
	// Get severity icon and color
	icon := google.SeverityIcon(log.Severity)
	color := lipgloss.Color(google.SeverityColor(log.Severity))

	iconStyle := lipgloss.NewStyle().Foreground(color)
	msgStyle := lipgloss.NewStyle().Foreground(styles.TextPrimary)

	// Format message - truncate if too long
	msg := log.Message
	if len(msg) > maxWidth-4 {
		msg = msg[:maxWidth-7] + "..."
	}

	return iconStyle.Render(icon) + " " + msgStyle.Render(msg)
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
