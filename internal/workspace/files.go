package workspace

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"yukti/internal/domain/project"
)

const (
	serverExt = ".gs"
	htmlExt   = ".html"
	jsonExt   = ".json"
)

// LocalFile is a local representation of an Apps Script file.
type LocalFile struct {
	Path   string
	Name   string
	Type   project.FileType
	Source string
	Hash   string
}

// DefaultFiles returns starter files for a new standalone script.
func DefaultFiles(title string) []project.File {
	manifest := `{
  "timeZone": "America/Los_Angeles",
  "dependencies": {},
  "exceptionLogging": "STACKDRIVER",
  "runtimeVersion": "V8"
}`

	code := fmt.Sprintf(`function main() {
  console.log("Hello from %s");
}
`, strings.ReplaceAll(title, `"`, `\"`))

	return []project.File{
		{Name: "appsscript", Type: project.FileTypeJSON, Source: manifest},
		{Name: "Code", Type: project.FileTypeServer, Source: code},
	}
}

// RemoteToLocal converts API content into deterministic local files.
func RemoteToLocal(content *project.Content) ([]LocalFile, error) {
	if content == nil {
		return nil, errors.New("remote content is nil")
	}

	files := make([]LocalFile, 0, len(content.Files))
	for _, remote := range content.Files {
		localPath, err := remotePath(remote)
		if err != nil {
			return nil, err
		}
		files = append(files, LocalFile{
			Path:   localPath,
			Name:   remote.Name,
			Type:   remote.Type,
			Source: remote.Source,
			Hash:   fileHash(remote.Name, remote.Type, remote.Source),
		})
	}
	sortLocalFiles(files)
	return files, nil
}

// LocalToRemote scans dir and converts supported local files into API content.
func LocalToRemote(dir string) (*project.Content, []LocalFile, error) {
	files := make([]LocalFile, 0)

	err := filepath.WalkDir(dir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)

		if entry.IsDir() {
			if shouldSkipDir(rel) {
				return filepath.SkipDir
			}
			return nil
		}
		if shouldSkipFile(rel) {
			return nil
		}

		name, typ, ok := localNameAndType(rel)
		if !ok {
			return nil
		}

		data, err := os.ReadFile(path) //nolint:gosec // Workspace sync intentionally reads files discovered under the workspace root.
		if err != nil {
			return fmt.Errorf("reading %s: %w", rel, err)
		}

		source := string(data)
		files = append(files, LocalFile{
			Path:   rel,
			Name:   name,
			Type:   typ,
			Source: source,
			Hash:   fileHash(name, typ, source),
		})
		return nil
	})
	if err != nil {
		return nil, nil, fmt.Errorf("scanning workspace: %w", err)
	}

	sortLocalFiles(files)

	remote := &project.Content{Files: make([]project.File, 0, len(files))}
	for _, file := range files {
		remote.Files = append(remote.Files, project.File{
			Name:   file.Name,
			Type:   file.Type,
			Source: file.Source,
		})
	}
	return remote, files, nil
}

// WriteLocalFiles writes files to dir, creating parent directories.
func WriteLocalFiles(dir string, files []LocalFile) error {
	for _, file := range files {
		path := filepath.Join(dir, filepath.FromSlash(file.Path))
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			return fmt.Errorf("creating directory for %s: %w", file.Path, err)
		}
		if err := os.WriteFile(path, []byte(file.Source), 0o600); err != nil {
			return fmt.Errorf("writing %s: %w", file.Path, err)
		}
	}
	return nil
}

// ContentHash returns a stable hash for a complete project snapshot.
func ContentHash(files []LocalFile) string {
	copyFiles := append([]LocalFile(nil), files...)
	sortLocalFiles(copyFiles)

	hasher := sha256.New()
	for _, file := range copyFiles {
		hasher.Write([]byte(file.Name))
		hasher.Write([]byte{0})
		hasher.Write([]byte(file.Type))
		hasher.Write([]byte{0})
		hasher.Write([]byte(file.Source))
		hasher.Write([]byte{0})
	}
	return hex.EncodeToString(hasher.Sum(nil))
}

func remotePath(file project.File) (string, error) {
	switch file.Type {
	case project.FileTypeServer:
		return file.Name + serverExt, nil
	case project.FileTypeHTML:
		return file.Name + htmlExt, nil
	case project.FileTypeJSON:
		if file.Name == "appsscript" {
			return "appsscript.json", nil
		}
		return file.Name + jsonExt, nil
	default:
		return "", fmt.Errorf("unsupported Apps Script file type %q for %s", file.Type, file.Name)
	}
}

func localNameAndType(path string) (string, project.FileType, bool) {
	if path == "appsscript.json" {
		return "appsscript", project.FileTypeJSON, true
	}

	ext := filepath.Ext(path)
	name := strings.TrimSuffix(path, ext)
	switch ext {
	case serverExt:
		return name, project.FileTypeServer, true
	case htmlExt:
		return name, project.FileTypeHTML, true
	case jsonExt:
		return "", "", false
	default:
		return "", "", false
	}
}

func shouldSkipDir(path string) bool {
	switch filepath.Base(path) {
	case ".git", ".shux", ".claude", ".local", "bin", "dist", "coverage", "node_modules", "tmp":
		return true
	default:
		return false
	}
}

func shouldSkipFile(path string) bool {
	base := filepath.Base(path)
	return base == ManifestName || base == ".DS_Store" || strings.HasSuffix(base, "~")
}

func fileHash(name string, typ project.FileType, source string) string {
	hasher := sha256.New()
	hasher.Write([]byte(name))
	hasher.Write([]byte{0})
	hasher.Write([]byte(typ))
	hasher.Write([]byte{0})
	hasher.Write([]byte(source))
	return hex.EncodeToString(hasher.Sum(nil))
}

func sortLocalFiles(files []LocalFile) {
	sort.Slice(files, func(i, j int) bool {
		return files[i].Path < files[j].Path
	})
}

func fileStates(files []LocalFile) []FileState {
	states := make([]FileState, 0, len(files))
	for _, file := range files {
		states = append(states, FileState{
			Name: file.Name,
			Type: string(file.Type),
			Path: file.Path,
			Hash: file.Hash,
		})
	}
	return states
}
