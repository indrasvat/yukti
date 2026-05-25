package workspace

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"yukti/internal/domain/project"
)

// Service coordinates local workspace operations with the Apps Script API.
type Service struct {
	repo project.Repository
}

// NewService creates a workspace sync service.
func NewService(repo project.Repository) *Service {
	return &Service{repo: repo}
}

// CreateOptions configures creating a new remote and local workspace.
type CreateOptions struct {
	Title    string
	Dir      string
	ParentID string
	Force    bool
}

// CloneOptions configures cloning an existing remote project.
type CloneOptions struct {
	ScriptID string
	Dir      string
	Force    bool
}

// PullOptions configures pulling remote content into an existing workspace.
type PullOptions struct {
	Dir   string
	Force bool
}

// PushOptions configures pushing local content to remote HEAD.
type PushOptions struct {
	Dir   string
	Force bool
}

// Result summarizes a sync operation.
type Result struct {
	ScriptID   string
	Title      string
	Dir        string
	RemoteHash string
	Changes    []Change
	Files      []LocalFile
}

// Create creates a new Apps Script project, writes starter files, and pushes them.
func (s *Service) Create(ctx context.Context, opts CreateOptions) (*Result, error) {
	if opts.Title == "" {
		return nil, errors.New("title is required")
	}
	dir := opts.Dir
	if dir == "" {
		dir = slug(opts.Title)
	}
	if err := prepareDir(dir, opts.Force); err != nil {
		return nil, err
	}

	proj, err := s.repo.Create(ctx, project.CreateRequest{Title: opts.Title, ParentID: opts.ParentID})
	if err != nil {
		return nil, fmt.Errorf("creating Apps Script project: %w", err)
	}

	files := DefaultFiles(opts.Title)
	if updateErr := s.repo.UpdateContent(ctx, proj.ID, &project.Content{ScriptID: proj.ID, Files: files}); updateErr != nil {
		return nil, fmt.Errorf("uploading starter files: %w", updateErr)
	}

	content, err := s.repo.GetContent(ctx, proj.ID)
	if err != nil {
		return nil, fmt.Errorf("fetching created project content: %w", err)
	}

	return materialize(dir, proj.ID, proj.Title, content)
}

// Clone writes a remote project into a new local workspace.
func (s *Service) Clone(ctx context.Context, opts CloneOptions) (*Result, error) {
	if opts.ScriptID == "" {
		return nil, errors.New("script ID is required")
	}
	dir := opts.Dir
	if dir == "" {
		dir = opts.ScriptID
	}
	if err := prepareDir(dir, opts.Force); err != nil {
		return nil, err
	}

	proj, err := s.repo.Get(ctx, opts.ScriptID)
	if err != nil {
		return nil, fmt.Errorf("getting project metadata: %w", err)
	}
	content, err := s.repo.GetContent(ctx, opts.ScriptID)
	if err != nil {
		return nil, fmt.Errorf("getting project content: %w", err)
	}

	return materialize(dir, proj.ID, proj.Title, content)
}

// Pull refreshes local files from remote HEAD.
func (s *Service) Pull(ctx context.Context, opts PullOptions) (*Result, error) {
	root, manifest, err := loadWorkspace(opts.Dir)
	if err != nil {
		return nil, err
	}

	_, localFiles, err := LocalToRemote(root)
	if err != nil {
		return nil, err
	}
	changes := Diff(localFiles, manifest)
	if Dirty(changes) && !opts.Force {
		return nil, fmt.Errorf("local workspace has unpushed changes (%s); rerun with --force to overwrite", Summary(changes))
	}

	content, err := s.repo.GetContent(ctx, manifest.ScriptID)
	if err != nil {
		return nil, fmt.Errorf("getting remote content: %w", err)
	}
	files, err := RemoteToLocal(content)
	if err != nil {
		return nil, err
	}
	if err := removeMissingTrackedFiles(root, manifest, files); err != nil {
		return nil, err
	}
	return materializeFiles(root, manifest.ScriptID, manifest.Title, files)
}

