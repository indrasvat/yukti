package cache

import (
	"context"
	"sync"
	"time"

	"yukti/internal/domain/project"
)

// CachingRepository wraps a project.Repository with validation-based caching.
// Uses project UpdateTime for cache validation - if the timestamp hasn't changed,
// the cached content is still valid.
type CachingRepository struct {
	inner project.Repository

	mu       sync.RWMutex
	contents map[string]*cachedContent
}

// cachedContent stores cached project content with its validation timestamp.
type cachedContent struct {
	content    *project.Content
	updateTime time.Time
}

// NewCachingRepository creates a new caching wrapper around a repository.
func NewCachingRepository(inner project.Repository) *CachingRepository {
	return &CachingRepository{
		inner:    inner,
		contents: make(map[string]*cachedContent),
	}
}

// List delegates to the inner repository (not cached).
func (r *CachingRepository) List(ctx context.Context, opts project.ListOptions) (*project.ListResult, error) {
	return r.inner.List(ctx, opts)
}

// Get delegates to the inner repository (not cached, used for validation).
func (r *CachingRepository) Get(ctx context.Context, id string) (*project.Project, error) {
	return r.inner.Get(ctx, id)
}

// GetContent returns cached content if valid, otherwise fetches fresh content.
// Validation: compares cached UpdateTime with current project metadata.
func (r *CachingRepository) GetContent(ctx context.Context, id string) (*project.Content, error) {
	// Check cache first
	r.mu.RLock()
	cached, exists := r.contents[id]
	r.mu.RUnlock()

	if exists {
		// Validate cache by checking project's UpdateTime
		proj, err := r.inner.Get(ctx, id)
		if err == nil && proj.UpdateTime.Equal(cached.updateTime) {
			return cached.content, nil
		}
		// Cache invalid - remove it
		r.mu.Lock()
		delete(r.contents, id)
		r.mu.Unlock()
	}

	// Fetch fresh content
	content, err := r.inner.GetContent(ctx, id)
	if err != nil {
		return nil, err
	}

	// Get project metadata for UpdateTime to enable caching
	// If this fails, we still return content but skip caching
	proj, getErr := r.inner.Get(ctx, id)
	if getErr == nil {
		r.mu.Lock()
		r.contents[id] = &cachedContent{
			content:    content,
			updateTime: proj.UpdateTime,
		}
		r.mu.Unlock()
	}

	return content, nil
}

// GetMetrics delegates to the inner repository (not cached).
func (r *CachingRepository) GetMetrics(ctx context.Context, id string, opts project.MetricsOptions) (*project.Metrics, error) {
	return r.inner.GetMetrics(ctx, id, opts)
}

// Create delegates to the inner repository and invalidates any cache.
func (r *CachingRepository) Create(ctx context.Context, req project.CreateRequest) (*project.Project, error) {
	return r.inner.Create(ctx, req)
}

// UpdateContent delegates to the inner repository and invalidates cache.
func (r *CachingRepository) UpdateContent(ctx context.Context, id string, content *project.Content) error {
	err := r.inner.UpdateContent(ctx, id, content)
	if err == nil {
		// Invalidate cache on successful update
		r.mu.Lock()
		delete(r.contents, id)
		r.mu.Unlock()
	}
	return err
}

// Invalidate removes a specific project from the cache.
func (r *CachingRepository) Invalidate(id string) {
	r.mu.Lock()
	delete(r.contents, id)
	r.mu.Unlock()
}

// InvalidateAll clears all cached content.
func (r *CachingRepository) InvalidateAll() {
	r.mu.Lock()
	r.contents = make(map[string]*cachedContent)
	r.mu.Unlock()
}

// Stats returns cache statistics.
func (r *CachingRepository) Stats() CacheStats {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return CacheStats{
		CachedProjects: len(r.contents),
	}
}

// CacheStats contains cache statistics.
type CacheStats struct {
	CachedProjects int
}
