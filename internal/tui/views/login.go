package views

import (
	"context"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"yukti/internal/infrastructure/google"
	"yukti/internal/tui/styles"
)

// LoginState represents the current state of the login view.
type LoginState int

const (
	LoginStateIdle LoginState = iota
	LoginStateAuthenticating
	LoginStateSuccess
	LoginStateError
)

// LoginView handles user authentication.
type LoginView struct {
	auth    *google.Authenticator
	width   int
	height  int
	state   LoginState
	spinner spinner.Model
	errMsg  string
}

// NewLoginView creates a new login view.
func NewLoginView(auth *google.Authenticator) *LoginView {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(styles.Primary)

	return &LoginView{
		auth:    auth,
		state:   LoginStateIdle,
		spinner: s,
		width:   80,
		height:  24,
	}
}

// Title implements tui.View.
func (v *LoginView) Title() string {
	return "Login"
}

// ShortHelp implements tui.View.
func (v *LoginView) ShortHelp() []key.Binding {
	if v.state == LoginStateIdle {
		return []key.Binding{
			key.NewBinding(
				key.WithKeys("enter"),
				key.WithHelp("enter", "login with Google"),
			),
		}
	}
	return nil
}

// Init implements tea.Model.
func (v *LoginView) Init() tea.Cmd {
	return v.spinner.Tick
}

// Update implements tea.Model.
func (v *LoginView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.width = msg.Width
		v.height = msg.Height
		return v, nil

	case tea.KeyMsg:
		if v.state == LoginStateIdle && msg.String() == "enter" {
			v.state = LoginStateAuthenticating
			return v, tea.Batch(
				v.spinner.Tick,
				v.startLogin(),
			)
		}

	case spinner.TickMsg:
		if v.state == LoginStateAuthenticating {
			var cmd tea.Cmd
			v.spinner, cmd = v.spinner.Update(msg)
			return v, cmd
		}

	case loginSuccessMsg:
		v.state = LoginStateSuccess
		// Return a command to navigate to the project list
		cmd := v.onLoginSuccess()
		return v, cmd

	case loginErrorMsg:
		v.state = LoginStateError
		v.errMsg = msg.err.Error()
		return v, nil
	}

	return v, nil
}

// View implements tea.Model.
func (v *LoginView) View() string {
	var content string

	switch v.state {
	case LoginStateIdle:
		content = v.renderIdleState()
	case LoginStateAuthenticating:
		content = v.renderAuthenticatingState()
	case LoginStateSuccess:
		content = v.renderSuccessState()
	case LoginStateError:
		content = v.renderErrorState()
	}

	return lipgloss.Place(
		v.width,
		v.height,
		lipgloss.Center,
		lipgloss.Center,
		content,
	)
}

func (v *LoginView) renderIdleState() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(styles.Primary).
		MarginBottom(1)

	subtitleStyle := lipgloss.NewStyle().
		Foreground(styles.TextSecondary).
		MarginBottom(2)

	buttonStyle := lipgloss.NewStyle().
		Foreground(styles.White).
		Background(styles.Primary).
		Padding(0, 3).
		Bold(true).
		MarginTop(2)

	helpStyle := lipgloss.NewStyle().
		Foreground(styles.TextMuted).
		MarginTop(2)

	title := titleStyle.Render("Welcome to Yukti")
	subtitle := subtitleStyle.Render("A beautiful TUI for Google Apps Script")

	button := buttonStyle.Render("Login with Google")
	help := helpStyle.Render("Press Enter to authenticate")

	return lipgloss.JoinVertical(
		lipgloss.Center,
		title,
		subtitle,
		"",
		button,
		help,
	)
}

func (v *LoginView) renderAuthenticatingState() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(styles.Primary).
		MarginBottom(1)

	messageStyle := lipgloss.NewStyle().
		Foreground(styles.TextSecondary)

	title := titleStyle.Render("Authenticating...")
	spinnerView := v.spinner.View()
	message := messageStyle.Render("Complete the login in your browser")

	return lipgloss.JoinVertical(
		lipgloss.Center,
		title,
		"",
		spinnerView,
		"",
		message,
	)
}

func (v *LoginView) renderSuccessState() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(styles.Success).
		MarginBottom(1)

	title := titleStyle.Render("Login Successful!")
	message := styles.SubtitleStyle.Render("Loading your projects...")

	return lipgloss.JoinVertical(
		lipgloss.Center,
		title,
		message,
	)
}

func (v *LoginView) renderErrorState() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(styles.Error).
		MarginBottom(1)

	messageStyle := lipgloss.NewStyle().
		Foreground(styles.TextSecondary).
		MarginBottom(2)

	helpStyle := lipgloss.NewStyle().
		Foreground(styles.TextMuted)

	title := titleStyle.Render("Login Failed")
	message := messageStyle.Render(v.errMsg)
	help := helpStyle.Render("Press Enter to try again")

	return lipgloss.JoinVertical(
		lipgloss.Center,
		title,
		message,
		help,
	)
}

// startLogin initiates the OAuth login flow.
func (v *LoginView) startLogin() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		_, err := v.auth.Login(ctx)
		if err != nil {
			return loginErrorMsg{err: err}
		}
		return loginSuccessMsg{}
	}
}

// onLoginSuccess returns a command after successful login.
// This can be customized to navigate to a specific view.
func (v *LoginView) onLoginSuccess() tea.Cmd {
	// For now, just return nil - the parent view can handle this
	// In a full implementation, this would navigate to the project list
	return nil
}

// Messages for login flow
type loginSuccessMsg struct{}
type loginErrorMsg struct {
	err error
}
