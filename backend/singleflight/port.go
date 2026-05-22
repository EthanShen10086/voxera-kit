// Package singleflight provides request deduplication so that only one
// in-flight call for a given key executes at a time.
package singleflight

import "context"

// Deduplicator suppresses duplicate calls that share the same key.
// Implementations must be safe for concurrent use.
type Deduplicator interface {
	// Do executes fn once for a given key, returning the result to all
	// callers that share the same in-flight request. The shared return
	// value is true when the result was produced by another caller.
	Do(ctx context.Context, key string, fn func() (any, error)) (any, bool, error)
}
