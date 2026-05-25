package styles

import "github.com/charmbracelet/lipgloss"

// Base styles for common elements.
var (
	BaseStyle = lipgloss.NewStyle().
			Background(Background).
			Foreground(TextPrimary)

	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(Primary).
			MarginBottom(1)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(TextSecondary)

	MutedStyle = lipgloss.NewStyle().
			Foreground(TextMuted)
)

// Component styles for layout elements.
var (
	HeaderStyle = lipgloss.NewStyle().
			Background(Background).
			Foreground(TextPrimary).
			Padding(0, 2).
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderForeground(Border)

	FooterStyle = lipgloss.NewStyle().
			Background(Background).
			Foreground(TextMuted).
			Padding(0, 2).
			BorderStyle(lipgloss.NormalBorder()).
			BorderTop(true).
			BorderForeground(Border)

	PanelStyle = lipgloss.NewStyle().
			Background(Surface).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(Border).
			Padding(1)

	FocusedPanelStyle = lipgloss.NewStyle().
				Background(Surface).
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(BorderFocus).
				Padding(1)
)

// List styles for list items.
var (
	ListItemStyle = lipgloss.NewStyle().
			PaddingLeft(2)

	SelectedItemStyle = lipgloss.NewStyle().
				PaddingLeft(2).
				Background(Overlay).
				Foreground(Primary)
)

// Button styles.
var (
	ButtonStyle = lipgloss.NewStyle().
			Foreground(TextPrimary).
			Background(Surface).
			Padding(0, 2).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(Border)

	PrimaryButtonStyle = lipgloss.NewStyle().
				Foreground(White).
				Background(Primary).
				Padding(0, 2).
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(Primary)
)

// Input styles.
var (
	InputStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(Border).
			Padding(0, 1)

	FocusedInputStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(BorderFocus).
				Padding(0, 1)
)

// Status badge helpers.

// SuccessBadge renders a success indicator with the given text.
func SuccessBadge(text string) string {
	return lipgloss.NewStyle().
		Foreground(Success).
		Render("✓ " + text)
}

// ErrorBadge renders an error indicator with the given text.
func ErrorBadge(text string) string {
	return lipgloss.NewStyle().
		Foreground(Error).
		Render("✗ " + text)
}

// WarningBadge renders a warning indicator with the given text.
func WarningBadge(text string) string {
	return lipgloss.NewStyle().
		Foreground(Warning).
		Render("⚠ " + text)
}

// InfoBadge renders an info indicator with the given text.
func InfoBadge(text string) string {
	return lipgloss.NewStyle().
		Foreground(Info).
		Render("ℹ " + text)
}

// Icon styles for project types.
var (
	StandaloneIcon = "📄"
	BoundIcon      = "📎"
	WebAppIcon     = "🌐"
	APIIcon        = "🔌"
	AddOnIcon      = "📋"
)
