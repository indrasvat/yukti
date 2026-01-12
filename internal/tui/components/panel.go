package components

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"

	"yukti/internal/tui/styles"
)

// Panel renders a bordered panel with a numbered title like lazygit.
type Panel struct {
	number   int
	title    string
	focused  bool
	width    int
	height   int
	content  string
	info     string // Right-aligned info in title bar (e.g., "3 of 15")
	subtitle string // Optional subtitle after title
}

// NewPanel creates a new panel with a number and title.
func NewPanel(number int, title string) *Panel {
	return &Panel{
		number: number,
		title:  title,
		width:  40,
		height: 10,
	}
}

// SetFocused sets whether the panel is focused.
func (p *Panel) SetFocused(focused bool) {
	p.focused = focused
}

// SetSize sets the panel dimensions.
func (p *Panel) SetSize(width, height int) {
	p.width = width
	p.height = height
}

// SetContent sets the panel content.
func (p *Panel) SetContent(content string) {
	p.content = content
}

// SetInfo sets the right-aligned info text (e.g., "3 of 15").
func (p *Panel) SetInfo(info string) {
	p.info = info
}

// SetSubtitle sets an optional subtitle after the title.
func (p *Panel) SetSubtitle(subtitle string) {
	p.subtitle = subtitle
}

// View renders the panel.
func (p *Panel) View() string {
	// Border color based on focus
	borderColor := styles.Border
	titleColor := styles.TextMuted
	if p.focused {
		borderColor = styles.Primary
		titleColor = styles.Primary
	}

	// Build title: [1]-Files
	titleStyle := lipgloss.NewStyle().
		Foreground(titleColor).
		Bold(true)

	numberStyle := lipgloss.NewStyle().
		Foreground(titleColor)

	subtitleStyle := lipgloss.NewStyle().
		Foreground(styles.TextMuted)

	infoStyle := lipgloss.NewStyle().
		Foreground(styles.TextMuted)

	// Create the title string
	title := fmt.Sprintf("[%d]─%s", p.number, p.title)
	titleStr := numberStyle.Render(fmt.Sprintf("[%d]", p.number)) +
		titleStyle.Render("─"+p.title)

	if p.subtitle != "" {
		titleStr += subtitleStyle.Render(" " + p.subtitle)
	}

	// Calculate available width for the title bar
	contentWidth := p.width - 4 // Account for border
	titleWidth := lipgloss.Width(titleStr)
	infoWidth := lipgloss.Width(p.info)

	// Build top border with title
	topBorder := p.buildTitleBorder(title, contentWidth, borderColor)

	// Build info section if present
	var infoSection string
	if p.info != "" {
		// Right-align info
		availableSpace := contentWidth - titleWidth - infoWidth - 2
		if availableSpace > 0 {
			infoSection = infoStyle.Render(p.info)
		}
	}

	// Content area style
	contentStyle := lipgloss.NewStyle().
		Width(contentWidth).
		Height(p.height - 2). // Account for top and bottom borders
		MaxHeight(p.height - 2)

	// Build the panel using custom border rendering
	return p.renderWithTitleBorder(topBorder, infoSection, contentStyle.Render(p.content), borderColor)
}

// buildTitleBorder creates a top border with the title embedded.
func (p *Panel) buildTitleBorder(title string, width int, color lipgloss.Color) string {
	borderStyle := lipgloss.NewStyle().Foreground(color)

	// Unicode box drawing characters for rounded corners
	topLeft := "╭"
	topRight := "╮"
	horizontal := "─"

	titleLen := len(title)
	remainingWidth := max(0, width-titleLen-2) // -2 for corners

	// Build: ╭[1]─Title────────────╮
	border := topLeft + title
	for range remainingWidth {
		border += horizontal
	}
	border += topRight

	return borderStyle.Render(border)
}

// renderWithTitleBorder renders the panel with a custom title border.
func (p *Panel) renderWithTitleBorder(topBorder, info, content string, borderColor lipgloss.Color) string {
	borderStyle := lipgloss.NewStyle().Foreground(borderColor)

	// Border characters
	vertical := "│"
	bottomLeft := "╰"
	bottomRight := "╯"
	horizontal := "─"

	contentWidth := p.width - 4
	lines := splitLines(content)

	var result string
	result += topBorder + "\n"

	// Add info line if present
	if info != "" {
		infoPadding := max(0, contentWidth-lipgloss.Width(info))
		infoLine := borderStyle.Render(vertical) +
			" " + info +
			lipgloss.NewStyle().Width(infoPadding).Render("") +
			borderStyle.Render(vertical)
		result += infoLine + "\n"
	}

	// Content lines
	for i := 0; i < p.height-2; i++ {
		var line string
		if i < len(lines) {
			line = lines[i]
		}

		// Pad or truncate line to fit
		lineWidth := lipgloss.Width(line)
		if lineWidth < contentWidth {
			line += lipgloss.NewStyle().Width(contentWidth - lineWidth).Render("")
		} else if lineWidth > contentWidth {
			line = truncate(line, contentWidth)
		}

		result += borderStyle.Render(vertical) + " " + line + " " + borderStyle.Render(vertical) + "\n"
	}

	// Bottom border
	bottomBorder := bottomLeft
	for i := 0; i < p.width-2; i++ {
		bottomBorder += horizontal
	}
	bottomBorder += bottomRight
	result += borderStyle.Render(bottomBorder)

	return result
}

// splitLines splits content into lines.
func splitLines(s string) []string {
	if s == "" {
		return nil
	}
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

// truncate truncates a string to the given width.
func truncate(s string, width int) string {
	if width <= 0 {
		return ""
	}
	if lipgloss.Width(s) <= width {
		return s
	}
	// Simple truncation - could be improved for unicode
	runes := []rune(s)
	if len(runes) > width-1 {
		return string(runes[:width-1]) + "…"
	}
	return s
}
