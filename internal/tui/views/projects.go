package views

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"yukti/internal/domain/project"
	"yukti/internal/tui"
	"yukti/internal/tui/components"
	"yukti/internal/tui/styles"
)

// ProjectListState represents the current state of the project list view.
type ProjectListState int

const (
	ProjectListStateLoading ProjectListState = iota
	ProjectListStateReady
	ProjectListStateError
)

// ProjectsView displays the list of Google Apps Script projects.
type ProjectsView struct {
	repo     project.Repository
	projects []project.Project
	filtered []project.Project // Filtered list (same as projects when no filter)
	spinner  spinner.Model
	state    ProjectListState
	errMsg   string
	selected int
	offset   int
	width    int
	height   int

	// Filter state
	filtering   bool
	filterInput textinput.Model

	// Help modal
	help *components.HelpModal
}

// NewProjectsView creates a new projects list view.
func NewProjectsView(repo project.Repository) *ProjectsView {
	// Create spinner
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(styles.Primary)

	// Create filter input
	ti := textinput.New()
	ti.Placeholder = "Type to filter..."
	ti.CharLimit = 50
	ti.Width = 30
	ti.PromptStyle = lipgloss.NewStyle().Foreground(styles.Primary)
	ti.TextStyle = lipgloss.NewStyle().Foreground(styles.TextPrimary)

	return &ProjectsView{
		repo:        repo,
		spinner:     s,
		filterInput: ti,
		state:       ProjectListStateLoading,
		selected:    0,
		offset:      0,
		width:       80,
		height:      24,
		help:        components.NewHelpModal(),
	}
}

// Title implements tui.View.
func (v *ProjectsView) Title() string {
	return "Projects"
}

// HasModal returns true if any modal is currently visible.
// Implements tui.ModalHandler to prevent app from intercepting Back key.
func (v *ProjectsView) HasModal() bool {
	return v.help.IsVisible() || v.filtering
}

// ShortHelp implements tui.View.
func (v *ProjectsView) ShortHelp() []key.Binding {
	if v.state != ProjectListStateReady {
		return nil
	}

	return []key.Binding{
		key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "open"),
		),
		key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "filter"),
		),
		key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "new"),
		),
		key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "refresh"),
		),
	}
}

// Init implements tea.Model.
func (v *ProjectsView) Init() tea.Cmd {
	return tea.Batch(
		v.spinner.Tick,
		v.loadProjects(),
	)
}

// Update implements tea.Model.
func (v *ProjectsView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle help modal first if visible
	if v.help.IsVisible() {
		v.help.Update(msg)
		return v, nil
	}

	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.width = msg.Width
		v.height = msg.Height
		return v, nil

	case tea.KeyMsg:
		// Handle `?` for help on any state
		if msg.String() == "?" {
			v.help.Toggle()
			return v, nil
		}
		if v.state == ProjectListStateReady {
			if cmd, handled := v.handleKeyMsg(msg); handled {
				return v, cmd
			}
		}

	case spinner.TickMsg:
		if v.state == ProjectListStateLoading {
			var cmd tea.Cmd
			v.spinner, cmd = v.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}

	case projectsLoadedMsg:
		v.state = ProjectListStateReady
		v.projects = msg.projects
		return v, nil

	case projectsErrorMsg:
		v.state = ProjectListStateError
		v.errMsg = msg.err.Error()
		return v, nil
	}

	return v, tea.Batch(cmds...)
}

// handleKeyMsg handles keyboard input for the projects list.
func (v *ProjectsView) handleKeyMsg(msg tea.KeyMsg) (tea.Cmd, bool) {
	// Handle filter mode
	if v.filtering {
		return v.handleFilterKeyMsg(msg)
	}

	// Get the active list (filtered or all projects)
	activeList := v.getActiveList()

	switch msg.String() {
	case "j", "down":
		if v.selected < len(activeList)-1 {
			v.selected++
			v.ensureVisible()
		}
		return nil, true
	case "k", "up":
		if v.selected > 0 {
			v.selected--
			v.ensureVisible()
		}
		return nil, true
	case "g":
		v.selected = 0
		v.offset = 0
		return nil, true
	case "G":
		v.selected = len(activeList) - 1
		v.ensureVisible()
		return nil, true
	case "r":
		v.state = ProjectListStateLoading
		v.clearFilter()
		return tea.Batch(v.spinner.Tick, v.loadProjects()), true
	case "enter":
		if v.selected < len(activeList) {
			return v.openProject(activeList[v.selected]), true
		}
	case "/":
		v.filtering = true
		v.filterInput.Focus()
		return textinput.Blink, true
	case "esc", "escape":
		if len(v.filtered) > 0 {
			v.clearFilter()
			return nil, true
		}
	}
	return nil, false
}

