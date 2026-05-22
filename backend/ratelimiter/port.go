// Package ratelimiter defines the port interfaces for configurable rate limiting.
// It supports multiple strategies including token bucket, sliding window, leaky bucket,
// and fixed window, allowing different backends to be used interchangeably.
package ratelimiter

import (
	"context"
	"time"
)

// Strategy represents the rate limiting algorithm to use.
type Strategy int

const (
	// TokenBucket uses a token bucket algorithm where tokens are added at a fixed rate.
	TokenBucket Strategy = iota
	// SlidingWindow uses a sliding window counter for rate limiting.
	SlidingWindow
	// LeakyBucket uses a leaky bucket algorithm with a constant drain rate.
	LeakyBucket
	// FixedWindow uses a fixed time window counter for rate limiting.
	FixedWindow
)

// Config holds the parameters for constructing a rate limiter.
type Config struct {
	// Enabled controls whether rate limiting is active.
	Enabled bool
	// Strategy selects the rate limiting algorithm.
	Strategy Strategy
	// Rate is the number of allowed events per second.
	Rate float64
	// Burst is the maximum number of events allowed in a single burst.
	Burst int
	// KeyExtractor derives a rate-limit key from the context (e.g., client IP, user ID).
	KeyExtractor func(ctx context.Context) string
	// WindowSize is the duration of the rate limiting window (used by windowed strategies).
	WindowSize time.Duration
}

// RateLimiter checks whether a given key is allowed to proceed under the configured rate.
// Implementations must be safe for concurrent use.
type RateLimiter interface {
	// Allow reports whether the event identified by key is permitted.
	Allow(ctx context.Context, key string) (bool, error)
	// Reset clears the rate limit state for the given key.
	Reset(ctx context.Context, key string) error
	// Close releases all resources held by the rate limiter.
	Close() error
}

// Factory creates RateLimiter instances from configuration.
type Factory interface {
	// Create returns a new RateLimiter configured according to cfg.
	Create(cfg Config) (RateLimiter, error)
}
