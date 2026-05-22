// Package semaphore provides a channel-based Bulkhead implementation.
package semaphore

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/EthanShen10086/voxera-kit/bulkhead"
)

// Adapter implements [bulkhead.Bulkhead] using a buffered channel as a
// counting semaphore.
type Adapter struct {
	sem    chan struct{}
	active atomic.Int64
	cfg    bulkhead.Config
}

// New creates a ready-to-use semaphore-backed bulkhead.
func New(cfg bulkhead.Config) *Adapter {
	return &Adapter{
		sem: make(chan struct{}, cfg.MaxConcurrent),
		cfg: cfg,
	}
}

var _ bulkhead.Bulkhead = (*Adapter)(nil)

// Execute acquires a semaphore slot, runs fn, and releases the slot. If no
// slot is available within MaxWaitTime (combined with the parent context
// deadline) it returns [bulkhead.ErrBulkheadFull].
func (a *Adapter) Execute(ctx context.Context, fn func() error) error {
	waitCtx, cancel := context.WithTimeout(ctx, a.cfg.MaxWaitTime)
	defer cancel()

	select {
	case a.sem <- struct{}{}:
	case <-waitCtx.Done():
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return bulkhead.ErrBulkheadFull
	}

	a.active.Add(1)
	defer func() {
		a.active.Add(-1)
		<-a.sem
	}()

	return fn()
}

// ActiveCount returns the number of goroutines currently executing inside
// the bulkhead.
func (a *Adapter) ActiveCount() int {
	return int(a.active.Load())
}

// Available returns the number of free slots.
func (a *Adapter) Available() int {
	return a.cfg.MaxConcurrent - int(a.active.Load())
}

// Name returns the configured bulkhead name.
func (a *Adapter) Name() string {
	return a.cfg.Name
}

// MaxWaitTime returns the configured maximum wait duration.
func (a *Adapter) MaxWaitTime() time.Duration {
	return a.cfg.MaxWaitTime
}
