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
			v.viewport.Style = lipgloss.NewStyle().
				Background(tuiStyles.Background)
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

	// Add function info if available
	if v.file.FunctionSet != nil && len(v.file.FunctionSet.Functions) > 0 {
		funcs := make([]string, 0, len(v.file.FunctionSet.Functions))
		for _, fn := range v.file.FunctionSet.Functions {
			funcs = append(funcs, fn.Name+"()")
		}
		info += fmt.Sprintf(" • Functions: %s", strings.Join(funcs, ", "))
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
	style := styles.Get("catppuccin-mocha")
	if style == nil {
		style = styles.Fallback
	}

	// Format with ANSI colors for terminal
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

	separator := lipgloss.NewStyle().
		Foreground(tuiStyles.Surface).
		Render(" │ ")

	var result strings.Builder
	for i, line := range lines {
		lineNum := lineNumStyle.Render(fmt.Sprintf("%d", i+1))
		result.WriteString(lineNum)
		result.WriteString(separator)
		result.WriteString(line)
		if i < len(lines)-1 {
			result.WriteString("\n")
		}
	}

	return result.String()
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
