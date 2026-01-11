package tui

import (
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"yukti/internal/tui/styles"
)

// AuthState represents the authentication status.
type AuthState int

const (
	AuthStateUnknown AuthState = iota
	AuthStateLoggedOut
	AuthStateLoggedIn
)

// AppOptions configures the application.
type AppOptions struct {
	AuthState AuthState
	UserEmail string
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

	// Toast notification state
	toast      string
	toastLevel ToastLevel

	// Whether we're quitting
	quitting bool
}

// NewApp creates a new application instance with the given initial view.
func NewApp(initialView View, opts ...AppOptions) *App {
	app := &App{
		router:    NewRouter(initialView),
		keys:      DefaultKeyMap(),
		width:     80,
		height:    24,
		authState: AuthStateUnknown,
	}

	if len(opts) > 0 {
		app.authState = opts[0].AuthState
		app.userEmail = opts[0].UserEmail
	}

	return app
}

// Init implements tea.Model.
func (a *App) Init() tea.Cmd {
	return a.router.Current().Init()
}

// Update implements tea.Model.
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Global key handling
		switch {
		case key.Matches(msg, a.keys.Quit):
			a.quitting = true
			return a, tea.Quit

		case key.Matches(msg, a.keys.Back):
			if a.router.CanGoBack() {
				a.router.Pop()
				return a, nil
			}
		}

	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height

	case NavigateMsg:
		cmd := a.router.Push(msg.View)
		return a, cmd

	case BackMsg:
		if a.router.CanGoBack() {
			a.router.Pop()
		}
		return a, nil

	case ToastMsg:
		a.toast = msg.Message
		a.toastLevel = msg.Level
		// Clear toast after a delay
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

	// Build layout
	header := a.renderHeader()
	content := a.renderContent(contentHeight)
	footer := a.renderFooter()

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		content,
		footer,
	)
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

	spacer := lipgloss.NewStyle().Width(spacing).Render("")
	headerContent := leftContent + spacer + authStatus

	return styles.HeaderStyle.
		Width(a.width).
		Render(headerContent)
}

// renderContent renders the main content area.
func (a *App) renderContent(height int) string {
	contentStyle := lipgloss.NewStyle().
		Width(a.width).
		Height(height).
		Background(styles.Background)

	return contentStyle.Render(a.router.Current().View())
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
		Background(styles.Overlay).
		Padding(0, 1).
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
