package analytics

import "time"

// FunnelQuery defines a multi-step funnel to analyze conversion rates.
type FunnelQuery struct {
	// TenantID scopes the analysis to a specific tenant.
	TenantID string
	// Steps lists the ordered funnel steps a user must complete.
	Steps []FunnelStep
	// From is the inclusive start of the analysis window.
	From time.Time
	// To is the exclusive end of the analysis window.
	To time.Time
	// WindowSec is the maximum number of seconds allowed between the
	// first and last step for a user to count as converted.
	WindowSec int64
	// GroupBy is an optional event property name to segment results by.
	GroupBy string
}

// FunnelStep represents one step in a conversion funnel.
type FunnelStep struct {
	// Name is the event name that qualifies for this step.
	Name string
	// Filters restricts matching to events whose properties satisfy
	// all key-value pairs (exact equality).
	Filters map[string]any
}

// FunnelResult holds the computed funnel metrics.
type FunnelResult struct {
	// Steps contains per-step conversion metrics.
	Steps []FunnelStepResult
	// OverallRate is the percentage of users who completed all steps.
	OverallRate float64
	// MedianTime is the median seconds elapsed from the first to the last step.
	MedianTime int64
}

// FunnelStepResult holds metrics for a single funnel step.
type FunnelStepResult struct {
	// Name is the event name of this step.
	Name string
	// EnteredCount is the number of users who reached this step.
	EnteredCount int64
	// DroppedCount is the number of users who did not proceed.
	DroppedCount int64
	// ConvertRate is the conversion rate from the previous step to this one.
	ConvertRate float64
	// MedianTime is the median seconds elapsed from the previous step.
	MedianTime int64
}
