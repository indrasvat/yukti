package views

import (
	"context"
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
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

// projectItem implements list.Item for displaying projects.
type projectItem struct {
	project project.Project
}

func (i projectItem) Title() string {
	return i.project.Title
}

func (i projectItem) Description() string {
	icon := styles.StandaloneIcon
	typeDesc := "Standalone"
	if i.project.IsBound() {
		icon = styles.BoundIcon
		typeDesc = "Bound"
	}

	timeAgo := formatTimeAgo(i.project.UpdateTime)
	return fmt.Sprintf("%s %s • Updated %s", icon, typeDesc, timeAgo)
}

func (i projectItem) FilterValue() string {
	return i.project.Title
}

// ProjectsView displays the list of Google Apps Script projects.
type ProjectsView struct {
	repo    project.Repository
	list    list.Model
	spinner spinner.Model
	state   ProjectListState
	errMsg  string
	width   int
	height  int
}

// NewProjectsView creates a new projects list view.
func NewProjectsView(repo project.Repository) *ProjectsView {
	// Create delegate for list items
	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(styles.Primary).
		BorderLeftForeground(styles.Primary)
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(styles.TextSecondary).
		BorderLeftForeground(styles.Primary)

	// Create list model
	l := list.New([]list.Item{}, delegate, 80, 24)
	l.Title = "Projects"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.SetShowHelp(false) // We use our own help
	l.Styles.Title = lipgloss.NewStyle().
		Foreground(styles.Primary).
		Bold(true).
		Padding(0, 1)
	l.Styles.FilterPrompt = lipgloss.NewStyle().
		Foreground(styles.Primary)
	l.Styles.FilterCursor = lipgloss.NewStyle().
		Foreground(styles.Primary)

	// Create spinner
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(styles.Primary)

	return &ProjectsView{
		repo:    repo,
		list:    l,
		spinner: s,
		state:   ProjectListStateLoading,
		width:   80,
		height:  24,
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
		v.list.SetSize(msg.Width, msg.Height-4) // Account for header/footer
		return v, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "r":
			// Refresh
			if v.state == ProjectListStateReady {
				v.state = ProjectListStateLoading
				return v, tea.Batch(v.spinner.Tick, v.loadProjects())
			}
		case "enter":
			// Open selected project
			if v.state == ProjectListStateReady {
				if item, ok := v.list.SelectedItem().(projectItem); ok {
					cmd := v.openProject(item.project)
					return v, cmd
				}
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
		items := make([]list.Item, 0, len(msg.projects))
		for i := range msg.projects {
			items = append(items, projectItem{project: msg.projects[i]})
		}
		v.list.SetItems(items)
		return v, nil

	case projectsErrorMsg:
		v.state = ProjectListStateError
		v.errMsg = msg.err.Error()
		return v, nil
	}

	// Pass events to the list
	if v.state == ProjectListStateReady {
		var cmd tea.Cmd
		v.list, cmd = v.list.Update(msg)
		cmds = append(cmds, cmd)
	}

	return v, tea.Batch(cmds...)
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
		loadingStyle.Render("Loading projects..."),
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
		titleStyle.Render("Failed to load projects"),
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
	return v.list.View()
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
