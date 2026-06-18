// Package memory provides an in-memory implementation of the cache.Cache interface
// for testing and single-process deployments.
package memory

import (
	"context"
	"sync"
	"time"

	"github.com/EthanShen10086/voxera-kit/cache"
)

type entry struct {
	value     []byte
	expiresAt time.Time
}

// Adapter implements cache.Cache using a mutex-protected map.
type Adapter struct {
	mu    sync.RWMutex
	items map[string]entry
}

// New creates a new in-memory cache adapter.
func New() *Adapter {
	return &Adapter{items: make(map[string]entry)}
}

// Get retrieves the value for the given key.
func (a *Adapter) Get(ctx context.Context, key string) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	a.mu.RLock()
	defer a.mu.RUnlock()

	e, ok := a.items[key]
	if !ok || a.expired(e) {
		return nil, cache.ErrNotFound
	}
	return append([]byte(nil), e.value...), nil
}

// Set stores a key-value pair with no expiration.
func (a *Adapter) Set(ctx context.Context, key string, value []byte) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	a.items[key] = entry{value: append([]byte(nil), value...)}
	return nil
}

// SetWithTTL stores a key-value pair that expires after the given duration.
func (a *Adapter) SetWithTTL(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	a.items[key] = entry{
		value:     append([]byte(nil), value...),
		expiresAt: time.Now().Add(ttl),
	}
	return nil
}

// Delete removes the given key.
func (a *Adapter) Delete(ctx context.Context, key string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	delete(a.items, key)
	return nil
}

// Exists checks whether a key exists and has not expired.
func (a *Adapter) Exists(ctx context.Context, key string) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	a.mu.RLock()
	defer a.mu.RUnlock()

	e, ok := a.items[key]
	return ok && !a.expired(e), nil
}

// Flush removes all entries.
func (a *Adapter) Flush(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	a.items = make(map[string]entry)
	return nil
}

// Close releases all resources held by the adapter.
func (a *Adapter) Close() error {
	return a.Flush(context.Background())
}

func (a *Adapter) expired(e entry) bool {
	return !e.expiresAt.IsZero() && time.Now().After(e.expiresAt)
}
