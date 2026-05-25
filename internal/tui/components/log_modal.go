package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	appprocess "yukti/internal/application/process"
	"yukti/internal/infrastructure/google"
	"yukti/internal/tui/styles"
)

// LogModal displays console logs in a full-screen modal.
type LogModal struct {
	entry       appprocess.ExecutionEntry
	width       int
	height      int
	scrollPos   int
	searchInput textinput.Model
	searching   bool
	searchQuery string
	searchHits  []int // Line indices matching search
	searchIdx   int   // Current search hit index
}

// LogModalKeyMap defines key bindings for the log modal.
type LogModalKeyMap struct {
	Up            key.Binding
	Down          key.Binding
	Top           key.Binding
	Bottom        key.Binding
	Search        key.Binding
	NextMatch     key.Binding
	PrevMatch     key.Binding
	Close         key.Binding
	CloseSearch   key.Binding
	ConfirmSearch key.Binding
}

// DefaultLogModalKeyMap returns the default key bindings.
func DefaultLogModalKeyMap() LogModalKeyMap {
	return LogModalKeyMap{
		Up:            key.NewBinding(key.WithKeys("k", "up"), key.WithHelp("k", "up")),
		Down:          key.NewBinding(key.WithKeys("j", "down"), key.WithHelp("j", "down")),
		Top:           key.NewBinding(key.WithKeys("g"), key.WithHelp("g", "top")),
		Bottom:        key.NewBinding(key.WithKeys("G"), key.WithHelp("G", "bottom")),
		Search:        key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "search")),
		NextMatch:     key.NewBinding(key.WithKeys("n"), key.WithHelp("n", "next match")),
		PrevMatch:     key.NewBinding(key.WithKeys("N"), key.WithHelp("N", "prev match")),
		Close:         key.NewBinding(key.WithKeys("esc", "q"), key.WithHelp("esc", "close")),
		CloseSearch:   key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "close search")),
		ConfirmSearch: key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "confirm search")),
	}
}

// CloseLogModalMsg is sent when the modal should be closed.
type CloseLogModalMsg struct{}

// NewLogModal creates a new log modal.
func NewLogModal(entry appprocess.ExecutionEntry) *LogModal {
	ti := textinput.New()
	ti.Placeholder = "Search..."
	ti.CharLimit = 100
	ti.Width = 30

	return &LogModal{
		entry:       entry,
		searchInput: ti,
	}
}

// SetSize sets the modal dimensions.
func (m *LogModal) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// Update handles input events.
func (m *LogModal) Update(msg tea.Msg) (*LogModal, tea.Cmd) {
	if m.searching {
		return m.updateSearch(msg)
	}
	return m.updateNormal(msg)
}

// updateNormal handles input when not in search mode.
func (m *LogModal) updateNormal(msg tea.Msg) (*LogModal, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	keys := DefaultLogModalKeyMap()

	switch {
	case key.Matches(keyMsg, keys.Close):
		return m, func() tea.Msg { return CloseLogModalMsg{} }
	case key.Matches(keyMsg, keys.Up):
		if m.scrollPos > 0 {
			m.scrollPos--
		}
	case key.Matches(keyMsg, keys.Down):
		maxScroll := m.maxScroll()
		if m.scrollPos < maxScroll {
			m.scrollPos++
		}
	case key.Matches(keyMsg, keys.Top):
		m.scrollPos = 0
	case key.Matches(keyMsg, keys.Bottom):
		m.scrollPos = m.maxScroll()
	case key.Matches(keyMsg, keys.Search):
		m.searching = true
		m.searchInput.Focus()
		return m, textinput.Blink
	case key.Matches(keyMsg, keys.NextMatch):
		m.nextSearchHit()
	case key.Matches(keyMsg, keys.PrevMatch):
		m.prevSearchHit()
	}

	return m, nil
}

