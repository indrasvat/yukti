package workspace

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"yukti/internal/domain/project"
)

func TestCloneMaterializesFilesAndManifest(t *testing.T) {
	t.Parallel()

	repo := newWorkspaceRepo()
	repo.projects["script-1"] = &project.Project{ID: "script-1", Title: "Demo"}
	repo.contents["script-1"] = demoContent()

	dir := filepath.Join(t.TempDir(), "demo")
	result, err := NewService(repo).Clone(context.Background(), CloneOptions{
		ScriptID: "script-1",
		Dir:      dir,
	})
	if err != nil {
		t.Fatalf("Clone() error = %v", err)
	}

	if result.ScriptID != "script-1" {
		t.Fatalf("ScriptID = %q", result.ScriptID)
	}
	assertFileContains(t, dir, "Code.gs", "function main")
	assertFileContains(t, dir, "appsscript.json", "STACKDRIVER")

	manifest, err := LoadManifest(dir)
	if err != nil {
		t.Fatalf("LoadManifest() error = %v", err)
	}
	if manifest.LastRemoteHash == "" {
		t.Fatal("manifest LastRemoteHash is empty")
	}
}

func TestPullRefusesToOverwriteDirtyWorkspaceWithoutForce(t *testing.T) {
	t.Parallel()

	repo := newWorkspaceRepo()
	repo.projects["script-1"] = &project.Project{ID: "script-1", Title: "Demo"}
	repo.contents["script-1"] = demoContent()

	dir := filepath.Join(t.TempDir(), "demo")
	service := NewService(repo)
	if _, err := service.Clone(context.Background(), CloneOptions{ScriptID: "script-1", Dir: dir}); err != nil {
		t.Fatalf("Clone() error = %v", err)
	}
	writeFile(t, dir, "Code.gs", "function changed() {}")

	_, err := service.Pull(context.Background(), PullOptions{Dir: dir})
	if err == nil || !strings.Contains(err.Error(), "unpushed changes") {
		t.Fatalf("Pull() error = %v, want unpushed changes", err)
	}
}

