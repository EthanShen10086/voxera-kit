// Package retry provides configurable retry policies with pluggable back-off
// strategies.
package retry

import (
	"context"
	"time"
)

// Policy describes how retries are performed: how many times, at what
// intervals, and with how much jitter.
type Policy struct {
	// MaxAttempts is the total number of attempts including the first call.
	MaxAttempts int
	// InitialBackoff is the delay before the first retry.
	InitialBackoff time.Duration
	// MaxBackoff caps the computed back-off duration.
	MaxBackoff time.Duration
	// Multiplier scales the back-off after each retry.
	Multiplier float64
	// JitterFraction is the fraction of the current back-off to randomize
	// (0 = no jitter, 1 = full jitter).
	JitterFraction float64
}

// IsRetryable decides whether err should be retried. Return false to abort
// immediately.
type IsRetryable func(err error) bool

// Retrier executes a function with automatic retries according to a policy.
type Retrier interface {
	// Execute calls fn repeatedly until it succeeds, the policy is
	// exhausted, or ctx is canceled.
	Execute(ctx context.Context, fn func(ctx context.Context) error) error
}
