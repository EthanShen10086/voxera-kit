// Package experiment provides an A/B testing system with statistical
// significance tracking, extending the featureflag module for user bucketing.
package experiment

import (
	"context"
	"time"
)

// Status represents the lifecycle state of an experiment.
type Status string

const (
	// StatusDraft indicates the experiment is configured but not yet running.
	StatusDraft Status = "draft"
	// StatusRunning indicates the experiment is actively enrolling users.
	StatusRunning Status = "running"
	// StatusPaused indicates the experiment is temporarily stopped.
	StatusPaused Status = "paused"
	// StatusComplete indicates the experiment has concluded.
	StatusComplete Status = "complete"
)

// Config defines an A/B experiment with its variants, metrics, and traffic
// allocation.
type Config struct {
	ID          string
	Key         string
	Name        string
	Description string
	Status      Status
	Variants    []Variant
	Metrics     []MetricDef
	TrafficPct  float64
	StartedAt   time.Time
	EndedAt     time.Time
	CreatedAt   time.Time
}

// Variant represents one arm of an experiment (e.g. control vs treatment).
type Variant struct {
	Key       string
	Name      string
	Weight    int
	IsControl bool
}

// MetricDef defines a metric to track for the experiment.
type MetricDef struct {
	Key       string
	Name      string
	Type      MetricType
	EventName string
	Property  string
}

// MetricType categorizes how a metric is computed.
type MetricType string

const (
	// MetricConversion tracks a binary outcome: did the user convert or not.
	MetricConversion MetricType = "conversion"
	// MetricCount tracks how many times an event occurred per user.
	MetricCount MetricType = "count"
	// MetricRevenue tracks the sum of a numeric property across events.
	MetricRevenue MetricType = "revenue"
	// MetricDuration tracks elapsed time between events.
	MetricDuration MetricType = "duration"
)

// Assignment records which variant a user was assigned to.
type Assignment struct {
	ExperimentKey string
	UserID        string
	VariantKey    string
	AssignedAt    time.Time
}

// MetricResult holds aggregated metric data for one variant of one metric.
type MetricResult struct {
	MetricKey     string
	VariantKey    string
	SampleSize    int64
	Mean          float64
	Variance      float64
	ConvRate      float64
	Confidence    float64
	IsSignificant bool
}

// Result holds the complete experiment results with per-variant breakdowns.
type Result struct {
	ExperimentKey string
	Status        Status
	TotalUsers    int64
	Metrics       []MetricResult
	Winner        string
	StartedAt     time.Time
	AnalyzedAt    time.Time
}

// Manager handles experiment lifecycle, assignment, and result computation.
type Manager interface {
	// Create registers a new experiment definition.
	Create(ctx context.Context, cfg Config) error
	// Get retrieves an experiment by its unique key.
	Get(ctx context.Context, key string) (*Config, error)
	// List returns all experiments matching the given status filter.
	List(ctx context.Context, status Status) ([]Config, error)
	// Start transitions an experiment to the running state.
	Start(ctx context.Context, key string) error
	// Stop ends an experiment and computes final results.
	Stop(ctx context.Context, key string) error
	// Assign assigns a user to a variant deterministically based on user ID.
	Assign(ctx context.Context, key string, userID string) (*Assignment, error)
	// GetAssignment retrieves the current variant assignment for a user.
	GetAssignment(ctx context.Context, key string, userID string) (*Assignment, error)
	// RecordMetric records a metric event for a user in an experiment.
	RecordMetric(ctx context.Context, key string, userID string, metricKey string, value float64) error
	// GetResults computes and returns current experiment results with
	// statistical significance analysis.
	GetResults(ctx context.Context, key string) (*Result, error)
}
