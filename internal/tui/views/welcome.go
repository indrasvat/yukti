// Package views contains all the view implementations for the Yukti TUI.
package views

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"yukti/internal/cli"
	"yukti/internal/tui/styles"
)

// WelcomeView is the initial landing screen.
type WelcomeView struct {
	width  int
	height int
}

// NewWelcomeView creates a new welcome view.
func NewWelcomeView() *WelcomeView {
	return &WelcomeView{
		width:  80,
		height: 24,
	}
}

// Title implements tui.View.
func (v *WelcomeView) Title() string {
	return "Welcome"
}

// ShortHelp implements tui.View.
func (v *WelcomeView) ShortHelp() []key.Binding {
	return []key.Binding{
		key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "get started"),
		),
		key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
	}
}

// Init implements tea.Model.
func (v *WelcomeView) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (v *WelcomeView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.WindowSizeMsg); ok {
		v.width = msg.Width
		v.height = msg.Height
	}
	return v, nil
}

// View implements tea.Model.
func (v *WelcomeView) View() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(styles.Primary).
		MarginBottom(1)

	subtitleStyle := lipgloss.NewStyle().
		Foreground(styles.TextSecondary).
		MarginBottom(2)

	versionStyle := lipgloss.NewStyle().
		Foreground(styles.TextMuted)

	infoStyle := lipgloss.NewStyle().
		Foreground(styles.TextSecondary).
		MarginTop(2)

	// Build the content
	title := titleStyle.Render("⚡ Yukti (युक्ति)")
	subtitle := subtitleStyle.Render("Beautiful TUI for Google Apps Script")

	versionInfo := versionStyle.Render(
		"Version: " + cli.Version +
			" | Commit: " + cli.Commit +
			" | Built: " + cli.BuildDate,
	)

	info := infoStyle.Render("Press Enter to get started, or ? for help")

	content := lipgloss.JoinVertical(
		lipgloss.Center,
		title,
		subtitle,
		versionInfo,
		"",
		info,
	)

	// Center the content
	return lipgloss.Place(
		v.width,
		v.height,
		lipgloss.Center,
		lipgloss.Center,
		content,
	)
}