// Push uploads local files to remote HEAD after checking the remote snapshot.
func (s *Service) Push(ctx context.Context, opts PushOptions) (*Result, error) {
	root, manifest, err := loadWorkspace(opts.Dir)
	if err != nil {
		return nil, err
	}

	remoteContent, err := s.repo.GetContent(ctx, manifest.ScriptID)
	if err != nil {
		return nil, fmt.Errorf("getting remote content: %w", err)
	}
	remoteFiles, err := RemoteToLocal(remoteContent)
	if err != nil {
		return nil, err
	}
	remoteHash := ContentHash(remoteFiles)
	if remoteHash != manifest.LastRemoteHash && !opts.Force {
		return nil, fmt.Errorf("remote HEAD changed since last pull; run yukti pull or rerun push with --force")
	}

	content, localFiles, err := LocalToRemote(root)
	if err != nil {
		return nil, err
	}
	content.ScriptID = manifest.ScriptID

	if len(content.Files) == 0 {
		return nil, errors.New("no Apps Script files found to push")
	}

	if err := ensureManifestFile(content); err != nil {
		return nil, err
	}

	changes := Diff(localFiles, manifest)
	if err := s.repo.UpdateContent(ctx, manifest.ScriptID, content); err != nil {
		return nil, fmt.Errorf("updating remote content: %w", err)
	}

	newHash := ContentHash(localFiles)
	nextManifest := NewManifest(manifest.ScriptID, manifest.Title, newHash, fileStates(localFiles))
	if err := nextManifest.Save(root); err != nil {
		return nil, err
	}

	return &Result{
		ScriptID:   manifest.ScriptID,
		Title:      manifest.Title,
		Dir:        root,
		RemoteHash: newHash,
		Changes:    changes,
		Files:      localFiles,
	}, nil
}

// Status compares local files against the last pulled remote snapshot.
func (s *Service) Status(dir string) (*Result, error) {
	root, manifest, err := loadWorkspace(dir)
	if err != nil {
		return nil, err
	}
	_, localFiles, err := LocalToRemote(root)
	if err != nil {
		return nil, err
	}
	return &Result{
		ScriptID:   manifest.ScriptID,
		Title:      manifest.Title,
		Dir:        root,
		RemoteHash: manifest.LastRemoteHash,
		Changes:    Diff(localFiles, manifest),
		Files:      localFiles,
	}, nil
}

func materialize(dir, scriptID, title string, content *project.Content) (*Result, error) {
	files, err := RemoteToLocal(content)
	if err != nil {
		return nil, err
	}
	return materializeFiles(dir, scriptID, title, files)
}

func materializeFiles(dir, scriptID, title string, files []LocalFile) (*Result, error) {
	if err := WriteLocalFiles(dir, files); err != nil {
		return nil, err
	}
	hash := ContentHash(files)
	manifest := NewManifest(scriptID, title, hash, fileStates(files))
	if err := manifest.Save(dir); err != nil {
		return nil, err
	}
	return &Result{
		ScriptID:   scriptID,
		Title:      title,
		Dir:        dir,
		RemoteHash: hash,
		Files:      files,
	}, nil
}

func removeMissingTrackedFiles(root string, manifest *Manifest, nextFiles []LocalFile) error {
	next := make(map[string]struct{}, len(nextFiles))
	for _, file := range nextFiles {
		next[file.Path] = struct{}{}
	}

	for path := range manifest.Files {
		if _, ok := next[path]; ok {
			continue
		}
		if err := os.Remove(filepath.Join(root, filepath.FromSlash(path))); err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("removing stale tracked file %s: %w", path, err)
		}
	}
	return nil
}

func loadWorkspace(dir string) (string, *Manifest, error) {
	if dir == "" {
		dir = "."
	}
	root, err := FindRoot(dir)
	if err != nil {
		return "", nil, err
	}
	manifest, err := LoadManifest(root)
	if err != nil {
		return "", nil, err
	}
	if manifest.ScriptID == "" {
		return "", nil, errors.New("workspace manifest is missing script_id")
	}
	return root, manifest, nil
}

func prepareDir(dir string, force bool) error {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("resolving target directory: %w", err)
	}
	entries, err := os.ReadDir(abs)
	if err == nil && len(entries) > 0 && !force {
		return fmt.Errorf("target directory %s is not empty; rerun with --force to use it", abs)
	}
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("checking target directory: %w", err)
	}
	if err := os.MkdirAll(abs, 0o700); err != nil {
		return fmt.Errorf("creating target directory: %w", err)
	}
	return nil
}

func ensureManifestFile(content *project.Content) error {
	for _, file := range content.Files {
		if file.Name == "appsscript" && file.Type == project.FileTypeJSON {
			return nil
		}
	}
	return errors.New("appsscript.json is required before push")
}

func slug(title string) string {
	lower := strings.ToLower(title)
	var b strings.Builder
	lastDash := false
	for _, r := range lower {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			lastDash = false
			continue
		}
		if !lastDash {
			b.WriteByte('-')
			lastDash = true
		}
	}
	out := strings.Trim(b.String(), "-")
	if out == "" {
		return "yukti-script"
	}
	return out
}
