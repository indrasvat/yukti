package views

import (
	"yukti/internal/domain/project"
	"yukti/internal/tui"
)

// Factory implements tui.ViewFactory for creating views.
type Factory struct{}

// NewFactory creates a new view factory.
func NewFactory() *Factory {
	return &Factory{}
}

// CreateProjectsView creates a new projects view.
func (f *Factory) CreateProjectsView(repo project.Repository) tui.View {
	return NewProjectsView(repo)
}

// CreateProjectDetailView creates a new project detail view.
func (f *Factory) CreateProjectDetailView(proj project.Project, repo project.Repository) tui.View {
	return NewProjectDetailView(proj, repo)
}

// CreateCodeViewerView creates a new code viewer for a file.
func (f *Factory) CreateCodeViewerView(file project.File) tui.View {
	return NewCodeViewerView(file)
}
