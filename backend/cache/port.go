// Package cache defines the port interface for caching operations.
// It abstracts away the underlying cache implementation (Redis, Memcached, local)
// allowing different backends to be used interchangeably.
package cache

import (
	"context"
	"errors"
	"time"
)

// ErrNotFound is returned when a requested cache key does not exist.
var ErrNotFound = errors.New("cache: key not found")

// Config holds the connection parameters for a cache backend.
type Config struct {
	// Address is the cache server address (e.g., "localhost:6379").
	Address string
	// Password is the authentication password for the cache server.
	Password string
	// DB is the database number to select (applicable to Redis).
	DB int
	// PoolSize controls the maximum number of connections in the pool.
	PoolSize int
	// DialTimeout is the maximum duration to wait for a connection to be established.
	DialTimeout time.Duration
	// ReadTimeout is the maximum duration to wait for a read operation.
	ReadTimeout time.Duration
	// WriteTimeout is the maximum duration to wait for a write operation.
	WriteTimeout time.Duration
}

// Cache is the interface for key-value cache operations.
// Implementations must be safe for concurrent use.
type Cache interface {
	// Get retrieves the value associated with the given key.
	// Returns an error if the key does not exist or the operation fails.
	Get(ctx context.Context, key string) ([]byte, error)
	// Set stores a key-value pair with no expiration.
	Set(ctx context.Context, key string, value []byte) error
	// SetWithTTL stores a key-value pair that expires after the given duration.
	SetWithTTL(ctx context.Context, key string, value []byte, ttl time.Duration) error
	// Delete removes the value associated with the given key.
	Delete(ctx context.Context, key string) error
	// Exists checks whether a key exists in the cache.
	Exists(ctx context.Context, key string) (bool, error)
	// Flush removes all entries from the cache.
	Flush(ctx context.Context) error
	// Close releases all resources held by the cache client.
	Close() error
}
