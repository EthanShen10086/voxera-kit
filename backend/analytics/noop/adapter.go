// Package noop provides an analytics.Store that silently discards all
// tracked events and returns empty results for all queries. Use it when
// analytics is disabled or in unit tests that do not exercise analytics.
package noop

import (
	"context"

	"github.com/EthanShen10086/voxera-kit/analytics"
)

// Adapter is a no-op implementation of analytics.Store.
type Adapter struct{}

// compile-time interface check.
var _ analytics.Store = (*Adapter)(nil)

// New creates a no-op analytics adapter.
func New() *Adapter {
	return &Adapter{}
}

// Track is a no-op that always returns nil.
func (a *Adapter) Track(_ context.Context, _ analytics.Event) error { return nil }

// TrackBatch is a no-op that always returns nil.
func (a *Adapter) TrackBatch(_ context.Context, _ []analytics.Event) error { return nil }

// Identify is a no-op that always returns nil.
func (a *Adapter) Identify(_ context.Context, _ analytics.UserProfile) error { return nil }

// Alias is a no-op that always returns nil.
func (a *Adapter) Alias(_ context.Context, _, _ string) error { return nil }

// QueryFunnel returns an empty FunnelResult.
func (a *Adapter) QueryFunnel(_ context.Context, _ analytics.FunnelQuery) (*analytics.FunnelResult, error) {
	return &analytics.FunnelResult{}, nil
}

// QueryRetention returns an empty RetentionResult.
func (a *Adapter) QueryRetention(_ context.Context, _ analytics.RetentionQuery) (*analytics.RetentionResult, error) {
	return &analytics.RetentionResult{}, nil
}

// QueryPath returns an empty PathResult.
func (a *Adapter) QueryPath(_ context.Context, _ analytics.PathQuery) (*analytics.PathResult, error) {
	return &analytics.PathResult{}, nil
}

// QueryEvents returns an empty event slice.
func (a *Adapter) QueryEvents(_ context.Context, _ analytics.EventQuery) ([]analytics.Event, error) {
	return nil, nil
}

// QueryUserProfile returns an empty UserProfile.
func (a *Adapter) QueryUserProfile(_ context.Context, _ string) (*analytics.UserProfile, error) {
	return &analytics.UserProfile{}, nil
}
