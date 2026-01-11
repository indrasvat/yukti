// Package styles defines the visual theme and styling for the Yukti TUI.
package styles

import "github.com/charmbracelet/lipgloss"

// Color palette - Catppuccin Mocha inspired theme (brighter, warmer dark theme).
var (
	// Primary colors
	Primary   = lipgloss.Color("#CBA6F7") // Mauve (brighter purple)
	Secondary = lipgloss.Color("#A6E3A1") // Green
	Accent    = lipgloss.Color("#F9E2AF") // Yellow

	// Background colors
	Background = lipgloss.Color("#1E1E2E") // Base
	Surface    = lipgloss.Color("#313244") // Surface0
	Overlay    = lipgloss.Color("#45475A") // Surface1

	// Text colors
	TextPrimary   = lipgloss.Color("#CDD6F4") // Text (bright)
	TextSecondary = lipgloss.Color("#BAC2DE") // Subtext1
	TextMuted     = lipgloss.Color("#6C7086") // Overlay1

	// Status colors
	Success = lipgloss.Color("#A6E3A1") // Green
	Warning = lipgloss.Color("#FAB387") // Peach
	Error   = lipgloss.Color("#F38BA8") // Red
	Info    = lipgloss.Color("#89B4FA") // Blue

	// Border colors
	Border      = lipgloss.Color("#45475A") // Surface1
	BorderFocus = lipgloss.Color("#CBA6F7") // Mauve

	// Special colors
	White = lipgloss.Color("#FFFFFF")
	Black = lipgloss.Color("#000000")

	// Additional accent colors
	Teal   = lipgloss.Color("#94E2D5") // Teal
	Pink   = lipgloss.Color("#F5C2E7") // Pink
	Sky    = lipgloss.Color("#89DCEB") // Sky
	Lavender = lipgloss.Color("#B4BEFE") // Lavender
)
