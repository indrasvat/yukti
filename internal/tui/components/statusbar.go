package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"yukti/internal/tui/styles"
)

// StatusItem represents a single item in the status bar.
type StatusItem struct {
	Key  string
	Desc string
}

// StatusBar renders a context-sensitive status bar like lazygit.
type StatusBar struct {
	width int
	items []StatusItem
	info  string // Right-aligned info like "3 of 15"
}

// NewStatusBar creates a new status bar.
func NewStatusBar() *StatusBar {
	return &StatusBar{
		width: 80,
	}
}

// SetWidth sets the status bar width.
func (s *StatusBar) SetWidth(width int) {
	s.width = width
}

// SetItems sets the keybinding hints to display.
func (s *StatusBar) SetItems(items []StatusItem) {
	s.items = items
}

// SetInfo sets the right-aligned info text (e.g., "3 of 15").
func (s *StatusBar) SetInfo(info string) {
	s.info = info
}

// View renders the status bar.
func (s *StatusBar) View() string {
	keyStyle := lipgloss.NewStyle().
		Foreground(styles.Primary).
		Bold(true)

	descStyle := lipgloss.NewStyle().
		Foreground(styles.TextSecondary)

	sepStyle := lipgloss.NewStyle().
		Foreground(styles.TextMuted)

	infoStyle := lipgloss.NewStyle().
		Foreground(styles.TextMuted)

	// Build keybinding hints
	parts := make([]string, 0, len(s.items))
	for _, item := range s.items {
		part := keyStyle.Render(item.Key) + " " + descStyle.Render(item.Desc)
		parts = append(parts, part)
	}

	left := strings.Join(parts, sepStyle.Render(" │ "))

	// Calculate spacing for right-aligned info
	leftWidth := lipgloss.Width(left)
	infoWidth := lipgloss.Width(s.info)
	spacing := max(1, s.width-leftWidth-infoWidth-4)

	spacer := strings.Repeat(" ", spacing)
	right := infoStyle.Render(s.info)

	return left + spacer + right
}

// CommonStatusItems provides standard keybinding items.
var CommonStatusItems = struct {
	Navigation []StatusItem
	Selection  []StatusItem
	Actions    []StatusItem
}{
	Navigation: []StatusItem{
		{Key: "↑↓/jk", Desc: "navigate"},
		{Key: "g/G", Desc: "top/bottom"},
	},
	Selection: []StatusItem{
		{Key: "enter", Desc: "select"},
		{Key: "esc", Desc: "back"},
	},
	Actions: []StatusItem{
		{Key: "?", Desc: "help"},
		{Key: "q", Desc: "quit"},
	},
}
