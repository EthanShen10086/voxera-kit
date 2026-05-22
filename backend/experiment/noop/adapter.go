// Package noop provides a no-op implementation of the experiment Manager
// interface that discards all operations and always assigns the control variant.
package noop

import (
	"context"
	"time"

	"github.com/EthanShen10086/voxera-kit/experiment"
)

// Adapter is a Manager that performs no real experiment tracking. All write
// operations are silently discarded, and Assign always returns the control
// variant.
type Adapter struct{}

// NewAdapter creates a new no-op experiment adapter.
func NewAdapter() *Adapter {
	return &Adapter{}
}

// Create discards the config and returns nil.
func (a *Adapter) Create(_ context.Context, _ experiment.Config) error {
	return nil
}

// Get always returns nil with no error, indicating no experiment was found.
func (a *Adapter) Get(_ context.Context, _ string) (*experiment.Config, error) {
	return nil, nil //nolint:nilnil // no-op adapter intentionally returns nil
}

// List returns an empty slice.
func (a *Adapter) List(_ context.Context, _ experiment.Status) ([]experiment.Config, error) {
	return nil, nil
}

// Start is a no-op.
func (a *Adapter) Start(_ context.Context, _ string) error {
	return nil
}

// Stop is a no-op.
func (a *Adapter) Stop(_ context.Context, _ string) error {
	return nil
}

// Assign always returns an assignment to the "control" variant.
func (a *Adapter) Assign(_ context.Context, key string, userID string) (*experiment.Assignment, error) {
	return &experiment.Assignment{
		ExperimentKey: key,
		UserID:        userID,
		VariantKey:    "control",
		AssignedAt:    time.Now(),
	}, nil
}

// GetAssignment always returns nil, indicating no prior assignment exists.
func (a *Adapter) GetAssignment(_ context.Context, _ string, _ string) (*experiment.Assignment, error) {
	return nil, nil //nolint:nilnil // no-op adapter intentionally returns nil
}

// RecordMetric discards the metric event and returns nil.
func (a *Adapter) RecordMetric(_ context.Context, _ string, _ string, _ string, _ float64) error {
	return nil
}

// GetResults returns empty results.
func (a *Adapter) GetResults(_ context.Context, key string) (*experiment.Result, error) {
	return &experiment.Result{
		ExperimentKey: key,
		Status:        experiment.StatusDraft,
		AnalyzedAt:    time.Now(),
	}, nil
}
