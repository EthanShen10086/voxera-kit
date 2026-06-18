// Package redis provides a Redis implementation of the cache.Cache interface.
// It uses github.com/redis/go-redis/v9 as the underlying client.
package redis

import (
	"context"
	"errors"
	"time"

	"github.com/EthanShen10086/voxera-kit/cache"
	"github.com/redis/go-redis/v9"
)

// Adapter implements the cache.Cache interface using Redis.
type Adapter struct {
	client *redis.Client
}

// New creates a new Redis Adapter with the provided configuration.
func New(cfg cache.Config) *Adapter {
	opts := &redis.Options{
		Addr:         cfg.Address,
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}
	return &Adapter{client: redis.NewClient(opts)}
}

// Get retrieves the value for the given key from Redis.
func (a *Adapter) Get(ctx context.Context, key string) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	val, err := a.client.Get(ctx, key).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, cache.ErrNotFound
	}
	return val, err
}

// Set stores a key-value pair in Redis with no expiration.
func (a *Adapter) Set(ctx context.Context, key string, value []byte) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return a.client.Set(ctx, key, value, 0).Err()
}

// SetWithTTL stores a key-value pair in Redis with the specified TTL.
func (a *Adapter) SetWithTTL(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return a.client.Set(ctx, key, value, ttl).Err()
}

// Delete removes the given key from Redis.
func (a *Adapter) Delete(ctx context.Context, key string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return a.client.Del(ctx, key).Err()
}

// Exists checks whether a key exists in Redis.
func (a *Adapter) Exists(ctx context.Context, key string) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	n, err := a.client.Exists(ctx, key).Result()
	return n > 0, err
}

// Flush removes all keys from the current Redis database.
func (a *Adapter) Flush(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return a.client.FlushDB(ctx).Err()
}

// Close shuts down the Redis client connection.
func (a *Adapter) Close() error {
	return a.client.Close()
}
