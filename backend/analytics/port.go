// Package analytics provides pluggable product analytics interfaces and
// adapters (similar to Amplitude, Mixpanel, or PostHog). It defines
// collection and querying contracts that can be backed by an in-memory
// engine, a PostHog instance, or any other analytics backend.
package analytics

import "context"

// Collector ingests raw events from frontends and backends.
type Collector interface {
	// Track records a single event.
	Track(ctx context.Context, event Event) error
	// TrackBatch records multiple events in one call.
	TrackBatch(ctx context.Context, events []Event) error
	// Identify sets or updates properties on a user profile.
	Identify(ctx context.Context, profile UserProfile) error
	// Alias links a previous anonymous ID to a new identified user ID.
	Alias(ctx context.Context, previousID, newID string) error
}

// Querier provides analytical query capabilities over collected events.
type Querier interface {
	// QueryFunnel computes conversion metrics for a multi-step funnel.
	QueryFunnel(ctx context.Context, query FunnelQuery) (*FunnelResult, error)
	// QueryRetention computes cohort retention over multiple periods.
	QueryRetention(ctx context.Context, query RetentionQuery) (*RetentionResult, error)
	// QueryPath computes user flow / path analysis between events.
	QueryPath(ctx context.Context, query PathQuery) (*PathResult, error)
	// QueryEvents returns raw events matching the given filters.
	QueryEvents(ctx context.Context, query EventQuery) ([]Event, error)
	// QueryUserProfile returns the aggregated profile for a single user.
	QueryUserProfile(ctx context.Context, userID string) (*UserProfile, error)
}

// Store combines collection and querying for full analytics support.
type Store interface {
	Collector
	Querier
}
