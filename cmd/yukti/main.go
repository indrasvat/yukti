// Package main is the entry point for the Yukti TUI application.
package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"yukti/internal/cli"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7C3AED"))

	subtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#9AA5CE"))

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#565F89"))
)

type model struct {
	width  int
	height int
}

func initialModel() model {
	return model{}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m model) View() string {
	title := titleStyle.Render("⚡ Yukti (युक्ति)")
	subtitle := subtitleStyle.Render("Beautiful TUI for Google Apps Script")

	version := infoStyle.Render(fmt.Sprintf(
		"Version: %s | Commit: %s | Built: %s",
		cli.Version, cli.Commit, cli.BuildDate,
	))

	help := infoStyle.Render("Press 'q' to quit")

	content := lipgloss.JoinVertical(
		lipgloss.Center,
		"",
		title,
		subtitle,
		"",
		version,
		"",
		help,
		"",
	)

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		content,
	)
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
