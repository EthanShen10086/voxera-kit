// Package audit provides structured audit logging for tracking user and system
// actions across tenants.
package audit

import (
	"context"
	"time"
)

// Entry represents a single audit log record capturing who did what, when, and
// where within the system.
type Entry struct {
	ID             string
	TenantID       string
	ActorID        string
	ActorType      string
	Action         string
	ResourceType   string
	ResourceID     string
	IPAddress      string
	UserAgent      string
	Metadata       map[string]any
	RequestBody    []byte
	ResponseStatus int
	Timestamp      time.Time
}

// Writer defines the interface for persisting audit entries.
type Writer interface {
	// Write persists a single audit entry.
	Write(ctx context.Context, entry Entry) error
	// WriteBatch persists multiple audit entries atomically.
	WriteBatch(ctx context.Context, entries []Entry) error
}

// Reader defines the interface for querying persisted audit entries.
type Reader interface {
	// Query returns audit entries matching the given filter.
	Query(ctx context.Context, filter Filter) ([]Entry, error)
	// Count returns the number of audit entries matching the given filter.
	Count(ctx context.Context, filter Filter) (int64, error)
}

// Filter specifies criteria for querying audit entries.
type Filter struct {
	TenantID     string
	ActorID      string
	Action       string
	ResourceType string
	ResourceID   string
	From         time.Time
	To           time.Time
	Limit        int
	Offset       int
}
