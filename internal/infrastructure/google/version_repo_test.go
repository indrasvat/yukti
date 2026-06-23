package google

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"yukti/internal/domain/version"
)

func TestVersionRepositoryCreateSendsDescription(t *testing.T) {
	t.Parallel()

	var got map[string]string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/projects/script-1/versions" {
			t.Fatalf("request = %s %s", r.Method, r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatalf("Decode() error = %v", err)
		}
		_, _ = w.Write([]byte(`{"scriptId":"script-1","versionNumber":7,"description":"ship it","createTime":"2026-06-23T02:00:00Z"}`))
	}))
	defer server.Close()

	repo := NewVersionRepository(testClient(server.URL))
	ver, err := repo.Create(context.Background(), "script-1", version.CreateRequest{Description: "ship it"})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if got["description"] != "ship it" {
		t.Fatalf("body = %+v", got)
	}
	if ver.VersionNumber != 7 || ver.ScriptID != "script-1" {
		t.Fatalf("version = %+v", ver)
	}
}

func TestVersionRepositoryListPaginates(t *testing.T) {
	t.Parallel()

	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if r.Method != http.MethodGet || r.URL.Path != "/projects/script-1/versions" {
			t.Fatalf("request = %s %s", r.Method, r.URL.Path)
		}
		switch r.URL.Query().Get("pageToken") {
		case "":
			_, _ = w.Write([]byte(`{"versions":[{"scriptId":"script-1","versionNumber":1,"description":"one"}],"nextPageToken":"next"}`))
		case "next":
			_, _ = w.Write([]byte(`{"versions":[{"scriptId":"script-1","versionNumber":2,"description":"two"}]}`))
		default:
			t.Fatalf("pageToken = %q", r.URL.Query().Get("pageToken"))
		}
	}))
	defer server.Close()

	repo := NewVersionRepository(testClient(server.URL))
	versions, err := repo.List(context.Background(), "script-1")
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if requests != 2 {
		t.Fatalf("requests = %d, want 2", requests)
	}
	if len(versions) != 2 || versions[1].VersionNumber != 2 {
		t.Fatalf("versions = %+v", versions)
	}
}
