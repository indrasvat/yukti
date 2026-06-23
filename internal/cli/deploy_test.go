package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"yukti/internal/domain/deployment"
	googleinfra "yukti/internal/infrastructure/google"
	"yukti/internal/workspace"
)

func TestResolveReleaseWorkspaceAllowsNoPushExplicitScriptFromAnotherWorkspace(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	manifest := workspace.NewManifest("script-a", "Script A", "hash-a", nil)
	if err := manifest.Save(dir); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	root, resolved, err := resolveReleaseWorkspace("script-b", dir, true)
	if err != nil {
		t.Fatalf("resolveReleaseWorkspace() error = %v", err)
	}
	if root != "" {
		t.Fatalf("root = %q, want no workspace root", root)
	}
	if resolved.ScriptID != "script-b" {
		t.Fatalf("ScriptID = %q, want explicit script", resolved.ScriptID)
	}
}

func TestResolveReleaseWorkspaceRejectsExplicitMismatchWhenPushing(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	manifest := workspace.NewManifest("script-a", "Script A", "hash-a", nil)
	if err := manifest.Save(dir); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	_, _, err := resolveReleaseWorkspace("script-b", dir, false)
	if err == nil || !strings.Contains(err.Error(), "does not match workspace") {
		t.Fatalf("resolveReleaseWorkspace() error = %v, want workspace mismatch", err)
	}
}

func TestSaveWorkspaceDeploymentIDReloadsFreshManifest(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	latest := workspace.NewManifest("script-a", "Script A", "fresh-hash", []workspace.FileState{
		{Name: "Code", Type: "SERVER_JS", Path: "Code.gs", Hash: "fresh-file-hash"},
	})
	if err := latest.Save(dir); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	if err := saveWorkspaceDeploymentID(dir, "dep-new"); err != nil {
		t.Fatalf("saveWorkspaceDeploymentID() error = %v", err)
	}

	manifest, err := workspace.LoadManifest(dir)
	if err != nil {
		t.Fatalf("LoadManifest() error = %v", err)
	}
	if manifest.LastRemoteHash != "fresh-hash" {
		t.Fatalf("LastRemoteHash = %q, want fresh hash", manifest.LastRemoteHash)
	}
	if manifest.Files["Code.gs"].Hash != "fresh-file-hash" {
		t.Fatalf("Code.gs hash = %q, want fresh file hash", manifest.Files["Code.gs"].Hash)
	}
	if manifest.DeploymentID != "dep-new" {
		t.Fatalf("DeploymentID = %q, want dep-new", manifest.DeploymentID)
	}
}

func TestResolveReleaseDeploymentIDClearsStaleManifestDeployment(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	manifest := workspace.NewManifest("script-a", "Script A", "hash-a", nil)
	manifest.DeploymentID = "dep-stale"
	if err := manifest.Save(dir); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	repo := &releaseDeploymentRepo{getErr: fmt.Errorf("lookup failed: %w", googleinfra.ErrNotFound)}

	targetID, err := resolveReleaseDeploymentID(context.Background(), repo, "script-a", dir, "", "dep-stale")
	if err != nil {
		t.Fatalf("resolveReleaseDeploymentID() error = %v", err)
	}
	if targetID != "" {
		t.Fatalf("targetID = %q, want create-new path", targetID)
	}
	loadedManifest, err := workspace.LoadManifest(dir)
	if err != nil {
		t.Fatalf("LoadManifest() error = %v", err)
	}
	if loadedManifest.DeploymentID != "" {
		t.Fatalf("DeploymentID = %q, want cleared stale ID", loadedManifest.DeploymentID)
	}
}

func TestResolveReleaseDeploymentIDDoesNotClearExplicitMissingDeployment(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	manifest := workspace.NewManifest("script-a", "Script A", "hash-a", nil)
	manifest.DeploymentID = "dep-saved"
	if err := manifest.Save(dir); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	repo := &releaseDeploymentRepo{getErr: fmt.Errorf("lookup failed: %w", googleinfra.ErrNotFound)}

	_, err := resolveReleaseDeploymentID(context.Background(), repo, "script-a", dir, "dep-explicit", "dep-saved")
	if err == nil || !strings.Contains(err.Error(), "validating deployment") {
		t.Fatalf("resolveReleaseDeploymentID() error = %v, want validation error", err)
	}
	loadedManifest, loadErr := workspace.LoadManifest(dir)
	if loadErr != nil {
		t.Fatalf("LoadManifest() error = %v", loadErr)
	}
	if loadedManifest.DeploymentID != "dep-saved" {
		t.Fatalf("DeploymentID = %q, want explicit failure to preserve saved ID", loadedManifest.DeploymentID)
	}
}

func TestResolveReleaseDeploymentIDRejectsHeadDeployment(t *testing.T) {
	t.Parallel()

	repo := &releaseDeploymentRepo{deployment: &deployment.Deployment{
		ID:     "head",
		Config: deployment.Config{VersionID: 0},
	}}

	_, err := resolveReleaseDeploymentID(context.Background(), repo, "script-a", "", "head", "")
	if err == nil || !strings.Contains(err.Error(), "automatic HEAD deployment") {
		t.Fatalf("resolveReleaseDeploymentID() error = %v, want HEAD rejection", err)
	}
}

type releaseDeploymentRepo struct {
	deployment *deployment.Deployment
	getErr     error
}

func (r *releaseDeploymentRepo) List(context.Context, string) ([]deployment.Deployment, error) {
	return nil, nil
}

func (r *releaseDeploymentRepo) Get(_ context.Context, _, _ string) (*deployment.Deployment, error) {
	if r.getErr != nil {
		return nil, r.getErr
	}
	if r.deployment != nil {
		copyDeployment := *r.deployment
		return &copyDeployment, nil
	}
	return &deployment.Deployment{ID: "dep-ok", Config: deployment.Config{VersionID: 1}}, nil
}

func (r *releaseDeploymentRepo) Create(context.Context, string, deployment.CreateRequest) (*deployment.Deployment, error) {
	return nil, nil
}

func (r *releaseDeploymentRepo) Update(context.Context, string, string, deployment.UpdateRequest) (*deployment.Deployment, error) {
	return nil, nil
}

func (r *releaseDeploymentRepo) Delete(context.Context, string, string) error {
	return nil
}

func TestResolveReleaseWorkspaceFindsAncestorManifest(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	manifest := workspace.NewManifest("script-a", "Script A", "hash-a", nil)
	if err := manifest.Save(root); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	child := filepath.Join(root, "nested")
	if err := os.MkdirAll(child, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	resolvedRoot, resolved, err := resolveReleaseWorkspace("", child, false)
	if err != nil {
		t.Fatalf("resolveReleaseWorkspace() error = %v", err)
	}
	if resolvedRoot != root {
		t.Fatalf("root = %q, want %q", resolvedRoot, root)
	}
	if resolved.ScriptID != "script-a" {
		t.Fatalf("ScriptID = %q, want script-a", resolved.ScriptID)
	}
}
