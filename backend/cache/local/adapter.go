// Package local provides an in-process cache implementation of the cache.Cache interface.
// It uses github.com/dgraph-io/ristretto as the underlying cache engine.
package local

import (
	"context"
	"time"

	"github.com/EthanShen10086/voxera-kit/cache"
	"github.com/dgraph-io/ristretto"
)

const defaultMaxCost = 1 << 30 // 1 GiB

// Adapter implements the cache.Cache interface using an in-process Ristretto cache.
type Adapter struct {
	cache *ristretto.Cache
}

// New creates a new local in-process cache Adapter.
func New(_ cache.Config) (*Adapter, error) {
	c, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e7,
		MaxCost:     defaultMaxCost,
		BufferItems: 64,
	})
	if err != nil {
		return nil, err
	}
	return &Adapter{cache: c}, nil
}

// Get retrieves the value for the given key from the local cache.
func (a *Adapter) Get(ctx context.Context, key string) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	val, ok := a.cache.Get(key)
	if !ok {
		return nil, cache.ErrNotFound
	}
	bytes, ok := val.([]byte)
	if !ok {
		return nil, cache.ErrNotFound
	}
	return bytes, nil
}

// Set stores a key-value pair in the local cache with no expiration.
func (a *Adapter) Set(ctx context.Context, key string, value []byte) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	a.cache.Set(key, value, int64(len(value)))
	a.cache.Wait()
	return nil
}

// SetWithTTL stores a key-value pair in the local cache with the specified TTL.
func (a *Adapter) SetWithTTL(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	a.cache.SetWithTTL(key, value, int64(len(value)), ttl)
	a.cache.Wait()
	return nil
}

// Delete removes the given key from the local cache.
func (a *Adapter) Delete(ctx context.Context, key string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	a.cache.Del(key)
	return nil
}

// Exists checks whether a key exists in the local cache.
func (a *Adapter) Exists(ctx context.Context, key string) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	_, ok := a.cache.Get(key)
	return ok, nil
}

// Flush removes all entries from the local cache.
func (a *Adapter) Flush(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	a.cache.Clear()
	return nil
}

// Close releases all resources held by the local cache.
func (a *Adapter) Close() error {
	a.cache.Close()
	return nil
}
