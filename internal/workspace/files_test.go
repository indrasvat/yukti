package workspace

import (
	"os"
	"path/filepath"
	"testing"

	"yukti/internal/domain/project"
)

func TestRemoteToLocalMapsAppsScriptNames(t *testing.T) {
	t.Parallel()

	content := &project.Content{ScriptID: "script-1", Files: []project.File{
		{Name: "appsscript", Type: project.FileTypeJSON, Source: "{}"},
		{Name: "Code", Type: project.FileTypeServer, Source: "function main() {}"},
		{Name: "views/Dialog", Type: project.FileTypeHTML, Source: "<p>Hello</p>"},
	}}

	files, err := RemoteToLocal(content)
	if err != nil {
		t.Fatalf("RemoteToLocal() error = %v", err)
	}

	got := paths(files)
	want := []string{"Code.gs", "appsscript.json", "views/Dialog.html"}
	if !equalStrings(got, want) {
		t.Fatalf("paths = %v, want %v", got, want)
	}
}

func TestLocalToRemoteScansSupportedFilesAndSkipsArtifacts(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeFile(t, dir, "appsscript.json", "{}")
	writeFile(t, dir, "Code.gs", "function main() {}")
	writeFile(t, dir, "ui/Dialog.html", "<p>Hello</p>")
	writeFile(t, dir, "config.json", `{"ignored": true}`)
	writeFile(t, dir, "ui/node_modules/pkg/ignored.gs", "function ignored() {}")
	writeFile(t, dir, "README.md", "# ignored")
	writeFile(t, dir, ".shux/out/frame.png", "ignored")
	writeFile(t, dir, ManifestName, "{}")

	content, files, err := LocalToRemote(dir)
	if err != nil {
		t.Fatalf("LocalToRemote() error = %v", err)
	}

	gotPaths := paths(files)
	wantPaths := []string{"Code.gs", "appsscript.json", "ui/Dialog.html"}
	if !equalStrings(gotPaths, wantPaths) {
		t.Fatalf("paths = %v, want %v", gotPaths, wantPaths)
	}

	gotRemote := make(map[string]project.FileType)
	for _, file := range content.Files {
		gotRemote[file.Name] = file.Type
	}
	if gotRemote["appsscript"] != project.FileTypeJSON {
		t.Fatalf("appsscript type = %q", gotRemote["appsscript"])
	}
	if gotRemote["ui/Dialog"] != project.FileTypeHTML {
		t.Fatalf("ui/Dialog type = %q", gotRemote["ui/Dialog"])
	}
}

func TestContentHashIsStableAcrossInputOrder(t *testing.T) {
	t.Parallel()

	a := []LocalFile{
		{Path: "B.gs", Name: "B", Type: project.FileTypeServer, Source: "b"},
		{Path: "A.gs", Name: "A", Type: project.FileTypeServer, Source: "a"},
	}
	b := []LocalFile{
		{Path: "A.gs", Name: "A", Type: project.FileTypeServer, Source: "a"},
		{Path: "B.gs", Name: "B", Type: project.FileTypeServer, Source: "b"},
	}

	if ContentHash(a) != ContentHash(b) {
		t.Fatal("ContentHash differs for same files in different order")
	}
}

func TestDiffDetectsAddedModifiedDeletedFiles(t *testing.T) {
	t.Parallel()

	manifest := NewManifest("script-1", "Example", "old", []FileState{
		{Path: "same.gs", Name: "same", Type: string(project.FileTypeServer), Hash: fileHash("same", project.FileTypeServer, "same")},
		{Path: "changed.gs", Name: "changed", Type: string(project.FileTypeServer), Hash: fileHash("changed", project.FileTypeServer, "old")},
		{Path: "removed.gs", Name: "removed", Type: string(project.FileTypeServer), Hash: fileHash("removed", project.FileTypeServer, "gone")},
	})

	changes := Diff([]LocalFile{
		{Path: "same.gs", Name: "same", Type: project.FileTypeServer, Source: "same", Hash: fileHash("same", project.FileTypeServer, "same")},
		{Path: "changed.gs", Name: "changed", Type: project.FileTypeServer, Source: "new", Hash: fileHash("changed", project.FileTypeServer, "new")},
		{Path: "added.gs", Name: "added", Type: project.FileTypeServer, Source: "add", Hash: fileHash("added", project.FileTypeServer, "add")},
	}, &manifest)

	got := map[string]ChangeKind{}
	for _, change := range changes {
		got[change.Path] = change.Kind
	}

	assertChange(t, got, "same.gs", ChangeUnchanged)
	assertChange(t, got, "changed.gs", ChangeModified)
	assertChange(t, got, "added.gs", ChangeAdded)
	assertChange(t, got, "removed.gs", ChangeDeleted)
}

func writeFile(t *testing.T, root, rel, body string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatalf("MkdirAll(%s) error = %v", rel, err)
	}
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatalf("WriteFile(%s) error = %v", rel, err)
	}
}

func paths(files []LocalFile) []string {
	got := make([]string, 0, len(files))
	for _, file := range files {
		got = append(got, file.Path)
	}
	return got
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func assertChange(t *testing.T, got map[string]ChangeKind, path string, want ChangeKind) {
	t.Helper()
	if got[path] != want {
		t.Fatalf("change[%s] = %q, want %q", path, got[path], want)
	}
}
