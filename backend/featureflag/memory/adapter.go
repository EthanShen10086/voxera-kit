// Package memory provides an in-memory implementation of the featureflag Store
// interface with deterministic percentage-based rollout using SHA-256 hashing.
package memory

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"sync"

	"github.com/EthanShen10086/voxera-kit/featureflag"
)

// Adapter stores feature flags in memory with RWMutex protection.
type Adapter struct {
	mu    sync.RWMutex
	flags map[string]featureflag.Flag
}

// NewAdapter creates a new in-memory feature flag adapter.
func NewAdapter() *Adapter {
	return &Adapter{
		flags: make(map[string]featureflag.Flag),
	}
}

// IsEnabled evaluates whether a flag is enabled for the given context. It
// checks deny/allow lists first, then falls back to percentage-based rollout
// using a deterministic hash of the user ID and flag key.
func (a *Adapter) IsEnabled(_ context.Context, key string, evalCtx featureflag.EvalContext) (bool, error) {
	a.mu.RLock()
	flag, ok := a.flags[key]
	a.mu.RUnlock()

	if !ok || !flag.Enabled {
		return false, nil
	}

	for _, denied := range flag.DenyList {
		if denied == evalCtx.UserID {
			return false, nil
		}
	}

	for _, allowed := range flag.AllowList {
		if allowed == evalCtx.UserID {
			return true, nil
		}
	}

	if flag.Percentage <= 0 {
		return false, nil
	}
	if flag.Percentage >= 100 {
		return true, nil
	}

	bucket := hashBucket(evalCtx.UserID, key)
	return bucket < flag.Percentage, nil
}

// GetFlags returns all defined feature flags.
func (a *Adapter) GetFlags(_ context.Context) ([]featureflag.Flag, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	flags := make([]featureflag.Flag, 0, len(a.flags))
	for _, f := range a.flags {
		flags = append(flags, f)
	}
	return flags, nil
}

// SetFlag creates or updates a feature flag in the store.
func (a *Adapter) SetFlag(_ context.Context, flag featureflag.Flag) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.flags[flag.Key] = flag
	return nil
}

// hashBucket produces a deterministic value in [0, 100) from a user ID and flag
// key using SHA-256.
func hashBucket(userID, key string) float64 {
	h := sha256.New()
	h.Write([]byte(userID + ":" + key))
	sum := h.Sum(nil)
	val := binary.BigEndian.Uint32(sum[:4])
	return float64(val) / float64(^uint32(0)) * 100
}
