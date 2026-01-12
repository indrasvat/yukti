package components

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"yukti/internal/tui/styles"
)

// HelpSection groups related keybindings.
type HelpSection struct {
	Title    string
	Bindings []HelpBinding
}

// HelpBinding represents a single keybinding in the help modal.
type HelpBinding struct {
	Key  string
	Desc string
}

// HelpModal displays a modal with all keybindings.
type HelpModal struct {
	visible  bool
	sections []HelpSection
	width    int
	height   int
}

// NewHelpModal creates a new help modal.
func NewHelpModal() *HelpModal {
	return &HelpModal{
		sections: defaultHelpSections(),
	}
}

// defaultHelpSections returns the default keybindings.
func defaultHelpSections() []HelpSection {
	return []HelpSection{
		{
			Title: "Navigation",
			Bindings: []HelpBinding{
				{Key: "j/↓", Desc: "Move down"},
				{Key: "k/↑", Desc: "Move up"},
				{Key: "g", Desc: "Go to top"},
				{Key: "G", Desc: "Go to bottom"},
				{Key: "tab", Desc: "Switch pane"},
				{Key: "esc", Desc: "Go back"},
			},
		},
		{
			Title: "Actions",
			Bindings: []HelpBinding{
				{Key: "enter", Desc: "Select / Open"},
				{Key: "ctrl+p", Desc: "Fuzzy finder"},
				{Key: "/", Desc: "Filter / Search"},
				{Key: "r", Desc: "Refresh"},
			},
		},
		{
			Title: "Global",
			Bindings: []HelpBinding{
				{Key: "?", Desc: "Toggle help"},
				{Key: "q", Desc: "Quit"},
			},
		},
	}
}

// SetSections sets custom help sections.
func (h *HelpModal) SetSections(sections []HelpSection) {
	h.sections = sections
}

// IsVisible returns whether the modal is visible.
func (h *HelpModal) IsVisible() bool {
	return h.visible
}

// Show shows the modal.
func (h *HelpModal) Show() {
	h.visible = true
}

// Hide hides the modal.
func (h *HelpModal) Hide() {
	h.visible = false
}

// Toggle toggles the modal visibility.
func (h *HelpModal) Toggle() {
	h.visible = !h.visible
}

// ShortHelp returns help for the modal itself.
func (h *HelpModal) ShortHelp() []key.Binding {
	return []key.Binding{
		key.NewBinding(
			key.WithKeys("esc", "?"),
			key.WithHelp("esc/?", "close"),
		),
	}
}

// Update handles input for the modal.
func (h *HelpModal) Update(msg tea.Msg) (*HelpModal, tea.Cmd) {
	if !h.visible {
		return h, nil
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h.width = msg.Width
		h.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "?", "q", "enter":
			h.Hide()
		}
	}

	return h, nil
}

// View renders the help modal.
func (h *HelpModal) View() string {
	if !h.visible {
		return ""
	}

	// Modal dimensions
	modalWidth := 50
	if h.width > 0 && h.width < 60 {
		modalWidth = h.width - 10
	}

	// Styles
	titleStyle := lipgloss.NewStyle().
		Foreground(styles.Primary).
		Bold(true).
		MarginBottom(1)

	sectionTitleStyle := lipgloss.NewStyle().
		Foreground(styles.TextPrimary).
		Bold(true).
		MarginTop(1)

	keyStyle := lipgloss.NewStyle().
		Foreground(styles.Info).
		Bold(true).
		Width(12)

	descStyle := lipgloss.NewStyle().
		Foreground(styles.TextSecondary)

	hintStyle := lipgloss.NewStyle().
		Foreground(styles.TextMuted).
		Italic(true).
		MarginTop(1)

	modalStyle := lipgloss.NewStyle().
		Background(styles.Surface).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Primary).
		Padding(1, 2).
		Width(modalWidth)

	// Build content
	var content strings.Builder
	content.WriteString(titleStyle.Render("⌨ Keybindings"))
	content.WriteString("\n")

	for _, section := range h.sections {
		content.WriteString(sectionTitleStyle.Render("─── " + section.Title + " ───"))
		content.WriteString("\n")

		for _, binding := range section.Bindings {
			line := keyStyle.Render(binding.Key) + descStyle.Render(binding.Desc)
			content.WriteString(line)
			content.WriteString("\n")
		}
	}

	content.WriteString(hintStyle.Render("Press ? or esc to close"))

	return modalStyle.Render(content.String())
}

// HelpToggleMsg is sent when help should be toggled.
type HelpToggleMsg struct{}