func TestPullForceRemovesTrackedFilesDeletedRemotely(t *testing.T) {
	t.Parallel()

	repo := newWorkspaceRepo()
	repo.projects["script-1"] = &project.Project{ID: "script-1", Title: "Demo"}
	repo.contents["script-1"] = &project.Content{ScriptID: "script-1", Files: []project.File{
		{Name: "appsscript", Type: project.FileTypeJSON, Source: "{}"},
		{Name: "Code", Type: project.FileTypeServer, Source: "function main() {}"},
		{Name: "Old", Type: project.FileTypeServer, Source: "function old() {}"},
	}}

	dir := filepath.Join(t.TempDir(), "demo")
	service := NewService(repo)
	if _, err := service.Clone(context.Background(), CloneOptions{ScriptID: "script-1", Dir: dir}); err != nil {
		t.Fatalf("Clone() error = %v", err)
	}

	repo.contents["script-1"] = &project.Content{ScriptID: "script-1", Files: []project.File{
		{Name: "appsscript", Type: project.FileTypeJSON, Source: "{}"},
		{Name: "Code", Type: project.FileTypeServer, Source: "function main() {}"},
	}}

	if _, err := service.Pull(context.Background(), PullOptions{Dir: dir, Force: true}); err != nil {
		t.Fatalf("Pull() error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "Old.gs")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("Old.gs still exists after pull; stat err = %v", err)
	}
}

func TestPushRefusesWhenRemoteHeadChanged(t *testing.T) {
	t.Parallel()

	repo := newWorkspaceRepo()
	repo.projects["script-1"] = &project.Project{ID: "script-1", Title: "Demo"}
	repo.contents["script-1"] = demoContent()

	dir := filepath.Join(t.TempDir(), "demo")
	service := NewService(repo)
	if _, err := service.Clone(context.Background(), CloneOptions{ScriptID: "script-1", Dir: dir}); err != nil {
		t.Fatalf("Clone() error = %v", err)
	}
	writeFile(t, dir, "Code.gs", "function local() {}")

	repo.contents["script-1"] = &project.Content{ScriptID: "script-1", Files: []project.File{
		{Name: "appsscript", Type: project.FileTypeJSON, Source: "{}"},
		{Name: "Code", Type: project.FileTypeServer, Source: "function remote() {}"},
	}}

	_, err := service.Push(context.Background(), PushOptions{Dir: dir})
	if err == nil || !strings.Contains(err.Error(), "remote HEAD changed") {
		t.Fatalf("Push() error = %v, want remote changed", err)
	}
}

func TestPushUpdatesRemoteAndRefreshesManifest(t *testing.T) {
	t.Parallel()

	repo := newWorkspaceRepo()
	repo.projects["script-1"] = &project.Project{ID: "script-1", Title: "Demo"}
	repo.contents["script-1"] = demoContent()

	dir := filepath.Join(t.TempDir(), "demo")
	service := NewService(repo)
	if _, err := service.Clone(context.Background(), CloneOptions{ScriptID: "script-1", Dir: dir}); err != nil {
		t.Fatalf("Clone() error = %v", err)
	}
	writeFile(t, dir, "Code.gs", "function local() {}")

	result, err := service.Push(context.Background(), PushOptions{Dir: dir})
	if err != nil {
		t.Fatalf("Push() error = %v", err)
	}
	if !Dirty(result.Changes) {
		t.Fatal("Push() changes were unexpectedly clean")
	}

	remote := repo.contents["script-1"]
	if remoteFileSource(remote, "Code") != "function local() {}" {
		t.Fatalf("remote Code source = %q", remoteFileSource(remote, "Code"))
	}

	status, err := service.Status(dir)
	if err != nil {
		t.Fatalf("Status() error = %v", err)
	}
	if Dirty(status.Changes) {
		t.Fatalf("workspace should be clean after push: %+v", status.Changes)
	}
}

type workspaceRepo struct {
	projects map[string]*project.Project
	contents map[string]*project.Content
}

func newWorkspaceRepo() *workspaceRepo {
	return &workspaceRepo{
		projects: make(map[string]*project.Project),
		contents: make(map[string]*project.Content),
	}
}

func (r *workspaceRepo) List(context.Context, project.ListOptions) (*project.ListResult, error) {
	return &project.ListResult{}, nil
}

func (r *workspaceRepo) Get(_ context.Context, id string) (*project.Project, error) {
	if proj, ok := r.projects[id]; ok {
		copyProj := *proj
		return &copyProj, nil
	}
	return nil, errors.New("project not found")
}

func (r *workspaceRepo) GetContent(_ context.Context, id string) (*project.Content, error) {
	content, ok := r.contents[id]
	if !ok {
		return nil, errors.New("content not found")
	}
	return cloneContent(content), nil
}

func (r *workspaceRepo) GetMetrics(context.Context, string, project.MetricsOptions) (*project.Metrics, error) {
	return nil, nil
}

func (r *workspaceRepo) Create(_ context.Context, req project.CreateRequest) (*project.Project, error) {
	id := "created-script"
	r.projects[id] = &project.Project{ID: id, Title: req.Title, ParentID: req.ParentID}
	r.contents[id] = &project.Content{ScriptID: id, Files: []project.File{
		{Name: "appsscript", Type: project.FileTypeJSON, Source: "{}"},
	}}
	return r.projects[id], nil
}

func (r *workspaceRepo) UpdateContent(_ context.Context, id string, content *project.Content) error {
	r.contents[id] = cloneContent(content)
	r.contents[id].ScriptID = id
	return nil
}

func demoContent() *project.Content {
	return &project.Content{ScriptID: "script-1", Files: []project.File{
		{Name: "appsscript", Type: project.FileTypeJSON, Source: `{"exceptionLogging":"STACKDRIVER"}`},
		{Name: "Code", Type: project.FileTypeServer, Source: "function main() {}"},
	}}
}

func cloneContent(content *project.Content) *project.Content {
	next := &project.Content{ScriptID: content.ScriptID, Files: make([]project.File, len(content.Files))}
	copy(next.Files, content.Files)
	return next
}

func remoteFileSource(content *project.Content, name string) string {
	for _, file := range content.Files {
		if file.Name == name {
			return file.Source
		}
	}
	return ""
}

func assertFileContains(t *testing.T, root, rel, want string) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
	if err != nil {
		t.Fatalf("ReadFile(%s) error = %v", rel, err)
	}
	if !strings.Contains(string(data), want) {
		t.Fatalf("%s does not contain %q: %s", rel, want, string(data))
	}
}
