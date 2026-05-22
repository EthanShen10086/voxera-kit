// Package bulkhead isolates workloads by limiting concurrency, preventing one
// component from exhausting shared resources.
package bulkhead

import (
	"context"
	"errors"
	"time"
)

// ErrBulkheadFull is returned when the bulkhead has no capacity and the wait
// time has been exceeded.
var ErrBulkheadFull = errors.New("bulkhead: full")

// Config controls the behavior of a Bulkhead.
type Config struct {
	// MaxConcurrent is the maximum number of calls allowed to execute
	// simultaneously.
	MaxConcurrent int
	// MaxWaitTime is the longest a caller will wait for a slot before
	// receiving ErrBulkheadFull.
	MaxWaitTime time.Duration
	// Name identifies this bulkhead in logs and metrics.
	Name string
}

// Bulkhead limits the number of concurrent executions of a function.
type Bulkhead interface {
	// Execute runs fn when a slot is available. It returns ErrBulkheadFull
	// if no slot becomes available within MaxWaitTime.
	Execute(ctx context.Context, fn func() error) error
	// ActiveCount returns the number of currently executing calls.
	ActiveCount() int
}
