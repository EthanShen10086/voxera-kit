// Package memcached provides a Memcached implementation of the cache.Cache interface.
// It is intended to use github.com/bradfitz/gomemcache/memcache as the underlying client.
package memcached

import (
	"context"
	"time"

	"github.com/EthanShen10086/voxera-kit/cache"
)

// Adapter implements the cache.Cache interface using Memcached.
//
// Intended dependency: github.com/bradfitz/gomemcache/memcache
type Adapter struct {
	// client *memcache.Client // TODO: uncomment when gomemcache dependency is added
	cfg cache.CacheConfig
}

// New creates a new Memcached Adapter with the provided configuration.
func New(cfg cache.CacheConfig) *Adapter {
	return &Adapter{cfg: cfg}
}

// Get retrieves the value for the given key from Memcached.
func (a *Adapter) Get(ctx context.Context, key string) ([]byte, error) {
	// TODO: implement using gomemcache
	return nil, nil
}

// Set stores a key-value pair in Memcached with no expiration.
func (a *Adapter) Set(ctx context.Context, key string, value []byte) error {
	// TODO: implement using gomemcache
	return nil
}

// SetWithTTL stores a key-value pair in Memcached with the specified TTL.
func (a *Adapter) SetWithTTL(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	// TODO: implement using gomemcache
	return nil
}

// Delete removes the given key from Memcached.
func (a *Adapter) Delete(ctx context.Context, key string) error {
	// TODO: implement using gomemcache
	return nil
}

// Exists checks whether a key exists in Memcached.
func (a *Adapter) Exists(ctx context.Context, key string) (bool, error) {
	// TODO: implement using gomemcache
	return false, nil
}

// Flush removes all keys from Memcached.
func (a *Adapter) Flush(ctx context.Context) error {
	// TODO: implement using gomemcache
	return nil
}

// Close releases all resources held by the Memcached client.
func (a *Adapter) Close() error {
	// TODO: implement using gomemcache
	return nil
}
