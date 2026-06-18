// Package memcached provides a Memcached implementation of the cache.Cache interface.
// It uses github.com/bradfitz/gomemcache/memcache as the underlying client.
package memcached

import (
	"context"
	"errors"
	"time"

	"github.com/EthanShen10086/voxera-kit/cache"
	"github.com/bradfitz/gomemcache/memcache"
)

// Adapter implements the cache.Cache interface using Memcached.
type Adapter struct {
	client *memcache.Client
}

// New creates a new Memcached Adapter with the provided configuration.
func New(cfg cache.Config) *Adapter {
	return &Adapter{client: memcache.New(cfg.Address)}
}

// Get retrieves the value for the given key from Memcached.
func (a *Adapter) Get(ctx context.Context, key string) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	item, err := a.client.Get(key)
	if errors.Is(err, memcache.ErrCacheMiss) {
		return nil, cache.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return item.Value, nil
}

// Set stores a key-value pair in Memcached with no expiration.
func (a *Adapter) Set(ctx context.Context, key string, value []byte) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return a.client.Set(&memcache.Item{Key: key, Value: value})
}

// SetWithTTL stores a key-value pair in Memcached with the specified TTL.
func (a *Adapter) SetWithTTL(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	secs := int32(ttl.Seconds())
	if secs <= 0 {
		secs = 1
	}
	return a.client.Set(&memcache.Item{
		Key:        key,
		Value:      value,
		Expiration: secs,
	})
}

// Delete removes the given key from Memcached.
func (a *Adapter) Delete(ctx context.Context, key string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	err := a.client.Delete(key)
	if errors.Is(err, memcache.ErrCacheMiss) {
		return nil
	}
	return err
}

// Exists checks whether a key exists in Memcached.
func (a *Adapter) Exists(ctx context.Context, key string) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	_, err := a.client.Get(key)
	if errors.Is(err, memcache.ErrCacheMiss) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// Flush removes all keys from Memcached.
func (a *Adapter) Flush(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return a.client.FlushAll()
}

// Close releases all resources held by the Memcached client.
func (a *Adapter) Close() error {
	return nil
}
