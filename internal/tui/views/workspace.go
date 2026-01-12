package views

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"yukti/internal/domain/project"
	"yukti/internal/tui/components"
	"yukti/internal/tui/styles"
)

// WorkspaceState represents the current state of the workspace.
type WorkspaceState int

const (
	WorkspaceStateLoading WorkspaceState = iota
	WorkspaceStateReady
	WorkspaceStateError
)

// WorkspaceView is a split-pane IDE-like view with file tree and code viewer.
type WorkspaceView struct {
	// Project context
	proj    project.Project
	repo    project.Repository
	content *project.Content

	// Components
	fileTree   *components.FileTree
	codeViewer *CodeViewerView
	fuzzy      *components.FuzzyFinder
	help       *components.HelpModal
	spinner    spinner.Model

	// State
	state       WorkspaceState
	errMsg      string
	focusedPane components.Pane

	// Layout
	width      int
	height     int
	splitRatio float64

	// Cache for highlighted content
	highlightCache map[string]string
}

// NewWorkspaceView creates a new workspace view for a project.
func NewWorkspaceView(proj project.Project, repo project.Repository) *WorkspaceView {
	// Create spinner
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(styles.Primary)

	// Create empty file tree (will be populated on load)
	ft := components.NewFileTree(nil)

	// Create fuzzy finder
	fuzzy := components.NewFuzzyFinder()

	// Create help modal
	help := components.NewHelpModal()

	return &WorkspaceView{
		proj:           proj,
		repo:           repo,
		fileTree:       ft,
		fuzzy:          fuzzy,
		help:           help,
		spinner:        s,
		state:          WorkspaceStateLoading,
		focusedPane:    components.LeftPane,
		width:          120,
		height:         40,
		splitRatio:     0.28,
		highlightCache: make(map[string]string),
	}
}

// Title implements tui.View.
func (v *WorkspaceView) Title() string {
	return v.proj.Title
}

// ShortHelp implements tui.View.
func (v *WorkspaceView) ShortHelp() []key.Binding {
	if v.state != WorkspaceStateReady {
		return nil
	}

	bindings := []key.Binding{
		key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "pane"),
		),
		key.NewBinding(
			key.WithKeys("ctrl+p"),
			key.WithHelp("^P", "find"),
		),
		key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
	}

	// Add context-specific bindings
	if v.focusedPane == components.LeftPane {
		bindings = append(bindings, v.fileTree.ShortHelp()...)
	} else if v.codeViewer != nil {
		bindings = append(bindings, v.codeViewer.ShortHelp()...)
	}

	return bindings
}

// Init implements tea.Model.
func (v *WorkspaceView) Init() tea.Cmd {
	return tea.Batch(
		v.spinner.Tick,
		v.loadContent(),
	)
}

// Update implements tea.Model.
func (v *WorkspaceView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle modals first (help and fuzzy finder)
	if handled, cmd := v.handleModals(msg); handled {
		return v, cmd
	}

	// Handle main messages
	return v.handleMainMessages(msg)
}

// handleModals handles help modal and fuzzy finder overlays.
func (v *WorkspaceView) handleModals(msg tea.Msg) (handled bool, cmd tea.Cmd) {
	// Handle help modal first if visible
	if v.help.IsVisible() {
		v.help.Update(msg)
		return true, nil
	}

	// Handle fuzzy finder if visible
	if v.fuzzy.IsVisible() {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			_, cmd = v.fuzzy.Update(msg)
			return true, cmd
		case components.FuzzySelectMsg:
			return true, v.handleFuzzySelect(msg.Item)
		}
		_, cmd = v.fuzzy.Update(msg)
		return true, cmd
	}

	return false, nil
}

// handleMainMessages handles the main workspace messages.
func (v *WorkspaceView) handleMainMessages(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.width = msg.Width
		v.height = msg.Height
		v.updateChildSizes()

	case tea.KeyMsg:
		if cmd := v.handleKeyMsg(msg); cmd != nil {
			return v, cmd
		}

	case spinner.TickMsg:
		if v.state == WorkspaceStateLoading {
			var cmd tea.Cmd
			v.spinner, cmd = v.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}

	case workspaceContentLoadedMsg:
		cmd := v.handleContentLoaded(msg)
		return v, cmd

	case workspaceContentErrorMsg:
		v.state = WorkspaceStateError
		v.errMsg = msg.err.Error()
		return v, nil

	case components.FileSelectedMsg:
		v.selectFile(msg.File)
		return v, nil

	case components.FuzzySelectMsg:
		cmd := v.handleFuzzySelect(msg.Item)
		return v, cmd
	}

	// Update focused component
	if v.state == WorkspaceStateReady {
		cmds = append(cmds, v.updateFocusedComponent(msg))
	}

	return v, tea.Batch(cmds...)
}

