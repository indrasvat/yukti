// Package components provides reusable TUI components.
package components

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"yukti/internal/tui/styles"
)

// Pane represents which pane has focus.
type Pane int

const (
	LeftPane Pane = iota
	RightPane
	BottomPane
)

// SplitPane is a horizontal split pane container with focus management.
type SplitPane struct {
	// Child models
	Left  tea.Model
	Right tea.Model

	// Layout
	width      int
	height     int
	splitRatio float64 // 0.0-1.0, percentage for left pane
	minLeft    int     // Minimum width for left pane
	minRight   int     // Minimum width for right pane

	// Focus
	focused Pane

	// Styling
	showBorder     bool
	focusedStyle   lipgloss.Style
	unfocusedStyle lipgloss.Style
}

// SplitPaneOption configures a SplitPane.
type SplitPaneOption func(*SplitPane)

// NewSplitPane creates a new split pane with the given child models.
func NewSplitPane(left, right tea.Model, opts ...SplitPaneOption) *SplitPane {
	sp := &SplitPane{
		Left:       left,
		Right:      right,
		splitRatio: 0.30, // Default: 30% left, 70% right
		minLeft:    20,
		minRight:   40,
		focused:    LeftPane,
		showBorder: true,
		focusedStyle: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(styles.Primary),
		unfocusedStyle: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(styles.Border),
	}

	for _, opt := range opts {
		opt(sp)
	}

	return sp
}

// WithSplitRatio sets the split ratio (0.0-1.0).
func WithSplitRatio(ratio float64) SplitPaneOption {
	return func(sp *SplitPane) {
		if ratio >= 0.1 && ratio <= 0.9 {
			sp.splitRatio = ratio
		}
	}
}

// WithMinWidths sets minimum widths for both panes.
func WithMinWidths(left, right int) SplitPaneOption {
	return func(sp *SplitPane) {
		sp.minLeft = left
		sp.minRight = right
	}
}

// WithFocus sets the initial focus.
func WithFocus(pane Pane) SplitPaneOption {
	return func(sp *SplitPane) {
		sp.focused = pane
	}
}

// Init implements tea.Model.
func (sp *SplitPane) Init() tea.Cmd {
	return tea.Batch(
		sp.Left.Init(),
		sp.Right.Init(),
	)
}

// Update implements tea.Model.
func (sp *SplitPane) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		sp.width = msg.Width
		sp.height = msg.Height
		// Propagate size to children with calculated dimensions
		leftWidth, rightWidth := sp.calculateWidths()
		cmds = append(cmds, sp.updateChildSizes(leftWidth, rightWidth))

	case tea.KeyMsg:
		// Handle pane switching
		switch msg.String() {
		case "tab", "ctrl+l":
			sp.focused = RightPane
			return sp, nil
		case "shift+tab", "ctrl+h":
			sp.focused = LeftPane
			return sp, nil
		}
	}

	// Route messages to focused pane only for key messages
	if _, isKey := msg.(tea.KeyMsg); isKey {
		if sp.focused == LeftPane {
			var cmd tea.Cmd
			sp.Left, cmd = sp.Left.Update(msg)
			cmds = append(cmds, cmd)
		} else {
			var cmd tea.Cmd
			sp.Right, cmd = sp.Right.Update(msg)
			cmds = append(cmds, cmd)
		}
	} else {
		// Non-key messages go to both panes
		var cmd tea.Cmd
		sp.Left, cmd = sp.Left.Update(msg)
		cmds = append(cmds, cmd)
		sp.Right, cmd = sp.Right.Update(msg)
		cmds = append(cmds, cmd)
	}

	return sp, tea.Batch(cmds...)
}

// View implements tea.Model.
func (sp *SplitPane) View() string {
	leftWidth, rightWidth := sp.calculateWidths()

	// Account for borders (2 chars each side)
	contentHeight := sp.height
	if sp.showBorder {
		contentHeight -= 2
	}

	// Style panes based on focus
	leftStyle := sp.unfocusedStyle
	rightStyle := sp.unfocusedStyle
	if sp.focused == LeftPane {
		leftStyle = sp.focusedStyle
	} else {
		rightStyle = sp.focusedStyle
	}

	// Render panes with borders
	leftContent := leftStyle.
		Width(leftWidth - 2). // Account for border
		Height(contentHeight).
		Render(sp.Left.View())

	rightContent := rightStyle.
		Width(rightWidth - 2). // Account for border
		Height(contentHeight).
		Render(sp.Right.View())

	return lipgloss.JoinHorizontal(lipgloss.Top, leftContent, rightContent)
}

// calculateWidths returns the widths for left and right panes.
func (sp *SplitPane) calculateWidths() (leftWidth, rightWidth int) {
	// Calculate based on ratio
	leftWidth = int(float64(sp.width) * sp.splitRatio)
	rightWidth = sp.width - leftWidth

	// Enforce minimums
	if leftWidth < sp.minLeft {
		leftWidth = sp.minLeft
		rightWidth = sp.width - leftWidth
	}
	if rightWidth < sp.minRight {
		rightWidth = sp.minRight
		leftWidth = sp.width - rightWidth
	}

	// Safety bounds
	if leftWidth < 10 {
		leftWidth = 10
	}
	if rightWidth < 10 {
		rightWidth = 10
	}

	return leftWidth, rightWidth
}

// updateChildSizes sends WindowSizeMsg to both children.
func (sp *SplitPane) updateChildSizes(leftWidth, rightWidth int) tea.Cmd {
	// Account for borders
	contentHeight := sp.height - 2

	return func() tea.Msg {
		return childSizeMsg{
			leftWidth:   leftWidth - 2,
			rightWidth:  rightWidth - 2,
			height:      contentHeight,
			leftHeight:  contentHeight,
			rightHeight: contentHeight,
		}
	}
}

// childSizeMsg is sent to notify children of their sizes.
type childSizeMsg struct {
	leftWidth   int
	rightWidth  int
	height      int
	leftHeight  int
	rightHeight int
}

// Focused returns which pane is currently focused.
func (sp *SplitPane) Focused() Pane {
	return sp.focused
}

// SetFocus sets the focused pane.
func (sp *SplitPane) SetFocus(pane Pane) {
	sp.focused = pane
}

// ToggleFocus switches focus between panes.
func (sp *SplitPane) ToggleFocus() {
	if sp.focused == LeftPane {
		sp.focused = RightPane
	} else {
		sp.focused = LeftPane
	}
}

// ShortHelp returns help keybindings for the split pane.
func (sp *SplitPane) ShortHelp() []key.Binding {
	return []key.Binding{
		key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "switch pane"),
		),
	}
}
