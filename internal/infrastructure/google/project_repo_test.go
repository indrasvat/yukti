package google

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"golang.org/x/oauth2"

	"yukti/internal/domain/project"
)

func TestProjectRepositoryCreateSendsTitleAndParent(t *testing.T) {
	t.Parallel()

	var got map[string]string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/projects" {
			t.Fatalf("request = %s %s", r.Method, r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatalf("Decode() error = %v", err)
		}
		_, _ = w.Write([]byte(`{"scriptId":"script-1","title":"Demo","parentId":"sheet-1"}`))
	}))
	defer server.Close()

	repo := NewProjectRepository(testClient(server.URL))
	proj, err := repo.Create(context.Background(), project.CreateRequest{
		Title:    "Demo",
		ParentID: "sheet-1",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if proj.ID != "script-1" {
		t.Fatalf("project ID = %q", proj.ID)
	}
	if got["title"] != "Demo" || got["parentId"] != "sheet-1" {
		t.Fatalf("body = %+v", got)
	}
}

func TestProjectRepositoryUpdateContentSendsFullFileSet(t *testing.T) {
	t.Parallel()

	var got struct {
		Files []struct {
			Name   string `json:"name"`
			Type   string `json:"type"`
			Source string `json:"source"`
		} `json:"files"`
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut || r.URL.Path != "/projects/script-1/content" {
			t.Fatalf("request = %s %s", r.Method, r.URL.Path)
		}
		if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
			t.Fatalf("Content-Type = %q", r.Header.Get("Content-Type"))
		}
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatalf("Decode() error = %v", err)
		}
		_, _ = w.Write([]byte(`{"scriptId":"script-1","files":[]}`))
	}))
	defer server.Close()

	repo := NewProjectRepository(testClient(server.URL))
	err := repo.UpdateContent(context.Background(), "script-1", &project.Content{
		ScriptID: "script-1",
		Files: []project.File{
			{Name: "appsscript", Type: project.FileTypeJSON, Source: "{}"},
			{Name: "Code", Type: project.FileTypeServer, Source: "function main() {}"},
		},
	})
	if err != nil {
		t.Fatalf("UpdateContent() error = %v", err)
	}

	if len(got.Files) != 2 {
		t.Fatalf("files len = %d, want 2", len(got.Files))
	}
	if got.Files[0].Name != "appsscript" || got.Files[0].Type != string(project.FileTypeJSON) {
		t.Fatalf("manifest file = %+v", got.Files[0])
	}
	if got.Files[1].Name != "Code" || got.Files[1].Type != string(project.FileTypeServer) {
		t.Fatalf("code file = %+v", got.Files[1])
	}
}

func testClient(baseURL string) *Client {
	client := NewClient(context.Background(), oauth2.StaticTokenSource(&oauth2.Token{AccessToken: "test"}), slog.Default())
	client.baseURL = baseURL
	return client
}
