package views

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"yukti/internal/domain/project"
	tuiStyles "yukti/internal/tui/styles"
)

// CodeViewerView displays the source code of a file with syntax highlighting.
type CodeViewerView struct {
	file     project.File
	viewport viewport.Model
	content  string
	width    int
	height   int
	ready    bool
}

// NewCodeViewerView creates a new code viewer for the given file.
func NewCodeViewerView(file project.File) *CodeViewerView {
	return &CodeViewerView{
		file:   file,
		width:  80,
		height: 24,
	}
}

// GetFileName returns the name of the file being viewed.
func (v *CodeViewerView) GetFileName() string {
	return v.file.Name
}

// Title implements tui.View.
func (v *CodeViewerView) Title() string {
	return v.file.Name
}

// ShortHelp implements tui.View.
func (v *CodeViewerView) ShortHelp() []key.Binding {
	return []key.Binding{
		key.NewBinding(
			key.WithKeys("j", "down"),
			key.WithHelp("j/down", "scroll down"),
		),
		key.NewBinding(
			key.WithKeys("k", "up"),
			key.WithHelp("k/up", "scroll up"),
		),
		key.NewBinding(
			key.WithKeys("g"),
			key.WithHelp("g", "top"),
		),
		key.NewBinding(
			key.WithKeys("G"),
			key.WithHelp("G", "bottom"),
		),
	}
}

// Init implements tea.Model.
func (v *CodeViewerView) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (v *CodeViewerView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.width = msg.Width
		v.height = msg.Height
		headerHeight := 3 // File info header
		footerHeight := 0
		viewportHeight := v.height - headerHeight - footerHeight

		if !v.ready {
			v.viewport = viewport.New(v.width, viewportHeight)
			// No background - let the parent container handle it
			v.ready = true
			v.content = v.highlightCode()
			v.viewport.SetContent(v.content)
		} else {
			v.viewport.Width = v.width
			v.viewport.Height = viewportHeight
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "g":
			v.viewport.GotoTop()
			return v, nil
		case "G":
			v.viewport.GotoBottom()
			return v, nil
		}
	}

	v.viewport, cmd = v.viewport.Update(msg)
	return v, cmd
}

// View implements tea.Model.
func (v *CodeViewerView) View() string {
	if !v.ready {
		return "Loading..."
	}

	// File info header
	infoStyle := lipgloss.NewStyle().
		Foreground(tuiStyles.TextMuted).
		Padding(0, 2)

	lines := strings.Count(v.file.Source, "\n") + 1
	info := fmt.Sprintf("%s • %d lines", fileTypeLabel(v.file.Type), lines)

	// Add function count (don't list all functions to avoid header wrapping)
	if v.file.FunctionSet != nil && len(v.file.FunctionSet.Functions) > 0 {
		info += fmt.Sprintf(" • %d functions", len(v.file.FunctionSet.Functions))
	}

	header := infoStyle.Render(info)

	// Scroll indicator
	scrollInfo := lipgloss.NewStyle().
		Foreground(tuiStyles.TextMuted).
		Align(lipgloss.Right).
		Width(v.width - 4).
		Render(fmt.Sprintf("%d%%", int(v.viewport.ScrollPercent()*100)))

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		scrollInfo,
		v.viewport.View(),
	)
}

// highlightCode applies syntax highlighting to the source code.
func (v *CodeViewerView) highlightCode() string {
	source := v.file.Source
	if source == "" {
		return lipgloss.NewStyle().
			Foreground(tuiStyles.TextMuted).
			Italic(true).
			Render("(empty file)")
	}

	// Determine lexer based on file type
	var lexer chroma.Lexer
	switch v.file.Type {
	case project.FileTypeServer:
		lexer = lexers.Get("javascript")
	case project.FileTypeHTML:
		lexer = lexers.Get("html")
	case project.FileTypeJSON:
		lexer = lexers.Get("json")
	default:
		lexer = lexers.Get("javascript")
	}

	if lexer == nil {
		lexer = lexers.Fallback
	}
	lexer = chroma.Coalesce(lexer)

	// Use a dark theme that works well in terminal
	// Note: Using "native" style which has no background color to avoid bleed
	style := styles.Get("native")
	if style == nil {
		style = styles.Fallback
	}

	// Format with ANSI colors for terminal (terminal256 for better colors)
	formatter := formatters.Get("terminal256")
	if formatter == nil {
		formatter = formatters.Fallback
	}

	// Tokenize and format
	iterator, err := lexer.Tokenise(nil, source)
	if err != nil {
		return v.addLineNumbers(source)
	}

	var buf bytes.Buffer
	err = formatter.Format(&buf, style, iterator)
	if err != nil {
		return v.addLineNumbers(source)
	}

	return v.addLineNumbers(buf.String())
}

// addLineNumbers adds line numbers to the source code.
func (v *CodeViewerView) addLineNumbers(source string) string {
	lines := strings.Split(source, "\n")
	maxLineNum := len(lines)
	numWidth := len(fmt.Sprintf("%d", maxLineNum))

	lineNumStyle := lipgloss.NewStyle().
		Foreground(tuiStyles.TextMuted).
		Width(numWidth + 2).
		Align(lipgloss.Right)

	separator := " │ "
	separatorWidth := 3

	// Calculate available width for source code
	// v.width is the panel width; subtract line number, separator, and padding
	availableCodeWidth := max(20, v.width-numWidth-2-separatorWidth-4) // -4 for borders and padding

	separatorStyled := lipgloss.NewStyle().
		Foreground(tuiStyles.Surface).
		Render(separator)

	var result strings.Builder
	for i, line := range lines {
		// Reset ANSI at start of each line to prevent bleed from previous line
		result.WriteString("\033[0m")
		lineNum := lineNumStyle.Render(fmt.Sprintf("%d", i+1))
		result.WriteString(lineNum)
		result.WriteString(separatorStyled)

		// Truncate line if it exceeds available width
		if lipgloss.Width(line) > availableCodeWidth {
			line = truncateCode(line, availableCodeWidth)
		}
		result.WriteString(line)
		// Reset ANSI at end of line to prevent bleed to next line
		result.WriteString("\033[0m")
		if i < len(lines)-1 {
			result.WriteString("\n")
		}
	}

	return result.String()
}

// truncateCode truncates a line of code to fit within maxWidth.
// Handles ANSI-colored strings by measuring display width.
func truncateCode(s string, maxWidth int) string {
	if maxWidth <= 3 {
		return "…"
	}

	// For ANSI strings, we need to be careful about where we cut
	// Use lipgloss.Width to measure display width
	if lipgloss.Width(s) <= maxWidth {
		return s
	}

	// Simple rune-based truncation - may cut in middle of ANSI codes
	// but works for most cases. For complex ANSI, would need smarter handling.
	runes := []rune(s)
	if len(runes) <= maxWidth-1 {
		return s
	}

	// Find a good cut point by measuring progressively
	for cutPoint := maxWidth - 1; cutPoint > 0; cutPoint-- {
		truncated := string(runes[:cutPoint]) + "…"
		if lipgloss.Width(truncated) <= maxWidth {
			return truncated
		}
	}

	return "…"
}

// fileTypeLabel returns a human-readable label for the file type.
func fileTypeLabel(ft project.FileType) string {
	switch ft {
	case project.FileTypeServer:
		return "JavaScript (Server)"
	case project.FileTypeHTML:
		return "HTML"
	case project.FileTypeJSON:
		return "JSON"
	default:
		return string(ft)
	}
}
