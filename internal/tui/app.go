package tui

import (
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"yukti/internal/tui/styles"
)

// App is the main application model that coordinates all views.
type App struct {
	router *Router
	keys   KeyMap

	// Window dimensions
	width  int
	height int

	// Toast notification state
	toast      string
	toastLevel ToastLevel

	// Whether we're quitting
	quitting bool
}

// NewApp creates a new application instance with the given initial view.
func NewApp(initialView View) *App {
	return &App{
		router: NewRouter(initialView),
		keys:   DefaultKeyMap(),
		width:  80,
		height: 24,
	}
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

	if contentHeight < 1 {
		contentHeight = 1
	}

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

	headerContent := title + breadcrumb

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

	// Build help text
	helpItems := make([]string, 0, len(bindings))
	for _, binding := range bindings {
		help := binding.Help()
		item := styles.MutedStyle.Render(help.Key) + " " +
			styles.SubtitleStyle.Render(help.Desc)
		helpItems = append(helpItems, item)
	}

	helpText := lipgloss.JoinHorizontal(
		lipgloss.Left,
		helpItems...,
	)

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
