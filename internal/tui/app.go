package tui

import (
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"yukti/internal/domain/project"
	"yukti/internal/tui/styles"
)

// AuthState represents the authentication status.
type AuthState int

const (
	AuthStateUnknown AuthState = iota
	AuthStateLoggedOut
	AuthStateLoggedIn
)

// ViewFactory creates views - used to avoid circular imports between tui and views.
type ViewFactory interface {
	CreateProjectsView(repo project.Repository) View
	CreateProjectDetailView(proj project.Project, repo project.Repository) View
	CreateCodeViewerView(file project.File) View
}

// AppOptions configures the application.
type AppOptions struct {
	AuthState   AuthState
	UserEmail   string
	ViewFactory ViewFactory
}

// App is the main application model that coordinates all views.
type App struct {
	router *Router
	keys   KeyMap

	// Window dimensions
	width  int
	height int

	// Auth state
	authState AuthState
	userEmail string

	// Project repository (nil if not authenticated)
	projectRepo project.Repository

	// View factory for creating views (avoids circular imports)
	viewFactory ViewFactory

	// Toast notification state
	toast      string
	toastLevel ToastLevel

	// Whether we're quitting
	quitting bool
}

// NewApp creates a new application instance with the given initial view.
func NewApp(initialView View, opts AppOptions, projectRepo project.Repository) *App {
	return &App{
		router:      NewRouter(initialView),
		keys:        DefaultKeyMap(),
		width:       80,
		height:      24,
		authState:   opts.AuthState,
		userEmail:   opts.UserEmail,
		projectRepo: projectRepo,
		viewFactory: opts.ViewFactory,
	}
}

// ProjectRepo returns the project repository for views to use.
func (a *App) ProjectRepo() project.Repository {
	return a.projectRepo
}

// Init implements tea.Model.
func (a *App) Init() tea.Cmd {
	return a.router.Current().Init()
}

// Update implements tea.Model.
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle navigation messages first
	if cmd, handled := a.handleNavigation(msg); handled {
		return a, cmd
	}

	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if cmd, handled := a.handleKeyMsg(msg); handled {
			return a, cmd
		}

	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		// Adjust the message for views: subtract header (3) and footer (3) heights
		// so views render to the content area, not the full terminal
		msg.Height = max(1, msg.Height-6)

	case ToastMsg:
		a.toast = msg.Message
		a.toastLevel = msg.Level
		cmds = append(cmds, clearToastAfterDelay())

	case clearToastMsg:
		a.toast = ""
		return a, nil
	}

	// Update current view
	currentView := a.router.Current()
	updatedModel, cmd := currentView.Update(msg)
	if updatedView, ok := updatedModel.(View); ok {
		a.router.Replace(updatedView)
	}
	cmds = append(cmds, cmd)

	return a, tea.Batch(cmds...)
}

// handleKeyMsg handles global keyboard shortcuts.
func (a *App) handleKeyMsg(msg tea.KeyMsg) (tea.Cmd, bool) {
	switch {
	case key.Matches(msg, a.keys.Quit):
		a.quitting = true
		return tea.Quit, true

	case key.Matches(msg, a.keys.Back):
		if a.router.CanGoBack() {
			a.router.Pop()
			return nil, true
		}
	}
	return nil, false
}

// handleNavigation handles all navigation-related messages.
func (a *App) handleNavigation(msg tea.Msg) (tea.Cmd, bool) {
	switch msg := msg.(type) {
	case NavigateMsg:
		initCmd := a.router.Push(msg.View)
		sizeCmd := func() tea.Msg {
			return tea.WindowSizeMsg{Width: a.width, Height: a.height}
		}
		return tea.Batch(initCmd, sizeCmd), true

	case BackMsg:
		if a.router.CanGoBack() {
			a.router.Pop()
		}
		return nil, true

	case NavigateToProjectsMsg:
		if a.projectRepo != nil {
			return a.navigateToProjects(), true
		}
		a.toast = "Please login first"
		a.toastLevel = ToastWarning
		return clearToastAfterDelay(), true

	case ProjectSelectedMsg:
		if a.viewFactory != nil && a.projectRepo != nil {
			view := a.viewFactory.CreateProjectDetailView(msg.Project, a.projectRepo)
			return Navigate(view), true
		}
		return nil, true

	case FileSelectedMsg:
		if a.viewFactory != nil {
			view := a.viewFactory.CreateCodeViewerView(msg.File)
			return Navigate(view), true
		}
		return nil, true
	}
	return nil, false
}

// View implements tea.Model.
func (a *App) View() string {
	if a.quitting {
		return ""
	}

	// Calculate available height for content
	headerHeight := 3
	footerHeight := 3
	contentHeight := a.height - headerHeight - footerHeight

	contentHeight = max(1, contentHeight)

	// Build layout components
	header := a.renderHeader()
	content := a.renderContent(contentHeight)
	footer := a.renderFooter()

	// Compose the full view
	view := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		content,
		footer,
	)

	// Render the view onto a fixed-size canvas with explicit background on every cell
	return renderFullScreen(view, a.width, a.height)
}

