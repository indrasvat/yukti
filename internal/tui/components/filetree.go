package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"yukti/internal/domain/project"
	"yukti/internal/tui/styles"
)

// FileTree displays a list of files with selection and optional filtering.
type FileTree struct {
	files    []project.File
	filtered []project.File
	selected int
	offset   int // For virtual scrolling

	// Dimensions
	width  int
	height int

	// Filter
	filterText string

	// Styling
	titleStyle       lipgloss.Style
	itemStyle        lipgloss.Style
	selectedStyle    lipgloss.Style
	typeStyle        lipgloss.Style
	countStyle       lipgloss.Style
	emptyStyle       lipgloss.Style
	sectionStyle     lipgloss.Style
	functionStyle    lipgloss.Style
	selectedFuncIcon string
}

// NewFileTree creates a new file tree component.
func NewFileTree(files []project.File) *FileTree {
	ft := &FileTree{
		files:    files,
		filtered: files,
		selected: 0,
		offset:   0,
		width:    30,
		height:   20,
		titleStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(styles.Primary).
			Padding(0, 1),
		itemStyle: lipgloss.NewStyle().
			Padding(0, 1),
		selectedStyle: lipgloss.NewStyle().
			Background(styles.Surface).
			Foreground(styles.Primary).
			Bold(true).
			Padding(0, 1),
		typeStyle: lipgloss.NewStyle().
			Foreground(styles.TextMuted),
		countStyle: lipgloss.NewStyle().
			Foreground(styles.TextSecondary),
		emptyStyle: lipgloss.NewStyle().
			Foreground(styles.TextMuted).
			Italic(true).
			Padding(1),
		sectionStyle: lipgloss.NewStyle().
			Foreground(styles.TextMuted).
			Bold(true).
			Padding(0, 1).
			MarginTop(1),
		functionStyle: lipgloss.NewStyle().
			Foreground(styles.TextSecondary).
			Padding(0, 3),
		selectedFuncIcon: "▸",
	}
	return ft
}

// SetFiles updates the file list.
func (ft *FileTree) SetFiles(files []project.File) {
	ft.files = files
	ft.applyFilter()
	if ft.selected >= len(ft.filtered) {
		ft.selected = max(0, len(ft.filtered)-1)
	}
}

// Init implements tea.Model.
func (ft *FileTree) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (ft *FileTree) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		ft.width = msg.Width
		ft.height = msg.Height

	case childSizeMsg:
		ft.width = msg.leftWidth
		ft.height = msg.leftHeight

	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			if ft.selected < len(ft.filtered)-1 {
				ft.selected++
				ft.ensureVisible()
			}
			return ft, nil

		case "k", "up":
			if ft.selected > 0 {
				ft.selected--
				ft.ensureVisible()
			}
			return ft, nil

		case "g":
			ft.selected = 0
			ft.offset = 0
			return ft, nil

		case "G":
			ft.selected = max(0, len(ft.filtered)-1)
			ft.ensureVisible()
			return ft, nil

		case "enter":
			if ft.selected < len(ft.filtered) {
				cmd := ft.selectFile()
				return ft, cmd
			}
		}
	}

	return ft, nil
}