// handleFilterKeyMsg handles keyboard input when in filter mode.
func (v *ProjectsView) handleFilterKeyMsg(msg tea.KeyMsg) (tea.Cmd, bool) {
	switch msg.String() {
	case "esc", "escape":
		v.filtering = false
		v.filterInput.Blur()
		return nil, true
	case "enter":
		v.filtering = false
		v.filterInput.Blur()
		return nil, true
	}

	// Update the filter input
	var cmd tea.Cmd
	v.filterInput, cmd = v.filterInput.Update(msg)

	// Apply filter
	v.applyFilter()

	return cmd, true
}

// getActiveList returns the filtered list if a filter is active, otherwise all projects.
func (v *ProjectsView) getActiveList() []project.Project {
	if len(v.filtered) > 0 || v.filterInput.Value() != "" {
		return v.filtered
	}
	return v.projects
}

// applyFilter filters the projects based on the current filter input.
func (v *ProjectsView) applyFilter() {
	query := strings.ToLower(v.filterInput.Value())
	if query == "" {
		v.filtered = nil
		return
	}

	v.filtered = make([]project.Project, 0)
	for i := range v.projects {
		if strings.Contains(strings.ToLower(v.projects[i].Title), query) {
			v.filtered = append(v.filtered, v.projects[i])
		}
	}

	// Reset selection if out of bounds
	if v.selected >= len(v.filtered) {
		v.selected = max(0, len(v.filtered)-1)
	}
	v.offset = 0
}

// clearFilter clears the current filter.
func (v *ProjectsView) clearFilter() {
	v.filterInput.SetValue("")
	v.filtered = nil
	v.selected = 0
	v.offset = 0
}

// ensureVisible adjusts offset to keep selected item visible.
func (v *ProjectsView) ensureVisible() {
	activeList := v.getActiveList()
	cardHeight := 6 // Height of each project card including spacing
	visibleCards := max((v.height-10)/cardHeight, 1)

	if v.selected < v.offset {
		v.offset = v.selected
	}
	if v.selected >= v.offset+visibleCards {
		v.offset = v.selected - visibleCards + 1
	}
	// Clamp to valid range
	if v.selected >= len(activeList) {
		v.selected = max(0, len(activeList)-1)
	}
}

// View implements tea.Model.
func (v *ProjectsView) View() string {
	var view string
	switch v.state {
	case ProjectListStateLoading:
		view = v.renderLoading()
	case ProjectListStateError:
		view = v.renderError()
	default:
		view = v.renderList()
	}

	// Overlay help modal if visible
	if v.help.IsVisible() {
		view = v.overlayModal(view, v.help.View())
	}

	return view
}

// overlayModal composites a modal onto styled background content.
func (v *ProjectsView) overlayModal(background, modal string) string {
	bgLines := strings.Split(background, "\n")
	modalLines := strings.Split(modal, "\n")

	bgHeight := len(bgLines)
	modalHeight := len(modalLines)
	modalWidth := lipgloss.Width(modal)

	// Calculate center position
	topOffset := max(0, (bgHeight-modalHeight)/3)
	leftOffset := max(0, (v.width-modalWidth)/2)

	// Composite: overlay modal lines onto background
	result := make([]string, len(bgLines))
	for i, bgLine := range bgLines {
		if i >= topOffset && i < topOffset+modalHeight {
			modalLineIdx := i - topOffset
			result[i] = composeProjectsModalLine(bgLine, modalLines[modalLineIdx], leftOffset, modalWidth, v.width)
		} else {
			result[i] = bgLine
		}
	}
	return strings.Join(result, "\n")
}

// composeProjectsModalLine overlays a modal line onto a background line.
func composeProjectsModalLine(bgLine, modalLine string, leftOffset, modalWidth, totalWidth int) string {
	leftPart := ansi.Cut(bgLine, 0, leftOffset)
	rightPart := ansi.Cut(bgLine, leftOffset+modalWidth, totalWidth)
	return leftPart + "\033[0m" + modalLine + "\033[0m" + rightPart
}

func (v *ProjectsView) renderLoading() string {
	loadingStyle := lipgloss.NewStyle().
		Foreground(styles.TextSecondary)

	content := lipgloss.JoinVertical(
		lipgloss.Center,
		v.spinner.View(),
		"",
		loadingStyle.Render("Loading your projects..."),
	)

	// Center content - View() will handle the fixed-size canvas
	return lipgloss.Place(
		v.width,
		v.height,
		lipgloss.Center,
		lipgloss.Center,
		content,
	)
}

func (v *ProjectsView) renderError() string {
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
		titleStyle.Render("✗ Failed to load projects"),
		messageStyle.Render(v.errMsg),
		hintStyle.Render("Press 'r' to retry or 'q' to quit"),
	)

	return lipgloss.Place(
		v.width,
		v.height,
		lipgloss.Center,
		lipgloss.Center,
		content,
	)
}

