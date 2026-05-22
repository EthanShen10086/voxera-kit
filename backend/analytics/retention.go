package analytics

import "time"

// Granularity defines time bucketing for retention analysis.
type Granularity string

const (
	// GranularityDay buckets retention data by calendar day.
	GranularityDay Granularity = "day"
	// GranularityWeek buckets retention data by ISO week.
	GranularityWeek Granularity = "week"
	// GranularityMonth buckets retention data by calendar month.
	GranularityMonth Granularity = "month"
)

// RetentionQuery defines parameters for retention / cohort analysis.
type RetentionQuery struct {
	// TenantID scopes the analysis to a specific tenant.
	TenantID string
	// CohortEvent is the event that defines cohort membership (e.g. "signup").
	CohortEvent string
	// ReturnEvent is the event that counts as "returned" (e.g. "login").
	ReturnEvent string
	// From is the inclusive start of the analysis window.
	From time.Time
	// To is the exclusive end of the analysis window.
	To time.Time
	// Granularity controls the time bucket size (day, week, or month).
	Granularity Granularity
	// Periods is the number of subsequent periods to track retention for.
	Periods int
}

// RetentionResult holds the computed cohort retention data.
type RetentionResult struct {
	// Cohorts contains one row per cohort period.
	Cohorts []CohortRow
}

// CohortRow represents one cohort — users who performed the cohort event in a given period.
type CohortRow struct {
	// Period is the label for this cohort, e.g. "2026-W21" or "2026-05-22".
	Period string
	// CohortSize is the number of unique users who entered this cohort.
	CohortSize int64
	// Retained lists the count of users retained in each subsequent period.
	Retained []int64
	// Rates lists the retention rate (0.0–1.0) for each subsequent period.
	Rates []float64
}
