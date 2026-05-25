package project

import (
	"context"
	"time"
)

// Repository defines the interface for project data access.
type Repository interface {
	// Queries
	List(ctx context.Context, opts ListOptions) (*ListResult, error)
	Get(ctx context.Context, id string) (*Project, error)
	GetContent(ctx context.Context, id string) (*Content, error)
	GetMetrics(ctx context.Context, id string, opts MetricsOptions) (*Metrics, error)

	// Commands
	Create(ctx context.Context, req CreateRequest) (*Project, error)
	UpdateContent(ctx context.Context, id string, content *Content) error
}

// ListOptions configures project listing behavior.
type ListOptions struct {
	PageSize  int
	PageToken string
}

// ListResult contains paginated project results.
type ListResult struct {
	Projects      []Project
	NextPageToken string
}

// CreateRequest contains parameters for creating a new project.
type CreateRequest struct {
	Title    string
	ParentID string // Optional, for bound scripts
}

// MetricsOptions configures metrics retrieval.
type MetricsOptions struct {
	StartTime time.Time
	EndTime   time.Time
}

// Metrics contains project usage metrics.
type Metrics struct {
	ActiveUsers   []MetricValue
	TotalUsers    []MetricValue
	Executions    []MetricValue
	FailedPercent []MetricValue
}

// MetricValue represents a single metric data point.
type MetricValue struct {
	Time  time.Time
	Value int64
}