// handleContentLoaded handles when project content is loaded.
func (v *WorkspaceView) handleContentLoaded(msg workspaceContentLoadedMsg) tea.Cmd {
	v.state = WorkspaceStateReady
	v.content = msg.content
	v.fileTree.SetFiles(msg.content.Files)
	v.fuzzy.SetItems(msg.content.Files)

	// Auto-select first file
	if len(msg.content.Files) > 0 {
		v.selectFile(msg.content.Files[0])
	}
	return nil
}

// handleKeyMsg handles keyboard input.
func (v *WorkspaceView) handleKeyMsg(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "tab":
		v.toggleFocus()
		return nil

	case "ctrl+p":
		return v.fuzzy.Show()

	case "?":
		v.help.Toggle()
		return nil

	case "r":
		if v.state == WorkspaceStateReady {
			v.state = WorkspaceStateLoading
			return tea.Batch(v.spinner.Tick, v.loadContent())
		}
	}

	// Pass to focused component
	if v.focusedPane == components.LeftPane {
		var cmd tea.Cmd
		_, cmd = v.fileTree.Update(msg)
		return cmd
	} else if v.codeViewer != nil {
		var cmd tea.Cmd
		_, cmd = v.codeViewer.Update(msg)
		return cmd
	}

	return nil
}

// handleFuzzySelect handles selection from fuzzy finder.
func (v *WorkspaceView) handleFuzzySelect(item components.FuzzyItem) tea.Cmd {
	if item.File != nil {
		v.selectFile(*item.File)
		v.focusedPane = components.RightPane

		// If a function was selected, scroll to it
		if item.Function != nil && v.codeViewer != nil {
			// TODO: Implement scroll to line
			_ = item.LineNum
		}
	}
	return nil
}

// toggleFocus switches focus between panes.
func (v *WorkspaceView) toggleFocus() {
	if v.focusedPane == components.LeftPane {
		v.focusedPane = components.RightPane
	} else {
		v.focusedPane = components.LeftPane
	}
}

// selectFile sets the current file in the code viewer.
func (v *WorkspaceView) selectFile(file project.File) {
	v.codeViewer = NewCodeViewerView(file)
	// Send size message to initialize viewport
	v.codeViewer.Update(tea.WindowSizeMsg{
		Width:  v.getRightPaneWidth(),
		Height: v.height - 6,
	})
}

// updateFocusedComponent sends messages to the focused component.
func (v *WorkspaceView) updateFocusedComponent(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	if v.focusedPane == components.LeftPane {
		_, cmd = v.fileTree.Update(msg)
	} else if v.codeViewer != nil {
		_, cmd = v.codeViewer.Update(msg)
	}
	return cmd
}

// updateChildSizes updates component sizes based on current dimensions.
func (v *WorkspaceView) updateChildSizes() {
	leftWidth := v.getLeftPaneWidth()
	rightWidth := v.getRightPaneWidth()
	contentHeight := v.height - 6 // Account for header and footer

	// Update file tree
	v.fileTree.Update(tea.WindowSizeMsg{
		Width:  leftWidth,
		Height: contentHeight,
	})

	// Update code viewer
	if v.codeViewer != nil {
		v.codeViewer.Update(tea.WindowSizeMsg{
			Width:  rightWidth,
			Height: contentHeight,
		})
	}

	// Update fuzzy finder
	v.fuzzy.Update(tea.WindowSizeMsg{
		Width:  v.width,
		Height: v.height,
	})
}

// getLeftPaneWidth returns the width of the left pane.
func (v *WorkspaceView) getLeftPaneWidth() int {
	width := int(float64(v.width) * v.splitRatio)
	return max(20, min(width, v.width-40))
}

// getRightPaneWidth returns the width of the right pane.
func (v *WorkspaceView) getRightPaneWidth() int {
	return v.width - v.getLeftPaneWidth() - 3 // 3 for separator
}