// View implements tea.Model.
func (ft *FileTree) View() string {
	if len(ft.filtered) == 0 {
		return ft.emptyStyle.Render("No files found")
	}

	var b strings.Builder

	// Header
	header := ft.titleStyle.Render(fmt.Sprintf("FILES (%d)", len(ft.filtered)))
	b.WriteString(header)
	b.WriteString("\n")

	// Calculate visible range (virtual scrolling)
	visibleHeight := max(1, ft.height-4) // Account for header and padding

	start := ft.offset
	end := min(start+visibleHeight, len(ft.filtered))

	// Group files by type for display
	for i := start; i < end; i++ {
		file := ft.filtered[i]
		isSelected := i == ft.selected

		// File icon based on type
		icon := ft.fileIcon(file.Type)

		// File name (truncate if too long)
		name := file.Name
		maxNameLen := ft.width - 8
		if maxNameLen > 0 && len(name) > maxNameLen {
			name = name[:maxNameLen-3] + "..."
		}

		// Line and function count
		lines := strings.Count(file.Source, "\n") + 1
		var meta string
		if file.FunctionSet != nil && len(file.FunctionSet.Functions) > 0 {
			meta = fmt.Sprintf("%d fn", len(file.FunctionSet.Functions))
		} else {
			meta = fmt.Sprintf("%d ln", lines)
		}

		// Build line
		var line string
		if isSelected {
			indicator := ft.selectedFuncIcon + " "
			content := fmt.Sprintf("%s%s %s", indicator, icon, name)
			metaStyled := ft.countStyle.Render(meta)
			line = ft.selectedStyle.Width(ft.width - 2).Render(content + " " + metaStyled)
		} else {
			content := fmt.Sprintf("  %s %s", icon, name)
			metaStyled := ft.countStyle.Render(meta)
			line = ft.itemStyle.Width(ft.width - 2).Render(content + " " + metaStyled)
		}

		b.WriteString(line)
		if i < end-1 {
			b.WriteString("\n")
		}
	}

	// Scroll indicator if needed
	if len(ft.filtered) > visibleHeight {
		scrollPct := float64(ft.offset) / float64(max(1, len(ft.filtered)-visibleHeight))
		scrollIndicator := ft.typeStyle.Render(fmt.Sprintf(" [%d/%d]", ft.selected+1, len(ft.filtered)))
		b.WriteString("\n")
		b.WriteString(scrollIndicator)
		_ = scrollPct // Can be used for scroll bar later
	}

	return b.String()
}

// fileIcon returns an icon for the file type.
func (ft *FileTree) fileIcon(fileType project.FileType) string {
	switch fileType {
	case project.FileTypeServer:
		return "📄"
	case project.FileTypeHTML:
		return "🌐"
	case project.FileTypeJSON:
		return "📋"
	default:
		return "📄"
	}
}

// applyFilter filters files based on filterText.
func (ft *FileTree) applyFilter() {
	if ft.filterText == "" {
		ft.filtered = ft.files
		return
	}

	query := strings.ToLower(ft.filterText)
	ft.filtered = make([]project.File, 0)

	for _, file := range ft.files {
		// Match on file name
		if strings.Contains(strings.ToLower(file.Name), query) {
			ft.filtered = append(ft.filtered, file)
			continue
		}

		// Match on function names
		if file.FunctionSet != nil {
			for _, fn := range file.FunctionSet.Functions {
				if strings.Contains(strings.ToLower(fn.Name), query) {
					ft.filtered = append(ft.filtered, file)
					break
				}
			}
		}
	}
}

// SetFilter sets the filter text and refilters.
func (ft *FileTree) SetFilter(text string) {
	ft.filterText = text
	ft.applyFilter()
	ft.selected = 0
	ft.offset = 0
}

// ensureVisible adjusts offset to keep selected item visible.
func (ft *FileTree) ensureVisible() {
	visibleHeight := max(1, ft.height-4)

	// If selected is above visible area, scroll up
	if ft.selected < ft.offset {
		ft.offset = ft.selected
	}

	// If selected is below visible area, scroll down
	if ft.selected >= ft.offset+visibleHeight {
		ft.offset = ft.selected - visibleHeight + 1
	}
}

// selectFile returns a command that sends the selected file.
func (ft *FileTree) selectFile() tea.Cmd {
	if ft.selected >= len(ft.filtered) {
		return nil
	}
	file := ft.filtered[ft.selected]
	return func() tea.Msg {
		return FileSelectedMsg{File: file}
	}
}

// FileSelectedMsg is sent when a file is selected.
type FileSelectedMsg struct {
	File project.File
}

// SelectedFile returns the currently selected file.
func (ft *FileTree) SelectedFile() *project.File {
	if ft.selected >= 0 && ft.selected < len(ft.filtered) {
		return &ft.filtered[ft.selected]
	}
	return nil
}

// GetFiles returns the filtered list of files.
func (ft *FileTree) GetFiles() []project.File {
	return ft.filtered
}

// GetSelectedIndex returns the index of the currently selected file.
func (ft *FileTree) GetSelectedIndex() int {
	return ft.selected
}

// ShortHelp returns help bindings.
func (ft *FileTree) ShortHelp() []key.Binding {
	return []key.Binding{
		key.NewBinding(
			key.WithKeys("j", "k"),
			key.WithHelp("j/k", "navigate"),
		),
		key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "open"),
		),
	}
}
