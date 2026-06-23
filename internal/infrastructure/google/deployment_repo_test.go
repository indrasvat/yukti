package google

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"yukti/internal/domain/deployment"
)

func TestDeploymentRepositoryCreateSendsConfig(t *testing.T) {
	t.Parallel()

	var got map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/projects/script-1/deployments" {
			t.Fatalf("request = %s %s", r.Method, r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatalf("Decode() error = %v", err)
		}
		_, _ = w.Write([]byte(`{
			"deploymentId":"dep-1",
			"deploymentConfig":{"scriptId":"script-1","versionNumber":7,"description":"prod"},
			"entryPoints":[{"entryPointType":"WEB_APP","webApp":{"url":"https://script.google.com/macros/s/dep-1/exec","entryPointConfig":{"access":"ANYONE"}}}]
		}`))
	}))
	defer server.Close()

	repo := NewDeploymentRepository(testClient(server.URL))
	dep, err := repo.Create(context.Background(), "script-1", deployment.CreateRequest{
		VersionNumber: 7,
		Description:   "prod",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if got["scriptId"] != "script-1" || got["description"] != "prod" {
		t.Fatalf("body = %+v", got)
	}
	if got["manifestFileName"] != "appsscript" {
		t.Fatalf("manifestFileName = %v", got["manifestFileName"])
	}
	if dep.ID != "dep-1" || dep.Config.VersionID != 7 || dep.WebAppURL() == "" {
		t.Fatalf("deployment = %+v", dep)
	}
}

func TestDeploymentRepositoryUpdateWrapsDeploymentConfig(t *testing.T) {
	t.Parallel()

	var got struct {
		Config struct {
			ScriptID         string `json:"scriptId"`
			VersionNumber    int    `json:"versionNumber"`
			Description      string `json:"description"`
			ManifestFileName string `json:"manifestFileName"`
		} `json:"deploymentConfig"`
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut || r.URL.Path != "/projects/script-1/deployments/dep-1" {
			t.Fatalf("request = %s %s", r.Method, r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatalf("Decode() error = %v", err)
		}
		_, _ = w.Write([]byte(`{"deploymentId":"dep-1","deploymentConfig":{"scriptId":"script-1","versionNumber":8,"description":"prod"}}`))
	}))
	defer server.Close()

	repo := NewDeploymentRepository(testClient(server.URL))
	dep, err := repo.Update(context.Background(), "script-1", "dep-1", deployment.UpdateRequest{
		VersionNumber: 8,
		Description:   "prod",
	})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	if got.Config.ScriptID != "script-1" || got.Config.VersionNumber != 8 || got.Config.ManifestFileName != "appsscript" {
		t.Fatalf("body = %+v", got.Config)
	}
	if dep.Config.VersionID != 8 {
		t.Fatalf("deployment version = %d", dep.Config.VersionID)
	}
}

func TestDeploymentRepositoryGetUsesDeploymentEndpoint(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/projects/script-1/deployments/dep-1" {
			t.Fatalf("request = %s %s", r.Method, r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"deploymentId":"dep-1","deploymentConfig":{"scriptId":"script-1","versionNumber":3,"description":"prod"}}`))
	}))
	defer server.Close()

	repo := NewDeploymentRepository(testClient(server.URL))
	dep, err := repo.Get(context.Background(), "script-1", "dep-1")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if dep.ID != "dep-1" || dep.Config.VersionID != 3 {
		t.Fatalf("deployment = %+v", dep)
	}
}

func TestDeploymentRepositoryDeleteUsesDeleteEndpoint(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/projects/script-1/deployments/dep-1" {
			t.Fatalf("request = %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	repo := NewDeploymentRepository(testClient(server.URL))
	if err := repo.Delete(context.Background(), "script-1", "dep-1"); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
}
