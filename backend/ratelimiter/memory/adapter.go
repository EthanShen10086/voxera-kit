// Package memory provides an in-process token bucket rate limiter implementation
// of the ratelimiter.RateLimiter interface.
package memory

import (
	"context"
	"sync"
	"time"

	"github.com/EthanShen10086/voxera-kit/ratelimiter"
)

type bucket struct {
	tokens     float64
	lastRefill time.Time
}

// Adapter implements ratelimiter.RateLimiter using an in-memory token bucket algorithm.
type Adapter struct {
	mu      sync.Mutex
	buckets map[string]*bucket
	rate    float64
	burst   int
}

// New creates a new in-memory token bucket rate limiter.
func New(cfg ratelimiter.Config) *Adapter {
	return &Adapter{
		buckets: make(map[string]*bucket),
		rate:    cfg.Rate,
		burst:   cfg.Burst,
	}
}

// Allow reports whether the event identified by key is permitted under the token bucket rate.
func (a *Adapter) Allow(_ context.Context, key string) (bool, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	now := time.Now()
	b, ok := a.buckets[key]
	if !ok {
		b = &bucket{tokens: float64(a.burst), lastRefill: now}
		a.buckets[key] = b
	}

	elapsed := now.Sub(b.lastRefill).Seconds()
	b.tokens += elapsed * a.rate
	if b.tokens > float64(a.burst) {
		b.tokens = float64(a.burst)
	}
	b.lastRefill = now

	if b.tokens < 1 {
		return false, nil
	}
	b.tokens--
	return true, nil
}

// Reset clears the rate limit state for the given key.
func (a *Adapter) Reset(_ context.Context, key string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	delete(a.buckets, key)
	return nil
}

// Close releases all resources held by the rate limiter.
func (a *Adapter) Close() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.buckets = make(map[string]*bucket)
	return nil
}
