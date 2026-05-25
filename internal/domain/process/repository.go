package process

import (
	"context"
	"time"
)

// Repository defines the interface for process data access.
type Repository interface {
	// List returns processes for a script with optional filters.
	List(ctx context.Context, scriptID string, opts ListOptions) (*ListResult, error)

	// ListUser returns all processes for the current user.
	ListUser(ctx context.Context, opts ListOptions) (*ListResult, error)
}

// ListOptions configures process listing behavior.
type ListOptions struct {
	PageSize      int
	PageToken     string
	FunctionName  string
	Statuses      []Status
	Types         []Type
	StartTime     time.Time
	EndTime       time.Time
	UserAccessKey string
}

// ListResult contains paginated process results.
type ListResult struct {
	Processes     []Process
	NextPageToken string
}