func (v *ProjectsView) renderList() string {
	activeList := v.getActiveList()

	if len(v.projects) == 0 {
		return v.renderEmpty()
	}

	// Header styles
	headerStyle := lipgloss.NewStyle().
		Foreground(styles.TextPrimary).
		Bold(true).
		MarginBottom(1)

	subHeaderStyle := lipgloss.NewStyle().
		Foreground(styles.TextMuted)

	searchHintStyle := lipgloss.NewStyle().
		Foreground(styles.TextMuted).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Border).
		Padding(0, 1)

	filterActiveStyle := lipgloss.NewStyle().
		Foreground(styles.Primary).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Primary).
		Padding(0, 1)

	// Build header
	title := headerStyle.Render("YOUR PROJECTS")

	// Subtitle shows filtered count if filter is active
	var subtitle string
	if v.filterInput.Value() != "" {
		subtitle = subHeaderStyle.Render(fmt.Sprintf("%d of %d projects", len(activeList), len(v.projects)))
	} else {
		subtitle = subHeaderStyle.Render(fmt.Sprintf("%d projects", len(v.projects)))
	}

	// Search hint or filter input
	var searchSection string
	switch {
	case v.filtering:
		// Show filter input
		searchSection = filterActiveStyle.Render("/ " + v.filterInput.View())
	case v.filterInput.Value() != "":
		// Show active filter
		searchSection = filterActiveStyle.Render("/ " + v.filterInput.Value() + " (esc to clear)")
	default:
		searchSection = searchHintStyle.Render("/ Search")
	}

	// Header row with search hint on the right
	headerWidth := v.width - 8
	titleSection := lipgloss.JoinVertical(lipgloss.Left, title, subtitle)
	headerRow := lipgloss.JoinHorizontal(
		lipgloss.Top,
		titleSection,
		lipgloss.NewStyle().Width(headerWidth-lipgloss.Width(titleSection)-lipgloss.Width(searchSection)).Render(""),
		searchSection,
	)

	// Divider
	dividerStyle := lipgloss.NewStyle().Foreground(styles.Border)
	divider := dividerStyle.Render(strings.Repeat("━", headerWidth))

	// Handle empty filtered list
	if len(activeList) == 0 && v.filterInput.Value() != "" {
		noMatchStyle := lipgloss.NewStyle().
			Foreground(styles.TextMuted).
			Italic(true)
		noMatchContent := lipgloss.JoinVertical(
			lipgloss.Left,
			headerRow,
			divider,
			"",
			noMatchStyle.Render("No projects match your filter"),
		)
		return lipgloss.NewStyle().
			Padding(1, 3).
			Width(v.width).
			Height(v.height).
			Render(noMatchContent)
	}

	// Calculate visible projects
	cardHeight := 6
	visibleCards := max((v.height-12)/cardHeight, 1)

	start := v.offset
	end := min(start+visibleCards, len(activeList))

	// Render project cards
	cards := make([]string, 0, end-start)
	for i := start; i < end; i++ {
		card := v.renderProjectCard(activeList[i], i == v.selected)
		cards = append(cards, card)
	}

	// Scroll indicator
	var scrollIndicator string
	if len(activeList) > visibleCards {
		scrollIndicator = subHeaderStyle.Render(fmt.Sprintf("  %d/%d", v.selected+1, len(activeList)))
	}

	// Combine content
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		headerRow,
		divider,
		"",
		strings.Join(cards, "\n"),
		scrollIndicator,
	)

	// Apply horizontal padding only - no height manipulation via lipgloss
	// lipgloss Height()/MaxHeight() cause issues with app.go header/footer layout
	paddedContent := lipgloss.NewStyle().
		PaddingLeft(3).
		PaddingRight(3).
		Width(v.width).
		Render(content)

	// Add top padding (1 empty line with full width for modal compositing)
	paddedContent = strings.Repeat(" ", v.width) + "\n" + paddedContent

	// Ensure EXACTLY v.height lines for modal compositing
	// Pass width so padding lines are full-width (required for overlay)
	return ensureExactHeight(paddedContent, v.height, v.width)
}

