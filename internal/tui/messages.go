package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"yukti/internal/domain/project"
)

// NavigateMsg requests navigation to a new view.
type NavigateMsg struct {
	View View
}

// Navigate returns a command that navigates to the given view.
func Navigate(view View) tea.Cmd {
	return func() tea.Msg {
		return NavigateMsg{View: view}
	}
}

// BackMsg requests navigation back to the previous view.
type BackMsg struct{}

// Back returns a command that navigates back.
func Back() tea.Cmd {
	return func() tea.Msg {
		return BackMsg{}
	}
}

// ToastMsg displays a toast notification.
type ToastMsg struct {
	Message string
	Level   ToastLevel
}

// ToastLevel defines the severity of a toast message.
type ToastLevel string

const (
	ToastInfo    ToastLevel = "info"
	ToastSuccess ToastLevel = "success"
	ToastWarning ToastLevel = "warning"
	ToastError   ToastLevel = "error"
)

// ShowToast returns a command that displays a toast.
func ShowToast(message string, level ToastLevel) tea.Cmd {
	return func() tea.Msg {
		return ToastMsg{Message: message, Level: level}
	}
}

// ErrorMsg represents an error that occurred.
type ErrorMsg struct {
	Err error
}

// Error implements the error interface.
func (e ErrorMsg) Error() string {
	return e.Err.Error()
}

// WindowSizeMsg is re-exported for convenience.
type WindowSizeMsg = tea.WindowSizeMsg

// NavigateToProjectsMsg requests navigation to the projects view.
type NavigateToProjectsMsg struct{}

// NavigateToProjects returns a command to navigate to the projects view.
func NavigateToProjects() tea.Cmd {
	return func() tea.Msg {
		return NavigateToProjectsMsg{}
	}
}

// ProjectSelectedMsg is sent when a project is selected from the list.
type ProjectSelectedMsg struct {
	Project project.Project
}

// FileSelectedMsg is sent when a file is selected from the project detail view.
type FileSelectedMsg struct {
	File project.File
}
