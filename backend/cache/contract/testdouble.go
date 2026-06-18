// Package contract provides shared contract tests for cache.Cache implementations.
package contract

import "github.com/EthanShen10086/voxera-kit/cache/memory"

// TestDouble is a minimal in-memory cache used to exercise contract tests
// without external dependencies.
type TestDouble struct {
	*memory.Adapter
}

// NewTestDouble returns a fresh in-memory cache for contract testing.
func NewTestDouble() *TestDouble {
	return &TestDouble{Adapter: memory.New()}
}
