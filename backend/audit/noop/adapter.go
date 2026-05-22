// Package noop provides a no-op implementation of the audit Writer interface
// that discards all entries silently.
package noop

import (
	"context"

	"github.com/EthanShen10086/voxera-kit/audit"
)

// Adapter is a Writer that discards all audit entries.
type Adapter struct{}

// NewAdapter creates a new no-op audit adapter.
func NewAdapter() *Adapter {
	return &Adapter{}
}

// Write discards the entry and returns nil.
func (a *Adapter) Write(_ context.Context, _ audit.Entry) error {
	return nil
}

// WriteBatch discards all entries and returns nil.
func (a *Adapter) WriteBatch(_ context.Context, _ []audit.Entry) error {
	return nil
}
