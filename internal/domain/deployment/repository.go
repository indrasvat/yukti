package deployment

import "context"

// Repository defines the interface for deployment data access.
type Repository interface {
	// List returns all deployments for a project.
	List(ctx context.Context, scriptID string) ([]Deployment, error)

	// Get returns a specific deployment.
	Get(ctx context.Context, scriptID, deploymentID string) (*Deployment, error)

	// Create creates a new deployment.
	Create(ctx context.Context, scriptID string, req CreateRequest) (*Deployment, error)

	// Update updates an existing deployment.
	Update(ctx context.Context, scriptID, deploymentID string, req UpdateRequest) (*Deployment, error)

	// Delete deletes a deployment.
	Delete(ctx context.Context, scriptID, deploymentID string) error
}

// CreateRequest contains parameters for creating a deployment.
type CreateRequest struct {
	VersionNumber int
	Description   string
	Config        *DeploymentConfigRequest
}

// UpdateRequest contains parameters for updating a deployment.
type UpdateRequest struct {
	VersionNumber int
	Description   string
	Config        *DeploymentConfigRequest
}

// DeploymentConfigRequest contains deployment configuration for create/update.
type DeploymentConfigRequest struct {
	Description string
}
