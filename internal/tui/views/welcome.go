// Package views contains all the view implementations for the Yukti TUI.
package views

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"yukti/internal/buildinfo"
	"yukti/internal/tui"
	"yukti/internal/tui/styles"
)

// ASCII art logo for Yukti
const logo = `
██╗   ██╗██╗   ██╗██╗  ██╗████████╗██╗
╚██╗ ██╔╝██║   ██║██║ ██╔╝╚══██╔══╝██║
 ╚████╔╝ ██║   ██║█████╔╝    ██║   ██║
  ╚██╔╝  ██║   ██║██╔═██╗    ██║   ██║
   ██║   ╚██████╔╝██║  ██╗   ██║   ██║
   ╚═╝    ╚═════╝ ╚═╝  ╚═╝   ╚═╝   ╚═╝`

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
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.width = msg.Width
		v.height = msg.Height
	case tea.KeyMsg:
		if msg.String() == "enter" {
			return v, tui.NavigateToProjects()
		}
	}
	return v, nil
}

// View implements tea.Model.
func (v *WelcomeView) View() string {
	taglineStyle := lipgloss.NewStyle().
		Foreground(styles.TextSecondary).
		Italic(true).
		MarginTop(1)

	// Feature card styles
	cardStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Border).
		Padding(1, 2).
		Width(24).
		Align(lipgloss.Center)

	cardTitleStyle := lipgloss.NewStyle().
		Foreground(styles.Primary).
		Bold(true)

	cardDescStyle := lipgloss.NewStyle().
		Foreground(styles.TextMuted).
		MarginTop(1)

	// CTA button style
	ctaStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Success).
		Foreground(styles.Success).
		Padding(0, 3).
		Bold(true).
		MarginTop(2)

	versionStyle := lipgloss.NewStyle().
		Foreground(styles.TextMuted).
		MarginTop(2)

	// Build the logo with gradient colors (Catppuccin palette)
	logoRendered := renderGradientLogo(logo)

	// Tagline
	tagline := taglineStyle.Render("The modern terminal interface for Google Apps Script")

	// Feature cards
	features := []struct {
		icon  string
		title string
		desc  string
	}{
		{"📁", "BROWSE", "Navigate projects\n& files with\nfuzzy search"},
		{"📝", "EDIT", "Syntax-aware\nediting with\nlive preview"},
		{"🚀", "DEPLOY", "One-click\ndeployments\n& versioning"},
	}

	cards := make([]string, 0, len(features))
	for _, f := range features {
		title := cardTitleStyle.Render(f.icon + " " + f.title)
		desc := cardDescStyle.Render(f.desc)
		card := cardStyle.Render(lipgloss.JoinVertical(lipgloss.Center, title, desc))
		cards = append(cards, card)
	}

	// Join cards horizontally with spacing
	cardRow := lipgloss.JoinHorizontal(lipgloss.Top, cards[0], "  ", cards[1], "  ", cards[2])

	// CTA button
	cta := ctaStyle.Render("⏎  Press Enter to continue")

	// Version info
	version := versionStyle.Render("v" + buildinfo.Version + " • Made with ♥")

	// Combine everything
	content := lipgloss.JoinVertical(
		lipgloss.Center,
		logoRendered,
		tagline,
		"",
		cardRow,
		"",
		cta,
		version,
	)

	// Center in the viewport
	return lipgloss.Place(
		v.width,
		v.height,
		lipgloss.Center,
		lipgloss.Center,
		content,
	)
}

// renderGradientLogo renders the logo with a vertical gradient effect.
// Uses Catppuccin Mocha color palette for a polished appearance.
func renderGradientLogo(logoText string) string {
	// Gradient colors from Catppuccin Mocha (top to bottom: Lavender → Blue → Sapphire)
	gradientColors := []lipgloss.Color{
		lipgloss.Color("#b4befe"), // Lavender
		lipgloss.Color("#89b4fa"), // Blue
		lipgloss.Color("#89b4fa"), // Blue
		lipgloss.Color("#74c7ec"), // Sapphire
		lipgloss.Color("#74c7ec"), // Sapphire
		lipgloss.Color("#89dceb"), // Sky
	}

	lines := strings.Split(strings.TrimPrefix(logoText, "\n"), "\n")
	styledLines := make([]string, 0, len(lines))

	for i, line := range lines {
		colorIdx := i
		if colorIdx >= len(gradientColors) {
			colorIdx = len(gradientColors) - 1
		}

		style := lipgloss.NewStyle().
			Foreground(gradientColors[colorIdx]).
			Bold(true)

		styledLines = append(styledLines, style.Render(line))
	}

	return strings.Join(styledLines, "\n")
}
