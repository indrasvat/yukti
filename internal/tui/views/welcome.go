// Package views contains all the view implementations for the Yukti TUI.
package views

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"yukti/internal/buildinfo"
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
	// Logo box with rounded border
	logoBoxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Primary).
		Padding(1, 4).
		Foreground(styles.Primary).
		Bold(true).
		Align(lipgloss.Center)

	// Feature list
	featureStyle := lipgloss.NewStyle().
		Foreground(styles.TextSecondary)

	featureIconStyle := lipgloss.NewStyle().
		Foreground(styles.Secondary)

	mutedStyle := lipgloss.NewStyle().
		Foreground(styles.TextMuted)

	hintStyle := lipgloss.NewStyle().
		Foreground(styles.Info).
		Italic(true)

	// Build features list
	features := []struct {
		icon string
		text string
	}{
		{"📁", "Browse and manage your Apps Script projects"},
		{"📝", "View and edit script files with syntax highlighting"},
		{"🚀", "Deploy and manage versions"},
		{"📊", "Monitor execution metrics and logs"},
	}

	featureLines := make([]string, 0, len(features))
	for _, f := range features {
		line := featureIconStyle.Render(f.icon) + "  " + featureStyle.Render(f.text)
		featureLines = append(featureLines, line)
	}
	featureList := strings.Join(featureLines, "\n")

	// Version info
	versionInfo := mutedStyle.Render(
		"v" + buildinfo.Version + " • " + buildinfo.Commit,
	)

	// Hint
	hint := hintStyle.Render("Press Enter to get started")

	// Build the logo box
	logoContent := "⚡ YUKTI\n\nBeautiful TUI for\nGoogle Apps Script"
	logoBox := logoBoxStyle.Render(logoContent)

	// Build the content
	content := lipgloss.JoinVertical(
		lipgloss.Center,
		logoBox,
		"",
		featureList,
		"",
		"",
		hint,
		"",
		versionInfo,
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
