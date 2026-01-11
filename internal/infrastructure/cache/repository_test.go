package cache

import (
	"context"
	"testing"
	"time"

	"yukti/internal/domain/project"
)

// mockRepository is a test double for project.Repository.
type mockRepository struct {
	projects   map[string]*project.Project
	contents   map[string]*project.Content
	getCalls   int
	getContent int
}

func newMockRepository() *mockRepository {
	return &mockRepository{
		projects: make(map[string]*project.Project),
		contents: make(map[string]*project.Content),
	}
}

func (m *mockRepository) List(_ context.Context, _ project.ListOptions) (*project.ListResult, error) {
	return &project.ListResult{}, nil
}

func (m *mockRepository) Get(_ context.Context, id string) (*project.Project, error) {
	m.getCalls++
	if p, ok := m.projects[id]; ok {
		return p, nil
	}
	return nil, nil
}

func (m *mockRepository) GetContent(_ context.Context, id string) (*project.Content, error) {
	m.getContent++
	if c, ok := m.contents[id]; ok {
		return c, nil
	}
	return nil, nil
}

func (m *mockRepository) GetMetrics(_ context.Context, _ string, _ project.MetricsOptions) (*project.Metrics, error) {
	return nil, nil
}

func (m *mockRepository) Create(_ context.Context, _ project.CreateRequest) (*project.Project, error) {
	return nil, nil
}

func (m *mockRepository) UpdateContent(_ context.Context, _ string, _ *project.Content) error {
	return nil
}

func TestCachingRepository_CachesContent(t *testing.T) {
	mock := newMockRepository()
	updateTime := time.Now()

	mock.projects["proj1"] = &project.Project{
		ID:         "proj1",
		UpdateTime: updateTime,
	}
	mock.contents["proj1"] = &project.Content{
		ScriptID: "proj1",
		Files: []project.File{
			{Name: "Code", Type: project.FileTypeServer, Source: "function test() {}"},
		},
	}

	cache := NewCachingRepository(mock)
	ctx := context.Background()

	// First call - should fetch from inner
	content1, err := cache.GetContent(ctx, "proj1")
	if err != nil {
		t.Fatalf("GetContent() error = %v", err)
	}
	if content1 == nil {
		t.Fatal("GetContent() returned nil")
	}
	if mock.getContent != 1 {
		t.Errorf("Expected 1 GetContent call, got %d", mock.getContent)
	}

	// Second call - should use cache
	content2, err := cache.GetContent(ctx, "proj1")
	if err != nil {
		t.Fatalf("GetContent() error = %v", err)
	}
	if content2 == nil {
		t.Fatal("GetContent() returned nil")
	}
	// Should still be 1 - using cache
	if mock.getContent != 1 {
		t.Errorf("Expected 1 GetContent call (cached), got %d", mock.getContent)
	}

	// Verify content is same
	if content1.ScriptID != content2.ScriptID {
		t.Error("Cached content should match original")
	}
}

func TestCachingRepository_InvalidatesOnUpdate(t *testing.T) {
	mock := newMockRepository()
	updateTime := time.Now()

	mock.projects["proj1"] = &project.Project{
		ID:         "proj1",
		UpdateTime: updateTime,
	}
	mock.contents["proj1"] = &project.Content{
		ScriptID: "proj1",
		Files:    []project.File{{Name: "Code", Source: "v1"}},
	}

	cache := NewCachingRepository(mock)
	ctx := context.Background()

	// Populate cache
	_, _ = cache.GetContent(ctx, "proj1")
	if mock.getContent != 1 {
		t.Errorf("Expected 1 GetContent call, got %d", mock.getContent)
	}

	// Simulate project update - change UpdateTime
	mock.projects["proj1"].UpdateTime = updateTime.Add(time.Hour)
	mock.contents["proj1"].Files[0].Source = "v2"

	// Next call should detect stale cache and refetch
	content, err := cache.GetContent(ctx, "proj1")
	if err != nil {
		t.Fatalf("GetContent() error = %v", err)
	}
	if mock.getContent != 2 {
		t.Errorf("Expected 2 GetContent calls (invalidated), got %d", mock.getContent)
	}
	if content.Files[0].Source != "v2" {
		t.Error("Should have fetched updated content")
	}
}

func TestCachingRepository_ManualInvalidate(t *testing.T) {
	mock := newMockRepository()
	updateTime := time.Now()

	mock.projects["proj1"] = &project.Project{
		ID:         "proj1",
		UpdateTime: updateTime,
	}
	mock.contents["proj1"] = &project.Content{ScriptID: "proj1"}

	cache := NewCachingRepository(mock)
	ctx := context.Background()

	// Populate cache
	_, _ = cache.GetContent(ctx, "proj1")

	// Manually invalidate
	cache.Invalidate("proj1")

	// Should refetch
	_, _ = cache.GetContent(ctx, "proj1")
	if mock.getContent != 2 {
		t.Errorf("Expected 2 GetContent calls after Invalidate, got %d", mock.getContent)
	}
}

func TestCachingRepository_InvalidateAll(t *testing.T) {
	mock := newMockRepository()
	updateTime := time.Now()

	for _, id := range []string{"proj1", "proj2"} {
		mock.projects[id] = &project.Project{ID: id, UpdateTime: updateTime}
		mock.contents[id] = &project.Content{ScriptID: id}
	}

	cache := NewCachingRepository(mock)
	ctx := context.Background()

	// Populate cache
	_, _ = cache.GetContent(ctx, "proj1")
	_, _ = cache.GetContent(ctx, "proj2")

	stats := cache.Stats()
	if stats.CachedProjects != 2 {
		t.Errorf("Expected 2 cached projects, got %d", stats.CachedProjects)
	}

	// Invalidate all
	cache.InvalidateAll()

	stats = cache.Stats()
	if stats.CachedProjects != 0 {
		t.Errorf("Expected 0 cached projects after InvalidateAll, got %d", stats.CachedProjects)
	}
}

func TestCachingRepository_UpdateContentInvalidates(t *testing.T) {
	mock := newMockRepository()
	updateTime := time.Now()

	mock.projects["proj1"] = &project.Project{
		ID:         "proj1",
		UpdateTime: updateTime,
	}
	mock.contents["proj1"] = &project.Content{ScriptID: "proj1"}

	cache := NewCachingRepository(mock)
	ctx := context.Background()

	// Populate cache
	_, _ = cache.GetContent(ctx, "proj1")

	stats := cache.Stats()
	if stats.CachedProjects != 1 {
		t.Errorf("Expected 1 cached project, got %d", stats.CachedProjects)
	}

	// UpdateContent should invalidate cache
	_ = cache.UpdateContent(ctx, "proj1", &project.Content{})

	stats = cache.Stats()
	if stats.CachedProjects != 0 {
		t.Errorf("Expected 0 cached projects after UpdateContent, got %d", stats.CachedProjects)
	}
}
