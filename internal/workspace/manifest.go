// Package workspace maps Google Apps Script projects to local directories.
package workspace

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	// ManifestName is the local metadata file Yukti writes at a workspace root.
	ManifestName = "yukti.json"

	manifestVersion = 1
)

// Manifest records the remote project and the last remote snapshot Yukti saw.
type Manifest struct {
	Version        int                  `json:"version"`
	ScriptID       string               `json:"script_id"`
	Title          string               `json:"title"`
	LastRemoteHash string               `json:"last_remote_hash"`
	LastPulledAt   time.Time            `json:"last_pulled_at"`
	Files          map[string]FileState `json:"files"`
}

// FileState records one local file from the last remote snapshot.
type FileState struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Path string `json:"path"`
	Hash string `json:"hash"`
}

var ErrManifestNotFound = errors.New("yukti workspace manifest not found")

// NewManifest creates metadata for a freshly materialized workspace.
func NewManifest(scriptID, title, remoteHash string, files []FileState) Manifest {
	states := make(map[string]FileState, len(files))
	for _, file := range files {
		states[file.Path] = file
	}

	return Manifest{
		Version:        manifestVersion,
		ScriptID:       scriptID,
		Title:          title,
		LastRemoteHash: remoteHash,
		LastPulledAt:   time.Now().UTC(),
		Files:          states,
	}
}

// LoadManifest loads yukti.json from dir.
func LoadManifest(dir string) (*Manifest, error) {
	path := filepath.Join(dir, ManifestName)
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, ErrManifestNotFound
		}
		return nil, fmt.Errorf("reading %s: %w", ManifestName, err)
	}

	var manifest Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", ManifestName, err)
	}
	if manifest.Version == 0 {
		manifest.Version = manifestVersion
	}
	if manifest.Files == nil {
		manifest.Files = make(map[string]FileState)
	}
	return &manifest, nil
}

// Save writes yukti.json to dir.
func (m Manifest) Save(dir string) error {
	if m.Version == 0 {
		m.Version = manifestVersion
	}
	if m.LastPulledAt.IsZero() {
		m.LastPulledAt = time.Now().UTC()
	}

	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding %s: %w", ManifestName, err)
	}
	data = append(data, '\n')

	path := filepath.Join(dir, ManifestName)
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0o600); err != nil {
		return fmt.Errorf("writing %s: %w", ManifestName, err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("replacing %s: %w", ManifestName, err)
	}
	return nil
}

// FindRoot walks upward from start until it finds yukti.json.
func FindRoot(start string) (string, error) {
	dir, err := filepath.Abs(start)
	if err != nil {
		return "", fmt.Errorf("resolving workspace path: %w", err)
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, ManifestName)); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", ErrManifestNotFound
		}
		dir = parent
	}
}
