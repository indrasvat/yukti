package views

import (
	"context"
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

	return &WorkspaceView{
		proj:           proj,
		repo:           repo,
		fileTree:       ft,
		fuzzy:          fuzzy,
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
	var cmds []tea.Cmd

	// Handle fuzzy finder first if visible
	if v.fuzzy.IsVisible() {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			var cmd tea.Cmd
			_, cmd = v.fuzzy.Update(msg)
			cmds = append(cmds, cmd)
			return v, tea.Batch(cmds...)

		case components.FuzzySelectMsg:
			// Handle selection
			cmd := v.handleFuzzySelect(msg.Item)
			return v, cmd
		}

		// Pass other messages to fuzzy
		_, cmd := v.fuzzy.Update(msg)
		cmds = append(cmds, cmd)
		return v, tea.Batch(cmds...)
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.width = msg.Width
		v.height = msg.Height
		v.updateChildSizes()

	case tea.KeyMsg:
		cmd := v.handleKeyMsg(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		return v, tea.Batch(cmds...)

	case spinner.TickMsg:
		if v.state == WorkspaceStateLoading {
			var cmd tea.Cmd
			v.spinner, cmd = v.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}

	case workspaceContentLoadedMsg:
		v.state = WorkspaceStateReady
		v.content = msg.content
		v.fileTree.SetFiles(msg.content.Files)
		v.fuzzy.SetItems(msg.content.Files)

		// Auto-select first file
		if len(msg.content.Files) > 0 {
			v.selectFile(msg.content.Files[0])
		}
		return v, nil

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

// handleKeyMsg handles keyboard input.
func (v *WorkspaceView) handleKeyMsg(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "tab":
		v.toggleFocus()
		return nil

	case "ctrl+p":
		return v.fuzzy.Show()

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
	if v.focusedPane == components.LeftPane {
		leftBorder = styles.Primary
	} else {
		rightBorder = styles.Primary
	}

	leftPaneStyle := lipgloss.NewStyle().
		Width(leftWidth - 2).
		Height(contentHeight).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(leftBorder)

	rightPaneStyle := lipgloss.NewStyle().
		Width(rightWidth - 2).
		Height(contentHeight).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(rightBorder)

	// Render panes
	leftPane := leftPaneStyle.Render(v.fileTree.View())

	var rightPane string
	if v.codeViewer != nil {
		rightPane = rightPaneStyle.Render(v.codeViewer.View())
	} else {
		emptyStyle := lipgloss.NewStyle().
			Foreground(styles.TextMuted).
			Italic(true)
		rightPane = rightPaneStyle.Render(
			lipgloss.Place(
				rightWidth-4,
				contentHeight-2,
				lipgloss.Center,
				lipgloss.Center,
				emptyStyle.Render("Select a file to view"),
			),
		)
	}

	// Join panes
	workspace := lipgloss.JoinHorizontal(lipgloss.Top, leftPane, " ", rightPane)

	// Add fuzzy finder overlay if visible
	if v.fuzzy.IsVisible() {
		fuzzyView := v.fuzzy.View()
		// Center the fuzzy finder
		workspace = v.overlayFuzzy(workspace, fuzzyView)
	}

	return workspace
}

// overlayFuzzy overlays the fuzzy finder on top of the workspace.
func (v *WorkspaceView) overlayFuzzy(background, overlay string) string {
	bgLines := lipgloss.Height(background)
	overlayLines := lipgloss.Height(overlay)
	overlayWidth := lipgloss.Width(overlay)

	// Calculate center position
	topPadding := (bgLines - overlayLines) / 3 // Slight bias toward top
	leftPadding := (v.width - overlayWidth) / 2

	if topPadding < 0 {
		topPadding = 0
	}
	if leftPadding < 0 {
		leftPadding = 0
	}

	// Create padded overlay
	paddedOverlay := lipgloss.NewStyle().
		MarginTop(topPadding).
		MarginLeft(leftPadding).
		Render(overlay)

	// For simplicity, just return the padded overlay on top
	// A more sophisticated approach would composite the strings
	return lipgloss.Place(
		v.width,
		v.height-6,
		lipgloss.Left,
		lipgloss.Top,
		paddedOverlay,
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
