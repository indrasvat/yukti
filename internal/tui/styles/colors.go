// Package styles defines the visual theme and styling for the Yukti TUI.
package styles

import "github.com/charmbracelet/lipgloss"

// Color palette - Tokyo Night inspired theme.
var (
	// Primary colors
	Primary   = lipgloss.Color("#7C3AED") // Purple
	Secondary = lipgloss.Color("#10B981") // Green
	Accent    = lipgloss.Color("#F59E0B") // Amber

	// Background colors
	Background = lipgloss.Color("#1A1B26")
	Surface    = lipgloss.Color("#24283B")
	Overlay    = lipgloss.Color("#414868")

	// Text colors
	TextPrimary   = lipgloss.Color("#C0CAF5")
	TextSecondary = lipgloss.Color("#9AA5CE")
	TextMuted     = lipgloss.Color("#565F89")

	// Status colors
	Success = lipgloss.Color("#9ECE6A")
	Warning = lipgloss.Color("#E0AF68")
	Error   = lipgloss.Color("#F7768E")
	Info    = lipgloss.Color("#7AA2F7")

	// Border colors
	Border      = lipgloss.Color("#414868")
	BorderFocus = lipgloss.Color("#7C3AED")

	// Special colors
	White = lipgloss.Color("#FFFFFF")
	Black = lipgloss.Color("#000000")
)
