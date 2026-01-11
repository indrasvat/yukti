// Package tui provides the terminal user interface for Yukti.
package tui

import "github.com/charmbracelet/bubbles/key"

// KeyMap defines all keybindings for the application.
type KeyMap struct {
	// Navigation
	Up     key.Binding
	Down   key.Binding
	Left   key.Binding
	Right  key.Binding
	Home   key.Binding
	End    key.Binding
	PageUp key.Binding
	PageDn key.Binding

	// Actions
	Enter   key.Binding
	Back    key.Binding
	Quit    key.Binding
	Help    key.Binding
	Refresh key.Binding

	// Search & Filter
	Search key.Binding
	Filter key.Binding

	// Project operations
	New    key.Binding
	Delete key.Binding
	Rename key.Binding
	Copy   key.Binding

	// File operations
	Push key.Binding
	Pull key.Binding

	// Deployment operations
	Deploy key.Binding

	// Command palette
	CommandPalette key.Binding

	// Tab navigation
	Tab      key.Binding
	ShiftTab key.Binding
}

// DefaultKeyMap returns the default keybindings.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		// Navigation - vim-style + arrows
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		Left: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("←/h", "left"),
		),
		Right: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("→/l", "right"),
		),
		Home: key.NewBinding(
			key.WithKeys("home", "g"),
			key.WithHelp("home/g", "first"),
		),
		End: key.NewBinding(
			key.WithKeys("end", "G"),
			key.WithHelp("end/G", "last"),
		),
		PageUp: key.NewBinding(
			key.WithKeys("pgup", "ctrl+u"),
			key.WithHelp("pgup", "page up"),
		),
		PageDn: key.NewBinding(
			key.WithKeys("pgdown", "ctrl+d"),
			key.WithHelp("pgdn", "page down"),
		),

		// Actions
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select"),
		),
		Back: key.NewBinding(
			key.WithKeys("esc", "backspace"),
			key.WithHelp("esc", "back"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("r", "ctrl+r"),
			key.WithHelp("r", "refresh"),
		),

		// Search & Filter
		Search: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "search"),
		),
		Filter: key.NewBinding(
			key.WithKeys("f"),
			key.WithHelp("f", "filter"),
		),

		// Project operations
		New: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "new"),
		),
		Delete: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "delete"),
		),
		Rename: key.NewBinding(
			key.WithKeys("R"),
			key.WithHelp("R", "rename"),
		),
		Copy: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "copy"),
		),

		// File operations
		Push: key.NewBinding(
			key.WithKeys("p"),
			key.WithHelp("p", "push"),
		),
		Pull: key.NewBinding(
			key.WithKeys("P"),
			key.WithHelp("P", "pull"),
		),

		// Deployment
		Deploy: key.NewBinding(
			key.WithKeys("D"),
			key.WithHelp("D", "deploy"),
		),

		// Command palette
		CommandPalette: key.NewBinding(
			key.WithKeys("ctrl+p"),
			key.WithHelp("ctrl+p", "commands"),
		),

		// Tab navigation
		Tab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "next pane"),
		),
		ShiftTab: key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("shift+tab", "prev pane"),
		),
	}
}

// Keys is the global key map instance.
var Keys = DefaultKeyMap()
