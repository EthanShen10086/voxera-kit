// Package memory provides an in-memory implementation of the audit Writer and
// Reader interfaces, suitable for testing and development.
package memory

import (
	"context"
	"sync"

	"github.com/EthanShen10086/voxera-kit/audit"
)

// Adapter stores audit entries in memory using a mutex-protected slice.
type Adapter struct {
	mu      sync.Mutex
	entries []audit.Entry
}

// NewAdapter creates a new in-memory audit adapter.
func NewAdapter() *Adapter {
	return &Adapter{}
}

// Write appends a single entry to the in-memory store.
func (a *Adapter) Write(_ context.Context, entry audit.Entry) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.entries = append(a.entries, entry)
	return nil
}

// WriteBatch appends multiple entries to the in-memory store.
func (a *Adapter) WriteBatch(_ context.Context, entries []audit.Entry) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.entries = append(a.entries, entries...)
	return nil
}

// Query returns entries matching the given filter criteria.
func (a *Adapter) Query(_ context.Context, filter audit.Filter) ([]audit.Entry, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	var results []audit.Entry
	for _, e := range a.entries {
		if !matches(e, filter) {
			continue
		}
		results = append(results, e)
	}

	if filter.Offset > 0 {
		if filter.Offset >= len(results) {
			return nil, nil
		}
		results = results[filter.Offset:]
	}

	if filter.Limit > 0 && filter.Limit < len(results) {
		results = results[:filter.Limit]
	}

	return results, nil
}

// Count returns the number of entries matching the given filter criteria.
func (a *Adapter) Count(_ context.Context, filter audit.Filter) (int64, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	var count int64
	for _, e := range a.entries {
		if matches(e, filter) {
			count++
		}
	}
	return count, nil
}

func matches(e audit.Entry, f audit.Filter) bool {
	if f.TenantID != "" && e.TenantID != f.TenantID {
		return false
	}
	if f.ActorID != "" && e.ActorID != f.ActorID {
		return false
	}
	if f.Action != "" && e.Action != f.Action {
		return false
	}
	if f.ResourceType != "" && e.ResourceType != f.ResourceType {
		return false
	}
	if f.ResourceID != "" && e.ResourceID != f.ResourceID {
		return false
	}
	if !f.From.IsZero() && e.Timestamp.Before(f.From) {
		return false
	}
	if !f.To.IsZero() && e.Timestamp.After(f.To) {
		return false
	}
	return true
}