// View implements tea.Model.
func (v *WorkspaceView) View() string {
	switch v.state {
	case WorkspaceStateLoading:
		return v.renderLoading()
	case WorkspaceStateError:
		return v.renderError()
	default:
		return v.renderWorkspace()
	}
}

func (v *WorkspaceView) renderLoading() string {
	loadingStyle := lipgloss.NewStyle().
		Foreground(styles.TextSecondary)

	content := lipgloss.JoinVertical(
		lipgloss.Center,
		v.spinner.View(),
		"",
		loadingStyle.Render("Loading project content..."),
	)

	return lipgloss.Place(
		v.width,
		v.height,
		lipgloss.Center,
		lipgloss.Center,
		content,
	)
}

func (v *WorkspaceView) renderError() string {
	titleStyle := lipgloss.NewStyle().
		Foreground(styles.Error).
		Bold(true)

	messageStyle := lipgloss.NewStyle().
		Foreground(styles.TextSecondary).
		MarginTop(1)

	hintStyle := lipgloss.NewStyle().
		Foreground(styles.TextMuted).
		Italic(true).
		MarginTop(2)

	content := lipgloss.JoinVertical(
		lipgloss.Center,
		titleStyle.Render("Failed to load project"),
		messageStyle.Render(v.errMsg),
		hintStyle.Render("Press 'r' to retry or 'esc' to go back"),
	)

	return lipgloss.Place(
		v.width,
		v.height,
		lipgloss.Center,
		lipgloss.Center,
		content,
	)
}

func (v *WorkspaceView) renderWorkspace() string {
	leftWidth := v.getLeftPaneWidth()
	rightWidth := v.getRightPaneWidth()
	contentHeight := v.height - 6

	// Pane styles based on focus
	leftBorder := styles.Border
	rightBorder := styles.Border
	leftTitleColor := styles.TextMuted
	rightTitleColor := styles.TextMuted
	if v.focusedPane == components.LeftPane {
		leftBorder = styles.Primary
		leftTitleColor = styles.Primary
	} else {
		rightBorder = styles.Primary
		rightTitleColor = styles.Primary
	}

	// Build panel titles with numbering like lazygit: [1]─Files
	leftTitleStyle := lipgloss.NewStyle().Foreground(leftTitleColor).Bold(true)
	rightTitleStyle := lipgloss.NewStyle().Foreground(rightTitleColor).Bold(true)

	// File count info (plain text, styled in buildTitleBorder)
	fileCount := len(v.fileTree.GetFiles())
	selectedIdx := v.fileTree.GetSelectedIndex()
	leftInfo := ""
	if fileCount > 0 {
		leftInfo = fmt.Sprintf("%d of %d", selectedIdx+1, fileCount)
	}

	// Build left panel with custom title border
	leftTitle := leftTitleStyle.Render("[1]─Files")
	leftTopBorder := v.buildTitleBorder(leftTitle, leftInfo, leftWidth-2, leftBorder)

	leftPaneStyle := lipgloss.NewStyle().
		Width(leftWidth - 2).
		Height(contentHeight - 2)

	leftContent := leftPaneStyle.Render(v.fileTree.View())
	leftPane := v.renderPanelWithTitle(leftTopBorder, leftContent, contentHeight, leftWidth-2, leftBorder)

	// Build right panel
	var rightContent string
	var rightTitle string
	if v.codeViewer != nil {
		rightTitle = rightTitleStyle.Render("[2]─" + v.codeViewer.GetFileName())
		rightContent = v.codeViewer.View()
	} else {
		rightTitle = rightTitleStyle.Render("[2]─Code")
		emptyStyle := lipgloss.NewStyle().
			Foreground(styles.TextMuted).
			Italic(true)
		rightContent = lipgloss.Place(
			rightWidth-4,
			contentHeight-4,
			lipgloss.Center,
			lipgloss.Center,
			emptyStyle.Render("Select a file to view"),
		)
	}

	rightTopBorder := v.buildTitleBorder(rightTitle, "", rightWidth-2, rightBorder)
	rightPaneStyle := lipgloss.NewStyle().
		Width(rightWidth - 2).
		Height(contentHeight - 2)

	rightContentStyled := rightPaneStyle.Render(rightContent)
	rightPane := v.renderPanelWithTitle(rightTopBorder, rightContentStyled, contentHeight, rightWidth-2, rightBorder)

	// Join panes
	workspace := lipgloss.JoinHorizontal(lipgloss.Top, leftPane, " ", rightPane)

	// Add help modal overlay if visible
	if v.help.IsVisible() {
		helpView := v.help.View()
		workspace = v.overlayModal(workspace, helpView)
	}

	// Add fuzzy finder overlay if visible
	if v.fuzzy.IsVisible() {
		fuzzyView := v.fuzzy.View()
		workspace = v.overlayModal(workspace, fuzzyView)
	}

	return workspace
}

