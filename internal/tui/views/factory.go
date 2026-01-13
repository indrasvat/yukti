package views

import (
	appprocess "yukti/internal/application/process"
	"yukti/internal/domain/project"
	"yukti/internal/tui"
)

// Factory implements tui.ViewFactory for creating views.
type Factory struct {
	processService *appprocess.Service
}

// NewFactory creates a new view factory.
func NewFactory() *Factory {
	return &Factory{}
}

// NewFactoryWithService creates a new view factory with a process service for script execution.
func NewFactoryWithService(processService *appprocess.Service) *Factory {
	return &Factory{
		processService: processService,
	}
}

// CreateProjectsView creates a new projects view.
func (f *Factory) CreateProjectsView(repo project.Repository) tui.View {
	return NewProjectsView(repo)
}

// CreateProjectDetailView creates a new project detail view.
// This now returns the new WorkspaceView with split-pane layout.
func (f *Factory) CreateProjectDetailView(proj project.Project, repo project.Repository) tui.View {
	return NewWorkspaceViewWithService(proj, repo, f.processService)
}

// CreateCodeViewerView creates a new code viewer for a file.
func (f *Factory) CreateCodeViewerView(file project.File) tui.View {
	return NewCodeViewerView(file)
}

// CreateWorkspaceView creates a new workspace view with split-pane layout.
func (f *Factory) CreateWorkspaceView(proj project.Project, repo project.Repository) tui.View {
	return NewWorkspaceViewWithService(proj, repo, f.processService)
}