// renderHeader renders the top navigation bar.
func (a *App) renderHeader() string {
	title := styles.TitleStyle.Render("⚡ Yukti")
	viewTitle := a.router.Current().Title()

	// Show breadcrumb if we have navigation history
	var breadcrumb string
	if a.router.CanGoBack() {
		breadcrumb = styles.MutedStyle.Render(" › ") + viewTitle
	} else {
		breadcrumb = styles.SubtitleStyle.Render(" - " + viewTitle)
	}

	leftContent := title + breadcrumb

	// Build auth status indicator
	var authStatus string
	switch a.authState {
	case AuthStateLoggedIn:
		statusStyle := lipgloss.NewStyle().
			Foreground(styles.Success).
			Bold(true)
		emailStyle := lipgloss.NewStyle().
			Foreground(styles.TextSecondary)

		indicator := statusStyle.Render("●")
		if a.userEmail != "" {
			authStatus = indicator + " " + emailStyle.Render(a.userEmail)
		} else {
			authStatus = indicator + " " + emailStyle.Render("Logged in")
		}
	case AuthStateLoggedOut:
		statusStyle := lipgloss.NewStyle().
			Foreground(styles.TextMuted)
		authStatus = statusStyle.Render("○ Not logged in")
	}

	// Calculate spacing to right-align auth status
	leftWidth := lipgloss.Width(leftContent)
	rightWidth := lipgloss.Width(authStatus)
	availableWidth := a.width - 4 // Account for padding
	spacing := availableWidth - leftWidth - rightWidth

	spacing = max(1, spacing)

	// Spacer must have background color to prevent bleed
	spacer := lipgloss.NewStyle().
		Width(spacing).
		Background(styles.Background).
		Render(strings.Repeat(" ", spacing))
	headerContent := leftContent + spacer + authStatus

	return styles.HeaderStyle.
		Width(a.width).
		Render(headerContent)
}

// renderContent renders the main content area.
func (a *App) renderContent(_ int) string {
	return a.router.Current().View()
}

// renderFooter renders the bottom help bar.
func (a *App) renderFooter() string {
	// Get help from current view
	bindings := a.router.Current().ShortHelp()

	// Add global bindings
	if a.router.CanGoBack() {
		bindings = append(bindings, a.keys.Back)
	}
	bindings = append(bindings, a.keys.Quit)

	// Build help text with proper styling
	keyStyle := lipgloss.NewStyle().
		Foreground(styles.Primary).
		Bold(true)

	descStyle := lipgloss.NewStyle().
		Foreground(styles.TextSecondary)

	separatorStyle := lipgloss.NewStyle().
		Foreground(styles.TextMuted).
		Padding(0, 1)

	helpItems := make([]string, 0, len(bindings))
	for _, binding := range bindings {
		help := binding.Help()
		item := keyStyle.Render(help.Key) + " " + descStyle.Render(help.Desc)
		helpItems = append(helpItems, item)
	}

	// Join with separator
	separator := separatorStyle.Render("│")
	helpText := lipgloss.JoinHorizontal(lipgloss.Left)
	for i, item := range helpItems {
		if i > 0 {
			helpText = lipgloss.JoinHorizontal(lipgloss.Left, helpText, separator)
		}
		helpText = lipgloss.JoinHorizontal(lipgloss.Left, helpText, item)
	}

	// Add toast if present
	if a.toast != "" {
		toastStyle := styles.InfoBadge
		switch a.toastLevel {
		case ToastSuccess:
			toastStyle = styles.SuccessBadge
		case ToastWarning:
			toastStyle = styles.WarningBadge
		case ToastError:
			toastStyle = styles.ErrorBadge
		}
		helpText = toastStyle(a.toast) + "  " + helpText
	}

	return styles.FooterStyle.
		Width(a.width).
		Render(helpText)
}

// clearToastMsg is sent to clear the toast notification.
type clearToastMsg struct{}

// clearToastAfterDelay returns a command that clears the toast after 3 seconds.
func clearToastAfterDelay() tea.Cmd {
	return tea.Tick(3*time.Second, func(_ time.Time) tea.Msg {
		return clearToastMsg{}
	})
}

// navigateToProjects creates and navigates to the projects view.
func (a *App) navigateToProjects() tea.Cmd {
	// Import views package dynamically via ViewFactory to avoid circular imports
	// The actual view creation is done via a factory function set during initialization
	if a.viewFactory != nil {
		view := a.viewFactory.CreateProjectsView(a.projectRepo)
		return Navigate(view)
	}
	return nil
}

// renderFullScreen places content on a fixed-size canvas.
// The terminal's background color is set via termenv in cli/tui.go,
// so empty cells use our app's background color automatically.
func renderFullScreen(content string, width, height int) string {
	return lipgloss.Place(
		width,
		height,
		lipgloss.Left,
		lipgloss.Top,
		content,
	)
}
