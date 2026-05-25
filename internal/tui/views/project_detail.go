package views

import (
	"context"
	"fmt"
	"strings"
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

// ProjectDetailState represents the current state of the project detail view.
type ProjectDetailState int

const (
	ProjectDetailStateLoading ProjectDetailState = iota
	ProjectDetailStateReady
	ProjectDetailStateError
)

// fileItem implements list.Item for displaying project files.
type fileItem struct {
	file project.File
}

func (i fileItem) Title() string {
	return i.file.Name
}

func (i fileItem) Description() string {
	icon := fileTypeIcon(i.file.Type)
	desc := string(i.file.Type)

	// Add function count if available
	if i.file.FunctionSet != nil && len(i.file.FunctionSet.Functions) > 0 {
		desc += fmt.Sprintf(" • %d functions", len(i.file.FunctionSet.Functions))
	}

	// Add line count estimate
	lines := strings.Count(i.file.Source, "\n") + 1
	desc += fmt.Sprintf(" • %d lines", lines)

	return fmt.Sprintf("%s %s", icon, desc)
}

func (i fileItem) FilterValue() string {
	return i.file.Name
}

// fileTypeIcon returns an icon for the file type.
func fileTypeIcon(ft project.FileType) string {
	switch ft {
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

// ProjectDetailView displays the details of a single project.
type ProjectDetailView struct {
	proj    project.Project
	repo    project.Repository
	content *project.Content
	list    list.Model
	spinner spinner.Model
	state   ProjectDetailState
	errMsg  string
	width   int
	height  int
}

// NewProjectDetailView creates a new project detail view.
func NewProjectDetailView(proj project.Project, repo project.Repository) *ProjectDetailView {
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
	l.Title = "Files"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.SetShowHelp(false)
	l.Styles.Title = lipgloss.NewStyle().
		Foreground(styles.Primary).
		Bold(true).
		Padding(0, 1)

	// Create spinner
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(styles.Primary)

	return &ProjectDetailView{
		proj:    proj,
		repo:    repo,
		list:    l,
		spinner: s,
		state:   ProjectDetailStateLoading,
		width:   80,
		height:  24,
	}
}

// Title implements tui.View.
func (v *ProjectDetailView) Title() string {
	return v.proj.Title
}

// ShortHelp implements tui.View.
func (v *ProjectDetailView) ShortHelp() []key.Binding {
	if v.state != ProjectDetailStateReady {
		return nil
	}

	return []key.Binding{
		key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "view"),
		),
		key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "filter"),
		),
		key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "refresh"),
		),
	}
}

// Init implements tea.Model.
func (v *ProjectDetailView) Init() tea.Cmd {
	return tea.Batch(
		v.spinner.Tick,
		v.loadContent(),
	)
}

// Update implements tea.Model.
func (v *ProjectDetailView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.width = msg.Width
		v.height = msg.Height
		v.list.SetSize(msg.Width, msg.Height-6) // Account for header/footer/project info
		return v, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "r":
			// Refresh
			if v.state == ProjectDetailStateReady {
				v.state = ProjectDetailStateLoading
				return v, tea.Batch(v.spinner.Tick, v.loadContent())
			}
		case "enter":
			// View selected file
			if v.state == ProjectDetailStateReady {
				if item, ok := v.list.SelectedItem().(fileItem); ok {
					return v, func() tea.Msg {
						return tui.FileSelectedMsg{File: item.file}
					}
				}
			}
		}

	case spinner.TickMsg:
		if v.state == ProjectDetailStateLoading {
			var cmd tea.Cmd
			v.spinner, cmd = v.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}

	case contentLoadedMsg:
		v.state = ProjectDetailStateReady
		v.content = msg.content
		items := make([]list.Item, 0, len(msg.content.Files))
		for i := range msg.content.Files {
			items = append(items, fileItem{file: msg.content.Files[i]})
		}
		v.list.SetItems(items)
		return v, nil

	case contentErrorMsg:
		v.state = ProjectDetailStateError
		v.errMsg = msg.err.Error()
		return v, nil
	}

	// Pass events to the list
	if v.state == ProjectDetailStateReady {
		var cmd tea.Cmd
		v.list, cmd = v.list.Update(msg)
		cmds = append(cmds, cmd)
	}

	return v, tea.Batch(cmds...)
}

// View implements tea.Model.
func (v *ProjectDetailView) View() string {
	switch v.state {
	case ProjectDetailStateLoading:
		return v.renderLoading()
	case ProjectDetailStateError:
		return v.renderError()
	default:
		return v.renderContent()
	}
}

func (v *ProjectDetailView) renderLoading() string {
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

func (v *ProjectDetailView) renderError() string {
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
		titleStyle.Render("Failed to load project content"),
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

func (v *ProjectDetailView) renderContent() string {
	// Project info header
	infoStyle := lipgloss.NewStyle().
		Foreground(styles.TextMuted).
		Padding(0, 2)

	info := fmt.Sprintf("ID: %s • Last modified: %s",
		truncateID(v.proj.ID),
		formatTimeAgo(v.proj.UpdateTime))

	header := infoStyle.Render(info)

	// File list
	fileList := v.list.View()

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		fileList,
	)
}

// loadContent fetches project content from the repository.
func (v *ProjectDetailView) loadContent() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		content, err := v.repo.GetContent(ctx, v.proj.ID)
		if err != nil {
			return contentErrorMsg{err: err}
		}

		return contentLoadedMsg{content: content}
	}
}

// Messages for the project detail view.
type contentLoadedMsg struct {
	content *project.Content
}

type contentErrorMsg struct {
	err error
}

// truncateID truncates a long ID for display.
func truncateID(id string) string {
	if len(id) <= 12 {
		return id
	}
	return id[:6] + "..." + id[len(id)-4:]
}
