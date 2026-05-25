package version

import "context"

// Repository defines the interface for version data access.
type Repository interface {
	// List returns all versions for a project.
	List(ctx context.Context, scriptID string) ([]Version, error)

	// Get returns a specific version.
	Get(ctx context.Context, scriptID string, versionNumber int) (*Version, error)

	// Create creates a new version (snapshot of current code).
	Create(ctx context.Context, scriptID string, req CreateRequest) (*Version, error)
}

// CreateRequest contains parameters for creating a version.
type CreateRequest struct {
	Description string
}
