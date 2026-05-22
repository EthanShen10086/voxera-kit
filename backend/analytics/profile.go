package analytics

import "time"

// UserProfile holds aggregated user-level analytics data.
type UserProfile struct {
	// UserID is the unique identifier for this user.
	UserID string
	// TenantID is the tenant this user belongs to.
	TenantID string
	// FirstSeen is the timestamp of the user's earliest tracked event.
	FirstSeen time.Time
	// LastSeen is the timestamp of the user's most recent tracked event.
	LastSeen time.Time
	// SessionCount is the total number of distinct sessions observed.
	SessionCount int64
	// EventCount is the total number of events tracked for this user.
	EventCount int64
	// Properties holds custom user properties set via Identify.
	Properties map[string]any
	// TopEvents lists the user's most frequent events in descending order.
	TopEvents []EventCount
	// Attribution holds the user's marketing channel attribution data.
	Attribution Attribution
}

// EventCount pairs an event name with its occurrence count.
type EventCount struct {
	// Name is the event name.
	Name string
	// Count is the number of occurrences.
	Count int64
}
