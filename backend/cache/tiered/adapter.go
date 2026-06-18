// Package tiered provides a multi-level cache.Cache that composes faster
// local layers with slower remote backends (e.g. ristretto + Redis).
package tiered

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/EthanShen10086/voxera-kit/cache"
)

// Adapter implements cache.Cache as a tiered stack (L1 → L2 → …).
// Reads populate upper layers on a lower-layer hit; writes go to all layers.
type Adapter struct {
	layers []cache.Cache
}

// New builds a tiered cache from ordered layers (index 0 = fastest / closest).
func New(layers ...cache.Cache) (*Adapter, error) {
	if len(layers) < 2 {
		return nil, fmt.Errorf("tiered: at least two cache layers are required")
	}
	for i, c := range layers {
		if c == nil {
			return nil, fmt.Errorf("tiered: layer %d is nil", i)
		}
	}
	copied := make([]cache.Cache, len(layers))
	copy(copied, layers)
	return &Adapter{layers: copied}, nil
}

// Get returns from the first layer that has the key, back-filling faster layers.
func (a *Adapter) Get(ctx context.Context, key string) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	var lastErr error
	for i, layer := range a.layers {
		val, err := layer.Get(ctx, key)
		if err == nil {
			for j := 0; j < i; j++ {
				_ = a.layers[j].Set(ctx, key, val)
			}
			return val, nil
		}
		if !errors.Is(err, cache.ErrNotFound) {
			lastErr = err
		}
	}
	if lastErr != nil {
		return nil, lastErr
	}
	return nil, cache.ErrNotFound
}

// Set stores the value in every layer.
func (a *Adapter) Set(ctx context.Context, key string, value []byte) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	for _, layer := range a.layers {
		if err := layer.Set(ctx, key, value); err != nil {
			return err
		}
	}
	return nil
}

// SetWithTTL stores the value in every layer with the same TTL.
func (a *Adapter) SetWithTTL(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	for _, layer := range a.layers {
		if err := layer.SetWithTTL(ctx, key, value, ttl); err != nil {
			return err
		}
	}
	return nil
}

// Delete removes the key from every layer.
func (a *Adapter) Delete(ctx context.Context, key string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	for _, layer := range a.layers {
		if err := layer.Delete(ctx, key); err != nil {
			return err
		}
	}
	return nil
}

// Exists reports true if any layer contains the key.
func (a *Adapter) Exists(ctx context.Context, key string) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	for _, layer := range a.layers {
		ok, err := layer.Exists(ctx, key)
		if err != nil {
			return false, err
		}
		if ok {
			return true, nil
		}
	}
	return false, nil
}

// Flush clears every layer.
func (a *Adapter) Flush(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	for _, layer := range a.layers {
		if err := layer.Flush(ctx); err != nil {
			return err
		}
	}
	return nil
}

// Close closes all layers in order.
func (a *Adapter) Close() error {
	var firstErr error
	for _, layer := range a.layers {
		if err := layer.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

var _ cache.Cache = (*Adapter)(nil)