// buildTitleBorder creates a top border with embedded title and optional info.
func (v *WorkspaceView) buildTitleBorder(title, info string, width int, borderColor lipgloss.Color) string {
	borderStyle := lipgloss.NewStyle().Foreground(borderColor)
	infoStyle := lipgloss.NewStyle().Foreground(styles.TextMuted)

	// Unicode box drawing
	topLeft := "╭"
	topRight := "╮"
	horizontal := "─"

	titleWidth := lipgloss.Width(title)
	infoWidth := lipgloss.Width(info)

	// Total width = 1 (╭) + titleWidth + dashes + infoWidth + 1 (╮)
	// Therefore: dashes = width - 2 - titleWidth - infoWidth
	dashCount := width - 2 - titleWidth - infoWidth
	dashCount = max(0, dashCount)
	dashes := strings.Repeat(horizontal, dashCount)

	// Build border: ╭title────info╮
	border := topLeft + title + dashes
	if info != "" {
		border += infoStyle.Render(info)
	}
	border += topRight

	return borderStyle.Render(border)
}

// renderPanelWithTitle renders a panel with custom title border.
func (v *WorkspaceView) renderPanelWithTitle(topBorder, content string, height, width int, borderColor lipgloss.Color) string {
	borderStyle := lipgloss.NewStyle().Foreground(borderColor)

	vertical := "│"
	bottomLeft := "╰"
	bottomRight := "╯"
	horizontal := "─"

	lines := splitIntoLines(content)

	var result string
	result += topBorder + "\n"

	// Content lines
	contentHeight := height - 2
	for i := range contentHeight {
		var line string
		if i < len(lines) {
			line = lines[i]
		}

		// Ensure line fits
		lineWidth := lipgloss.Width(line)
		if lineWidth < width-2 {
			line += lipgloss.NewStyle().Width(width - 2 - lineWidth).Render("")
		}

		result += borderStyle.Render(vertical) + line + borderStyle.Render(vertical) + "\n"
	}

	// Bottom border
	bottom := bottomLeft
	for range width - 2 {
		bottom += horizontal
	}
	bottom += bottomRight
	result += borderStyle.Render(bottom)

	return result
}

// splitIntoLines splits content into lines.
func splitIntoLines(s string) []string {
	if s == "" {
		return nil
	}
	var lines []string
	start := 0
	for i := range len(s) {
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

// overlayModal overlays a modal on top of the background.
func (v *WorkspaceView) overlayModal(background, modal string) string {
	bgHeight := lipgloss.Height(background)
	modalHeight := lipgloss.Height(modal)
	modalWidth := lipgloss.Width(modal)

	// Calculate center position
	topPadding := (bgHeight - modalHeight) / 3
	leftPadding := (v.width - modalWidth) / 2

	topPadding = max(0, topPadding)
	leftPadding = max(0, leftPadding)

	// Create padded overlay
	paddedModal := lipgloss.NewStyle().
		MarginTop(topPadding).
		MarginLeft(leftPadding).
		Render(modal)

	return lipgloss.Place(
		v.width,
		v.height-6,
		lipgloss.Left,
		lipgloss.Top,
		paddedModal,
	)
}

// loadContent fetches the project content.
func (v *WorkspaceView) loadContent() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		content, err := v.repo.GetContent(ctx, v.proj.ID)
		if err != nil {
			return workspaceContentErrorMsg{err: err}
		}

		return workspaceContentLoadedMsg{content: content}
	}
}

// Messages for workspace view.
type workspaceContentLoadedMsg struct {
	content *project.Content
}

type workspaceContentErrorMsg struct {
	err error
}
