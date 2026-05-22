// Package featureflag provides a feature flag evaluation system with support
// for percentage-based rollouts and allow/deny lists.
package featureflag

import "context"

// EvalContext holds the contextual information used to evaluate whether a
// feature flag is enabled for a particular request.
type EvalContext struct {
	UserID      string
	TenantID    string
	Environment string
	Attributes  map[string]any
}

// Flag represents a feature flag definition with its rollout configuration.
type Flag struct {
	Key        string
	Enabled    bool
	Percentage float64
	AllowList  []string
	DenyList   []string
}

// Store defines the interface for feature flag storage and evaluation.
type Store interface {
	// IsEnabled evaluates whether a flag is enabled for the given context.
	IsEnabled(ctx context.Context, key string, evalCtx EvalContext) (bool, error)
	// GetFlags returns all defined feature flags.
	GetFlags(ctx context.Context) ([]Flag, error)
	// SetFlag creates or updates a feature flag.
	SetFlag(ctx context.Context, flag Flag) error
}
