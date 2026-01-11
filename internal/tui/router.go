package tui

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// View represents a screen/page in the application.
type View interface {
	tea.Model

	// Title returns the view's title for the header.
	Title() string

	// ShortHelp returns key bindings to show in the footer.
	ShortHelp() []key.Binding
}

// Router manages view navigation with a stack-based history.
type Router struct {
	current View
	stack   []View
}

// NewRouter creates a new router with the given initial view.
func NewRouter(initial View) *Router {
	return &Router{
		current: initial,
		stack:   make([]View, 0),
	}
}

// Current returns the current view.
func (r *Router) Current() View {
	return r.current
}

// Push navigates to a new view, saving the current view to history.
func (r *Router) Push(view View) tea.Cmd {
	r.stack = append(r.stack, r.current)
	r.current = view
	return r.current.Init()
}

// Pop returns to the previous view.
// Returns false if there's no history to go back to.
func (r *Router) Pop() bool {
	if len(r.stack) == 0 {
		return false
	}

	// Pop from stack
	r.current = r.stack[len(r.stack)-1]
	r.stack = r.stack[:len(r.stack)-1]
	return true
}

// CanGoBack returns true if there's history to navigate back to.
func (r *Router) CanGoBack() bool {
	return len(r.stack) > 0
}

// StackDepth returns the number of views in the history stack.
func (r *Router) StackDepth() int {
	return len(r.stack)
}

// Clear removes all history, keeping only the current view.
func (r *Router) Clear() {
	r.stack = r.stack[:0]
}

// Replace replaces the current view without affecting history.
func (r *Router) Replace(view View) tea.Cmd {
	r.current = view
	return r.current.Init()
}
