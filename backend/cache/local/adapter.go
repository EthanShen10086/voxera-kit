// Package local provides an in-process cache implementation of the cache.Cache interface.
// It is intended to use github.com/dgraph-io/ristretto as the underlying cache engine.
package local

import (
	"context"
	"time"

	"github.com/EthanShen10086/voxera-kit/cache"
)

// Adapter implements the cache.Cache interface using an in-process cache (Ristretto).
//
// Intended dependency: github.com/dgraph-io/ristretto
type Adapter struct {
	// cache *ristretto.Cache[string, []byte] // TODO: uncomment when ristretto dependency is added
	cfg cache.Config
}

// New creates a new local in-process cache Adapter.
func New(cfg cache.Config) *Adapter {
	return &Adapter{cfg: cfg}
}

// Get retrieves the value for the given key from the local cache.
func (a *Adapter) Get(ctx context.Context, key string) ([]byte, error) {
	// TODO: implement using ristretto
	return nil, nil
}

// Set stores a key-value pair in the local cache with no expiration.
func (a *Adapter) Set(ctx context.Context, key string, value []byte) error {
	// TODO: implement using ristretto
	return nil
}

// SetWithTTL stores a key-value pair in the local cache with the specified TTL.
func (a *Adapter) SetWithTTL(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	// TODO: implement using ristretto
	return nil
}

// Delete removes the given key from the local cache.
func (a *Adapter) Delete(ctx context.Context, key string) error {
	// TODO: implement using ristretto
	return nil
}

// Exists checks whether a key exists in the local cache.
func (a *Adapter) Exists(ctx context.Context, key string) (bool, error) {
	// TODO: implement using ristretto
	return false, nil
}

// Flush removes all entries from the local cache.
func (a *Adapter) Flush(ctx context.Context) error {
	// TODO: implement using ristretto
	return nil
}

// Close releases all resources held by the local cache.
func (a *Adapter) Close() error {
	// TODO: implement using ristretto
	return nil
}