// updateSearch handles input in search mode.
func (m *LogModal) updateSearch(msg tea.Msg) (*LogModal, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		keys := DefaultLogModalKeyMap()

		switch {
		case key.Matches(keyMsg, keys.CloseSearch):
			m.searching = false
			m.searchInput.Blur()
			return m, nil
		case key.Matches(keyMsg, keys.ConfirmSearch):
			m.searching = false
			m.searchInput.Blur()
			m.searchQuery = m.searchInput.Value()
			m.updateSearchHits()
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.searchInput, cmd = m.searchInput.Update(msg)
	return m, cmd
}

// maxScroll returns the maximum scroll position.
func (m *LogModal) maxScroll() int {
	contentHeight := m.height - 6 // Account for borders, header, footer
	if len(m.entry.Logs) <= contentHeight {
		return 0
	}
	return len(m.entry.Logs) - contentHeight
}

// updateSearchHits finds all lines matching the search query.
func (m *LogModal) updateSearchHits() {
	m.searchHits = nil
	m.searchIdx = 0

	if m.searchQuery == "" {
		return
	}

	query := strings.ToLower(m.searchQuery)
	for i, log := range m.entry.Logs {
		if strings.Contains(strings.ToLower(log.Message), query) {
			m.searchHits = append(m.searchHits, i)
		}
	}

	// Jump to first hit
	if len(m.searchHits) > 0 {
		m.scrollToHit(0)
	}
}

// nextSearchHit jumps to the next search hit.
func (m *LogModal) nextSearchHit() {
	if len(m.searchHits) == 0 {
		return
	}
	m.searchIdx = (m.searchIdx + 1) % len(m.searchHits)
	m.scrollToHit(m.searchIdx)
}

// prevSearchHit jumps to the previous search hit.
func (m *LogModal) prevSearchHit() {
	if len(m.searchHits) == 0 {
		return
	}
	m.searchIdx--
	if m.searchIdx < 0 {
		m.searchIdx = len(m.searchHits) - 1
	}
	m.scrollToHit(m.searchIdx)
}

// scrollToHit scrolls to make the given search hit visible.
func (m *LogModal) scrollToHit(idx int) {
	if idx >= len(m.searchHits) {
		return
	}
	lineIdx := m.searchHits[idx]
	contentHeight := m.height - 6

	// Center the hit in the view if possible
	m.scrollPos = max(0, min(lineIdx-contentHeight/2, m.maxScroll()))
}

// View renders the log modal.
func (m *LogModal) View() string {
	borderColor := styles.Primary
	borderStyle := lipgloss.NewStyle().Foreground(borderColor)
	titleStyle := lipgloss.NewStyle().Foreground(styles.Primary).Bold(true)
	mutedStyle := lipgloss.NewStyle().Foreground(styles.TextMuted)

	contentWidth := m.width - 4
	contentHeight := m.height - 4 // Header + footer

	// Build title
	title := fmt.Sprintf(" Console Logs: %s ", m.entry.FunctionName)
	titleRendered := titleStyle.Render(title)

	// Top border with title
	leftPadding := (m.width - lipgloss.Width(titleRendered) - 2) / 2
	rightPadding := m.width - lipgloss.Width(titleRendered) - leftPadding - 2

	topBorder := borderStyle.Render("╭") +
		borderStyle.Render(strings.Repeat("─", max(0, leftPadding))) +
		titleRendered +
		borderStyle.Render(strings.Repeat("─", max(0, rightPadding))) +
		borderStyle.Render("╮")

	// Build content
	var content strings.Builder
	content.WriteString(topBorder)
	content.WriteString("\n")

	// Vertical border character
	verticalBorder := borderStyle.Render("│")

	// Calculate visible range
	visibleStart := m.scrollPos
	visibleEnd := min(visibleStart+contentHeight-2, len(m.entry.Logs)) // -2 for header/footer

	// Render log entries
	for i := visibleStart; i < visibleEnd; i++ {
		log := m.entry.Logs[i]
		logLine := m.formatLogLine(log, contentWidth-2, i)

		// Pad to full width
		lineWidth := lipgloss.Width(logLine)
		padding := ""
		if lineWidth < contentWidth {
			padding = strings.Repeat(" ", contentWidth-lineWidth)
		}

		content.WriteString(verticalBorder)
		content.WriteString(" ")
		content.WriteString(logLine)
		content.WriteString(padding)
		content.WriteString(" ")
		content.WriteString(verticalBorder)
		content.WriteString("\n")
	}

	// Fill remaining height with empty lines
	for i := visibleEnd - visibleStart; i < contentHeight-2; i++ {
		content.WriteString(verticalBorder)
		content.WriteString(" ")
		content.WriteString(strings.Repeat(" ", contentWidth))
		content.WriteString(" ")
		content.WriteString(verticalBorder)
		content.WriteString("\n")
	}

	// Build footer
	footer := m.buildFooter(contentWidth)
	footerWidth := lipgloss.Width(footer)
	footerPadding := max(0, contentWidth-footerWidth)

	content.WriteString(verticalBorder)
	content.WriteString(" ")
	content.WriteString(footer)
	content.WriteString(strings.Repeat(" ", footerPadding))
	content.WriteString(" ")
	content.WriteString(verticalBorder)
	content.WriteString("\n")

	// Bottom border
	closeHint := mutedStyle.Render(" Esc close ")
	bottomWidth := max(0, m.width-2-lipgloss.Width(closeHint))
	bottomBorder := borderStyle.Render("╰") +
		borderStyle.Render(strings.Repeat("─", bottomWidth)) +
		closeHint +
		borderStyle.Render("╯")
	content.WriteString(bottomBorder)

	return content.String()
}

// formatLogLine formats a single log entry for display.
func (m *LogModal) formatLogLine(log google.LogEntry, maxWidth, lineIdx int) string {
	icon := google.SeverityIcon(log.Severity)
	color := lipgloss.Color(google.SeverityColor(log.Severity))

	iconStyle := lipgloss.NewStyle().Foreground(color)
	msgStyle := lipgloss.NewStyle().Foreground(styles.TextPrimary)
	timeStyle := lipgloss.NewStyle().Foreground(styles.TextMuted)

	// Format timestamp
	timestamp := log.Timestamp.Format("15:04:05")

	// Check if this line is a search hit
	isHit := false
	for i, hitIdx := range m.searchHits {
		if hitIdx == lineIdx {
			isHit = true
			if i == m.searchIdx {
				// Current hit - highlight more
				msgStyle = lipgloss.NewStyle().Foreground(styles.Background).Background(styles.Accent)
			} else {
				// Other hit - subtle highlight
				msgStyle = lipgloss.NewStyle().Foreground(styles.TextPrimary).Bold(true)
			}
			break
		}
	}
	_ = isHit // Suppress unused variable warning

	// Format message
	msg := log.Message
	availableWidth := maxWidth - 15 // icon + timestamp + spaces
	if len(msg) > availableWidth {
		msg = msg[:availableWidth-3] + "..."
	}

	return iconStyle.Render(icon) + " " +
		timeStyle.Render(timestamp) + " " +
		msgStyle.Render(msg)
}

// buildFooter builds the footer with stats and hints.
func (m *LogModal) buildFooter(width int) string {
	mutedStyle := lipgloss.NewStyle().Foreground(styles.TextMuted)
	accentStyle := lipgloss.NewStyle().Foreground(styles.Accent)

	// Stats
	stats := fmt.Sprintf("%d entries", len(m.entry.Logs))
	if m.entry.Duration > 0 {
		stats += fmt.Sprintf(" │ %s", formatDuration(m.entry.Duration))
	}

	// Search status
	searchStatus := ""
	switch {
	case m.searching:
		searchStatus = " │ " + m.searchInput.View()
	case m.searchQuery != "" && len(m.searchHits) > 0:
		searchStatus = fmt.Sprintf(" │ %d/%d matches", m.searchIdx+1, len(m.searchHits))
	case m.searchQuery != "":
		searchStatus = " │ no matches"
	}

	// Scroll position
	scrollInfo := ""
	if m.maxScroll() > 0 {
		percent := float64(m.scrollPos) / float64(m.maxScroll()) * 100
		scrollInfo = fmt.Sprintf(" │ %.0f%%", percent)
	}

	// Hints
	hints := " │ j/k scroll │ / search"

	fullFooter := mutedStyle.Render(stats) +
		accentStyle.Render(searchStatus) +
		mutedStyle.Render(scrollInfo) +
		mutedStyle.Render(hints)

	return fullFooter
}

// ShortHelp returns key bindings for help.
func (m *LogModal) ShortHelp() []key.Binding {
	keys := DefaultLogModalKeyMap()
	if m.searching {
		return []key.Binding{keys.ConfirmSearch, keys.CloseSearch}
	}
	return []key.Binding{keys.Up, keys.Down, keys.Search, keys.Close}
}
