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
	"yukti/internal/tui"
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
	spinner  spinner.Model
	state    ProjectListState
	errMsg   string
	selected int
	offset   int
	width    int
	height   int
}

// NewProjectsView creates a new projects list view.
func NewProjectsView(repo project.Repository) *ProjectsView {
	// Create spinner
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(styles.Primary)

	return &ProjectsView{
		repo:     repo,
		spinner:  s,
		state:    ProjectListStateLoading,
		selected: 0,
		offset:   0,
		width:    80,
		height:   24,
	}
}

// Title implements tui.View.
func (v *ProjectsView) Title() string {
	return "Projects"
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
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.width = msg.Width
		v.height = msg.Height
		return v, nil

	case tea.KeyMsg:
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
	switch msg.String() {
	case "j", "down":
		if v.selected < len(v.projects)-1 {
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
		v.selected = len(v.projects) - 1
		v.ensureVisible()
		return nil, true
	case "r":
		v.state = ProjectListStateLoading
		return tea.Batch(v.spinner.Tick, v.loadProjects()), true
	case "enter":
		if v.selected < len(v.projects) {
			return v.openProject(v.projects[v.selected]), true
		}
	}
	return nil, false
}

// ensureVisible adjusts offset to keep selected item visible.
func (v *ProjectsView) ensureVisible() {
	cardHeight := 6 // Height of each project card including spacing
	visibleCards := max((v.height-10)/cardHeight, 1)

	if v.selected < v.offset {
		v.offset = v.selected
	}
	if v.selected >= v.offset+visibleCards {
		v.offset = v.selected - visibleCards + 1
	}
}

// View implements tea.Model.
func (v *ProjectsView) View() string {
	switch v.state {
	case ProjectListStateLoading:
		return v.renderLoading()
	case ProjectListStateError:
		return v.renderError()
	default:
		return v.renderList()
	}
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

	// Build header
	title := headerStyle.Render("YOUR PROJECTS")
	subtitle := subHeaderStyle.Render(fmt.Sprintf("%d projects", len(v.projects)))
	searchHint := searchHintStyle.Render("/ Search")

	// Header row with search hint on the right
	headerWidth := v.width - 8
	titleSection := lipgloss.JoinVertical(lipgloss.Left, title, subtitle)
	headerRow := lipgloss.JoinHorizontal(
		lipgloss.Top,
		titleSection,
		lipgloss.NewStyle().Width(headerWidth-lipgloss.Width(titleSection)-lipgloss.Width(searchHint)).Render(""),
		searchHint,
	)

	// Divider
	dividerStyle := lipgloss.NewStyle().Foreground(styles.Border)
	divider := dividerStyle.Render(strings.Repeat("━", headerWidth))

	// Calculate visible projects
	cardHeight := 6
	visibleCards := max((v.height-12)/cardHeight, 1)

	start := v.offset
	end := min(start+visibleCards, len(v.projects))

	// Render project cards
	cards := make([]string, 0, end-start)
	for i := start; i < end; i++ {
		card := v.renderProjectCard(v.projects[i], i == v.selected)
		cards = append(cards, card)
	}

	// Scroll indicator
	var scrollIndicator string
	if len(v.projects) > visibleCards {
		scrollIndicator = subHeaderStyle.Render(fmt.Sprintf("  %d/%d", v.selected+1, len(v.projects)))
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

	// Add padding
	paddedContent := lipgloss.NewStyle().
		Padding(1, 3).
		Render(content)

	return paddedContent
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
