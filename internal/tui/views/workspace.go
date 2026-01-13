package views

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	appprocess "yukti/internal/application/process"
	domainprocess "yukti/internal/domain/process"
	"yukti/internal/domain/project"
	"yukti/internal/infrastructure/logger"
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
	fileTree     *components.FileTree
	codeViewer   *CodeViewerView
	executionLog *components.ExecutionLog
	logModal     *components.LogModal // Full screen log viewer
	fuzzy        *components.FuzzyFinder
	help         *components.HelpModal
	spinner      spinner.Model

	// Process service for running functions
	processService *appprocess.Service

	// State
	state         WorkspaceState
	errMsg        string
	focusedPane   components.Pane
	showRunPicker bool // Show function picker for running
	showLogPath   bool // Show log path info modal
	runningFunc   string

	// Layout
	width      int
	height     int
	splitRatio float64

	// Cache for highlighted content
	highlightCache map[string]string
}

// NewWorkspaceView creates a new workspace view for a project.
func NewWorkspaceView(proj project.Project, repo project.Repository) *WorkspaceView {
	return NewWorkspaceViewWithService(proj, repo, nil)
}

// NewWorkspaceViewWithService creates a workspace view with an optional process service.
func NewWorkspaceViewWithService(proj project.Project, repo project.Repository, processService *appprocess.Service) *WorkspaceView {
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

	// Create execution log
	execLog := components.NewExecutionLog()

	return &WorkspaceView{
		proj:           proj,
		repo:           repo,
		fileTree:       ft,
		executionLog:   execLog,
		fuzzy:          fuzzy,
		help:           help,
		spinner:        s,
		processService: processService,
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

// HasModal returns true if any modal is currently visible.
// Implements tui.ModalHandler to prevent app from intercepting Back key.
func (v *WorkspaceView) HasModal() bool {
	return v.showLogPath || v.help.IsVisible() || v.fuzzy.IsVisible()
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
			key.WithKeys("ctrl+r"),
			key.WithHelp("^R", "run"),
		),
	}

	// Add log toggle hint based on state
	if v.executionLog.IsExpanded() {
		bindings = append(bindings, key.NewBinding(
			key.WithKeys("L"),
			key.WithHelp("L", "hide logs"),
		))
	} else {
		bindings = append(bindings, key.NewBinding(
			key.WithKeys("L"),
			key.WithHelp("L", "logs"),
		))
	}

	bindings = append(bindings,
		key.NewBinding(
			key.WithKeys("ctrl+p"),
			key.WithHelp("^P", "find"),
		),
		key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
	)

	// Add context-specific bindings
	switch v.focusedPane {
	case components.LeftPane:
		bindings = append(bindings, v.fileTree.ShortHelp()...)
	case components.RightPane:
		if v.codeViewer != nil {
			bindings = append(bindings, v.codeViewer.ShortHelp()...)
		}
	case components.BottomPane:
		bindings = append(bindings, v.executionLog.ShortHelp()...)
		// Add log-specific hints when execution log is focused
		bindings = append(bindings,
			key.NewBinding(key.WithKeys("p"), key.WithHelp("p", "path")),
			key.NewBinding(key.WithKeys("O"), key.WithHelp("O", "open dir")),
		)
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
	// Handle log modal first (full screen overlay)
	if v.logModal != nil {
		v.logModal, cmd = v.logModal.Update(msg)
		return true, cmd
	}

	// Handle log path info modal
	if v.showLogPath {
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			switch keyMsg.String() {
			case "O":
				// Open log directory, then close modal
				v.openLogDirectory()
				v.showLogPath = false
			default:
				// Any other key just closes the modal
				v.showLogPath = false
			}
			return true, nil
		}
		return true, nil
	}

	// Handle help modal first if visible
	if v.help.IsVisible() {
		v.help.Update(msg)
		return true, nil
	}

	// Handle run picker if visible
	if v.showRunPicker && v.fuzzy.IsVisible() {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			if msg.String() == "esc" {
				v.showRunPicker = false
				v.fuzzy.Hide()
				return true, nil
			}
			_, cmd = v.fuzzy.Update(msg)
			return true, cmd
		case components.FuzzySelectMsg:
			v.showRunPicker = false
			return true, v.handleRunFunctionSelect(msg.Item)
		}
		_, cmd = v.fuzzy.Update(msg)
		return true, cmd
	}

	// Handle regular fuzzy finder if visible
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
	// Try log-related messages first
	if model, cmd, handled := v.handleLogMessages(msg); handled {
		return model, cmd
	}

	// Try execution-related messages
	if model, cmd, handled := v.handleExecutionMessages(msg); handled {
		return model, cmd
	}

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

	case components.SpinnerTickMsg:
		// Forward to execution log
		_, cmd := v.executionLog.Update(msg)
		cmds = append(cmds, cmd)

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
		if v.showRunPicker {
			v.showRunPicker = false
			cmd := v.handleRunFunctionSelect(msg.Item)
			return v, cmd
		}
		cmd := v.handleFuzzySelect(msg.Item)
		return v, cmd
	}

	// Update focused component
	if v.state == WorkspaceStateReady {
		cmds = append(cmds, v.updateFocusedComponent(msg))
	}

	return v, tea.Batch(cmds...)
}