func (v *ProjectsView) renderProjectCard(p project.Project, selected bool) string {
	cardWidth := v.width - 12

	// Determine colors based on selection
	borderColor := styles.Border
	titleColor := styles.TextPrimary
	indicator := "  "
	if selected {
		borderColor = styles.Primary
		titleColor = styles.Primary
		indicator = "▸ "
	}

	// Card container style
	cardStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(0, 2).
		Width(cardWidth)

	// Title with selection indicator
	titleStyle := lipgloss.NewStyle().
		Foreground(titleColor).
		Bold(true)

	// Badge style for project type
	badgeStyle := lipgloss.NewStyle().
		Foreground(styles.Background).
		Bold(true).
		Padding(0, 1)

	if p.IsBound() {
		badgeStyle = badgeStyle.Background(styles.Info)
	} else {
		badgeStyle = badgeStyle.Background(styles.Success)
	}

	// Stats style
	statsStyle := lipgloss.NewStyle().
		Foreground(styles.TextSecondary)

	// Meta style (time, author)
	metaStyle := lipgloss.NewStyle().
		Foreground(styles.TextMuted)

	// Build the card content
	title := indicator + titleStyle.Render(p.Title)

	// Badge
	badgeText := "STANDALONE"
	if p.IsBound() {
		badgeText = "BOUND"
	}
	badge := badgeStyle.Render(badgeText)

	// Title row with badge on the right
	titleRowWidth := cardWidth - 6
	spacer := lipgloss.NewStyle().
		Width(titleRowWidth - lipgloss.Width(title) - lipgloss.Width(badge)).
		Render("")
	titleRow := lipgloss.JoinHorizontal(lipgloss.Top, title, spacer, badge)

	// Stats row (placeholder - we'd need file count from content)
	stats := statsStyle.Render("📄 Files  •  ƒ Functions")

	// Meta row
	timeAgo := formatTimeAgo(p.UpdateTime)
	author := "you"
	if p.LastModifier.Email != "" && p.LastModifier.Email != p.Creator.Email {
		author = p.LastModifier.Email
	}
	meta := metaStyle.Render(fmt.Sprintf("🕐 Updated %s by %s", timeAgo, author))

	// Combine card content
	cardContent := lipgloss.JoinVertical(
		lipgloss.Left,
		titleRow,
		stats,
		meta,
	)

	return cardStyle.Render(cardContent)
}

func (v *ProjectsView) renderEmpty() string {
	emptyStyle := lipgloss.NewStyle().
		Foreground(styles.TextMuted).
		Italic(true)

	hintStyle := lipgloss.NewStyle().
		Foreground(styles.Info).
		MarginTop(2)

	content := lipgloss.JoinVertical(
		lipgloss.Center,
		emptyStyle.Render("No projects found"),
		hintStyle.Render("Press 'n' to create a new project"),
	)

	return lipgloss.Place(
		v.width,
		v.height,
		lipgloss.Center,
		lipgloss.Center,
		content,
	)
}

// loadProjects fetches projects from the repository.
func (v *ProjectsView) loadProjects() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		result, err := v.repo.List(ctx, project.ListOptions{
			PageSize: 50,
		})
		if err != nil {
			return projectsErrorMsg{err: err}
		}

		return projectsLoadedMsg{projects: result.Projects}
	}
}

// openProject creates a command to open a project detail view.
func (v *ProjectsView) openProject(p project.Project) tea.Cmd {
	return func() tea.Msg {
		return tui.ProjectSelectedMsg{Project: p}
	}
}

// Messages for the projects view.
type projectsLoadedMsg struct {
	projects []project.Project
}

type projectsErrorMsg struct {
	err error
}

// ensureExactHeight pads or truncates content to exactly the specified height.
// This is critical for modal compositing - background must have exact line count
// AND each line must be full width for proper overlay compositing.
func ensureExactHeight(content string, height, width int) string {
	lines := strings.Split(content, "\n")

	// Truncate if too many lines
	if len(lines) > height {
		lines = lines[:height]
	}

	// Create a full-width empty line for padding
	emptyLine := strings.Repeat(" ", width)

	// Pad if too few lines - use full-width empty lines for proper modal compositing
	for len(lines) < height {
		lines = append(lines, emptyLine)
	}

	return strings.Join(lines, "\n")
}

// formatTimeAgo formats a time as a human-readable relative time.
func formatTimeAgo(t time.Time) string {
	if t.IsZero() {
		return "unknown"
	}

	duration := time.Since(t)

	switch {
	case duration < time.Minute:
		return "just now"
	case duration < time.Hour:
		mins := int(duration.Minutes())
		if mins == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", mins)
	case duration < 24*time.Hour:
		hours := int(duration.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	case duration < 7*24*time.Hour:
		days := int(duration.Hours() / 24)
		if days == 1 {
			return "yesterday"
		}
		return fmt.Sprintf("%d days ago", days)
	case duration < 30*24*time.Hour:
		weeks := int(duration.Hours() / 24 / 7)
		if weeks == 1 {
			return "1 week ago"
		}
		return fmt.Sprintf("%d weeks ago", weeks)
	case duration < 365*24*time.Hour:
		months := int(duration.Hours() / 24 / 30)
		if months == 1 {
			return "1 month ago"
		}
		return fmt.Sprintf("%d months ago", months)
	default:
		years := int(duration.Hours() / 24 / 365)
		if years == 1 {
			return "1 year ago"
		}
		return fmt.Sprintf("%d years ago", years)
	}
}