// handleExecutionMessages handles function execution result messages.
func (v *WorkspaceView) handleExecutionMessages(msg tea.Msg) (tea.Model, tea.Cmd, bool) {
	if msg, ok := msg.(runFunctionResultMsg); ok {
		v.runningFunc = ""
		if msg.isNew {
			// Replace the placeholder entry with the actual result
			v.executionLog.ReplacePlaceholder(msg.entry)
		} else {
			v.executionLog.UpdateEntry(msg.entry)
		}
		return v, nil, true
	}
	return nil, nil, false
}

// handleLogMessages handles log-related messages.
func (v *WorkspaceView) handleLogMessages(msg tea.Msg) (tea.Model, tea.Cmd, bool) {
	switch msg := msg.(type) {
	case components.FetchLogsMsg:
		cmd := v.fetchLogsCmd(msg.EntryID, msg.ScriptID)
		return v, cmd, true

	case components.LogsFetchedMsg:
		v.executionLog.Update(msg)
		return v, nil, true

	case components.OpenLogModalMsg:
		v.logModal = components.NewLogModal(msg.Entry)
		v.logModal.SetSize(v.width, v.height)
		return v, nil, true

	case components.CloseLogModalMsg:
		v.logModal = nil
		return v, nil, true
	}
	return nil, nil, false
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
//
//nolint:gocyclo // Key handlers are naturally complex
func (v *WorkspaceView) handleKeyMsg(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "tab":
		v.cycleFocus()
		return nil

	case "ctrl+p":
		v.showRunPicker = false
		return v.fuzzy.Show()

	case "ctrl+r", "x":
		// Open function picker for running
		if v.processService != nil && v.content != nil {
			v.showRunPicker = true
			v.fuzzy.SetFunctionsOnly(true)
			v.fuzzy.SetTitle("Run Function")
			return v.fuzzy.Show()
		}

	case "L":
		v.executionLog.Toggle()
		v.updateChildSizes()
		return nil

	case "?":
		v.help.Toggle()
		return nil

	case "r":
		if v.state == WorkspaceStateReady && !v.showRunPicker {
			v.state = WorkspaceStateLoading
			return tea.Batch(v.spinner.Tick, v.loadContent())
		}
	}

	// Pass to focused component
	switch v.focusedPane {
	case components.LeftPane:
		var cmd tea.Cmd
		_, cmd = v.fileTree.Update(msg)
		return cmd
	case components.RightPane:
		if v.codeViewer != nil {
			var cmd tea.Cmd
			_, cmd = v.codeViewer.Update(msg)
			return cmd
		}
	case components.BottomPane:
		// Handle log-specific keybindings
		switch msg.String() {
		case "O":
			v.openLogDirectory()
			return nil
		case "p":
			v.showLogPath = true
			return nil
		}
		_, cmd := v.executionLog.Update(msg)
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

// cycleFocus cycles focus through panes (Left -> Right -> Bottom -> Left).
func (v *WorkspaceView) cycleFocus() {
	switch v.focusedPane {
	case components.LeftPane:
		v.focusedPane = components.RightPane
	case components.RightPane:
		if v.executionLog.IsExpanded() {
			v.focusedPane = components.BottomPane
		} else {
			v.focusedPane = components.LeftPane
		}
	case components.BottomPane:
		v.focusedPane = components.LeftPane
	}
	v.updateFocusState()
}

// updateFocusState updates component focus states.
func (v *WorkspaceView) updateFocusState() {
	v.executionLog.SetFocused(v.focusedPane == components.BottomPane)
}

// selectFile sets the current file in the code viewer.
func (v *WorkspaceView) selectFile(file project.File) {
	v.codeViewer = NewCodeViewerView(file)

	// Calculate content height - reduce if execution log is expanded
	logHeight := 0
	if v.executionLog.IsExpanded() {
		logHeight = 8
	}
	contentHeight := v.height - 6 - logHeight // Account for header, footer, and log panel

	// Send size message to initialize viewport
	v.codeViewer.Update(tea.WindowSizeMsg{
		Width:  v.getRightPaneWidth(),
		Height: contentHeight,
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

	// Calculate content heights - reduce if execution log is expanded
	logHeight := 0
	if v.executionLog.IsExpanded() {
		logHeight = 8 // Fixed height for execution log panel
	}
	contentHeight := v.height - 6 - logHeight // Account for header, footer, and log panel

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

	// Update execution log size - must match combined width of top panels
	// Top panels width = (leftWidth-2) + 1 (space) + (rightWidth-2)
	if v.executionLog.IsExpanded() {
		logWidth := (leftWidth - 2) + 1 + (rightWidth - 2)
		v.executionLog.SetSize(logWidth, logHeight)
	}
}

// getLeftPaneWidth returns the width of the left pane.
func (v *WorkspaceView) getLeftPaneWidth() int {
	width := int(float64(v.width) * v.splitRatio)
	return max(20, min(width, v.width-40))
}

// getRightPaneWidth returns the width of the right pane.
func (v *WorkspaceView) getRightPaneWidth() int {
	// Subtract extra 1 to ensure right border doesn't get cut off
	return v.width - v.getLeftPaneWidth() - 4
}

// View implements tea.Model.
func (v *WorkspaceView) View() string {
	// Log modal is full-screen overlay
	if v.logModal != nil {
		return v.logModal.View()
	}

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

	// Calculate content heights - reduce if execution log is expanded
	logHeight := 0
	if v.executionLog.IsExpanded() {
		logHeight = 8 // Fixed height for execution log panel
	}
	contentHeight := v.height - 6 - logHeight

	// Pane styles based on focus
	leftBorder := styles.Border
	rightBorder := styles.Border
	leftTitleColor := styles.TextMuted
	rightTitleColor := styles.TextMuted
	switch v.focusedPane {
	case components.LeftPane:
		leftBorder = styles.Primary
		leftTitleColor = styles.Primary
	case components.RightPane:
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
		Width(leftWidth - 4). // -4 to leave room for vertical borders on each side
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
		Width(rightWidth - 4). // -4 to leave room for vertical borders on each side
		Height(contentHeight - 2)

	rightContentStyled := rightPaneStyle.Render(rightContent)
	rightPane := v.renderPanelWithTitle(rightTopBorder, rightContentStyled, contentHeight, rightWidth-2, rightBorder)

	// Join top panes horizontally
	topPanes := lipgloss.JoinHorizontal(lipgloss.Top, leftPane, " ", rightPane)

	// Build the workspace layout
	var workspace string
	if v.executionLog.IsExpanded() {
		// Update execution log size - must match combined width of top panels
		logWidth := (leftWidth - 2) + 1 + (rightWidth - 2)
		v.executionLog.SetSize(logWidth, logHeight)

		// Get execution log view
		logView := v.executionLog.View()

		// Stack top panes and execution log vertically
		workspace = lipgloss.JoinVertical(lipgloss.Left, topPanes, logView)
	} else {
		workspace = topPanes
	}

	// Add log path modal overlay if visible
	if v.showLogPath {
		logPathView := v.renderLogPathModal()
		workspace = v.overlayModal(workspace, logPathView)
	}

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
// IMPORTANT: Styles each segment separately to avoid nested ANSI code issues.
func (v *WorkspaceView) buildTitleBorder(title, info string, width int, borderColor lipgloss.Color) string {
	borderStyle := lipgloss.NewStyle().Foreground(borderColor)
	infoStyle := lipgloss.NewStyle().Foreground(styles.TextMuted)

	// Unicode box drawing - style each border character separately
	topLeft := borderStyle.Render("╭")
	topRight := borderStyle.Render("╮")
	horizontal := "─"

	titleWidth := lipgloss.Width(title)
	infoWidth := lipgloss.Width(info)

	// Total width = 1 (╭) + titleWidth + dashes + infoWidth + 1 (╮)
	// Therefore: dashes = width - 2 - titleWidth - infoWidth
	dashCount := width - 2 - titleWidth - infoWidth
	dashCount = max(0, dashCount)
	dashes := borderStyle.Render(strings.Repeat(horizontal, dashCount))

	// Build border by concatenating pre-styled segments (no nesting)
	// title is already styled by caller
	var result strings.Builder
	result.WriteString(topLeft)
	result.WriteString(title)
	result.WriteString(dashes)
	if info != "" {
		result.WriteString(infoStyle.Render(info))
	}
	result.WriteString(topRight)

	return result.String()
}

// renderPanelWithTitle renders a panel with custom title border.
// IMPORTANT: Styles each border element separately to avoid nested ANSI code issues.
func (v *WorkspaceView) renderPanelWithTitle(topBorder, content string, height, width int, borderColor lipgloss.Color) string {
	borderStyle := lipgloss.NewStyle().Foreground(borderColor)

	// Pre-style border elements
	verticalBorder := borderStyle.Render("│")
	bottomLeft := borderStyle.Render("╰")
	bottomRight := borderStyle.Render("╯")
	horizontal := "─"

	lines := splitIntoLines(content)

	var result strings.Builder
	result.WriteString(topBorder)
	result.WriteString("\033[0m\n") // Reset after top border

	// Content lines
	contentHeight := height - 2
	for i := range contentHeight {
		var line string
		if i < len(lines) {
			line = lines[i]
		}

		// Ensure line fits the exact width needed
		// Use plain spaces for padding to avoid ANSI background issues
		lineWidth := lipgloss.Width(line)
		padding := ""
		if lineWidth < width-2 {
			padding = strings.Repeat(" ", width-2-lineWidth)
		}

		// Add ANSI reset before border to ensure clean state
		result.WriteString("\033[0m")
		result.WriteString(verticalBorder)
		result.WriteString(line)
		result.WriteString(padding)
		result.WriteString("\033[0m")
		result.WriteString(verticalBorder)
		result.WriteString("\n")
	}

	// Bottom border - style separately
	result.WriteString(bottomLeft)
	result.WriteString(borderStyle.Render(strings.Repeat(horizontal, width-2)))
	result.WriteString(bottomRight)

	return result.String()
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

// overlayModal places a modal centered over the background content.
// The background remains visible around the modal.
func (v *WorkspaceView) overlayModal(background, modal string) string {
	bgLines := strings.Split(background, "\n")
	modalLines := strings.Split(modal, "\n")

	bgHeight := len(bgLines)
	modalHeight := len(modalLines)
	modalWidth := lipgloss.Width(modal)

	// Calculate center position (1/3 from top for visual balance)
	topOffset := max(0, (bgHeight-modalHeight)/3)
	leftOffset := max(0, (v.width-modalWidth)/2)

	// Composite: overlay modal lines onto background
	result := make([]string, len(bgLines))
	for i, bgLine := range bgLines {
		if i >= topOffset && i < topOffset+modalHeight {
			modalLineIdx := i - topOffset
			if modalLineIdx < len(modalLines) {
				result[i] = composeModalLine(bgLine, modalLines[modalLineIdx], leftOffset, modalWidth, v.width)
			} else {
				result[i] = bgLine
			}
		} else {
			result[i] = bgLine
		}
	}

	return strings.Join(result, "\n")
}

// composeModalLine overlays a modal line onto a background line.
// Background is visible on sides; modal replaces the center portion.
func composeModalLine(bgLine, modalLine string, leftOffset, modalWidth, totalWidth int) string {
	// Use ansi.Cut to extract background content while preserving ANSI codes
	// Left portion: characters 0 to leftOffset
	leftPart := ansi.Cut(bgLine, 0, leftOffset)

	// Right portion: characters from leftOffset+modalWidth to end
	rightStart := leftOffset + modalWidth
	rightPart := ansi.Cut(bgLine, rightStart, totalWidth)

	// Compose: left bg + reset + modal + reset + right bg
	return leftPart + "\033[0m" + modalLine + "\033[0m" + rightPart
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

// runFunctionResultMsg is sent when a function execution completes.
type runFunctionResultMsg struct {
	entry appprocess.ExecutionEntry
	isNew bool // If true, replace placeholder; otherwise update existing
}

// handleRunFunctionSelect handles selection from the run function picker.
func (v *WorkspaceView) handleRunFunctionSelect(item components.FuzzyItem) tea.Cmd {
	if item.Function == nil || v.processService == nil {
		return nil
	}

	funcName := item.Function.Name
	v.runningFunc = funcName

	// Run the function - the process service creates and tracks the entry
	runCmd := func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		result, err := v.processService.RunFunction(ctx, v.proj.ID, funcName)
		if err != nil {
			// Create failed entry for display
			return runFunctionResultMsg{
				entry: appprocess.ExecutionEntry{
					ID:           fmt.Sprintf("%d", time.Now().UnixNano()),
					FunctionName: funcName,
					Status:       domainprocess.StatusFailed,
					StartTime:    time.Now(),
					Error:        err.Error(),
					ScriptID:     v.proj.ID,
				},
				isNew: true,
			}
		}
		return runFunctionResultMsg{entry: *result, isNew: true}
	}

	// Show "Running..." immediately
	placeholder := appprocess.ExecutionEntry{
		ID:           "placeholder",
		FunctionName: funcName,
		Status:       domainprocess.StatusRunning,
		StartTime:    time.Now(),
		ScriptID:     v.proj.ID,
	}
	v.executionLog.AddEntry(placeholder)
	v.executionLog.SetExpanded(true)

	// Start spinner animation
	spinnerCmd := v.executionLog.StartSpinner()

	return tea.Batch(spinnerCmd, runCmd)
}

// fetchLogsCmd creates a command to fetch logs for an execution entry.
func (v *WorkspaceView) fetchLogsCmd(entryID, scriptID string) tea.Cmd {
	if v.processService == nil {
		return func() tea.Msg {
			return components.LogsFetchedMsg{
				EntryID: entryID,
				Error:   "Process service not available",
			}
		}
	}

	return func() tea.Msg {
		entry := v.processService.GetEntryByID(scriptID, entryID)
		if entry == nil {
			return components.LogsFetchedMsg{
				EntryID: entryID,
				Error:   "Entry not found",
			}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Fetch logs (10 for preview)
		err := v.processService.FetchLogsForEntry(ctx, entry, 10)
		if err != nil {
			return components.LogsFetchedMsg{
				EntryID: entryID,
				Error:   err.Error(),
			}
		}

		return components.LogsFetchedMsg{
			EntryID: entryID,
			Logs:    entry.Logs,
			Error:   entry.LogsError,
		}
	}
}

// openLogDirectory opens the log directory in the system file manager.
func (v *WorkspaceView) openLogDirectory() {
	logPath := logger.Path()
	// Extract directory from log file path
	dir := logPath
	if idx := strings.LastIndex(logPath, "/"); idx > 0 {
		dir = logPath[:idx]
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", dir) //nolint:gosec // Opening known log directory
	case "linux":
		cmd = exec.Command("xdg-open", dir) //nolint:gosec // Opening known log directory
	case "windows":
		cmd = exec.Command("explorer", dir) //nolint:gosec // Opening known log directory
	default:
		return
	}

	_ = cmd.Start()
}

// renderLogPathModal renders a simple modal showing the full log path.
func (v *WorkspaceView) renderLogPathModal() string {
	logPath := logger.Path()

	// Calculate modal width based on content (path is usually longest)
	modalWidth := max(len(logPath)+6, 50) // +6 for padding

	// Styles - no Background on inner elements (container provides Surface)
	titleStyle := lipgloss.NewStyle().
		Foreground(styles.Primary).
		Bold(true).
		MarginBottom(1)

	pathStyle := lipgloss.NewStyle().
		Foreground(styles.Info).
		Bold(true)

	hintStyle := lipgloss.NewStyle().
		Foreground(styles.TextMuted).
		Italic(true).
		MarginTop(1)

	// No Background() - modal border provides visual separation
	// Background causes bleed when composited via overlayModal
	modalStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Primary).
		Padding(1, 2).
		Width(modalWidth)

	// Build content
	var content strings.Builder
	content.WriteString(titleStyle.Render("📁 Log File Path"))
	content.WriteString("\n\n")
	content.WriteString(pathStyle.Render(logPath))
	content.WriteString("\n")
	content.WriteString(hintStyle.Render("Press any key to close • O to open directory"))

	return modalStyle.Render(content.String())
}
