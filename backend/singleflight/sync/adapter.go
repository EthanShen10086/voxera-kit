// Package sync provides a Deduplicator backed by golang.org/x/sync/singleflight.
package sync

import (
	"context"

	sf "github.com/EthanShen10086/voxera-kit/singleflight"
	"golang.org/x/sync/singleflight"
)

// Adapter wraps a singleflight.Group to implement [sf.Deduplicator].
type Adapter struct {
	group singleflight.Group
}

// New returns a ready-to-use Adapter.
func New() *Adapter {
	return &Adapter{}
}

var _ sf.Deduplicator = (*Adapter)(nil)

// Do delegates to the underlying singleflight.Group. If ctx is canceled
// before fn completes, the context error is returned.
func (a *Adapter) Do(ctx context.Context, key string, fn func() (any, error)) (any, bool, error) {
	ch := a.group.DoChan(key, func() (any, error) {
		return fn()
	})

	select {
	case <-ctx.Done():
		return nil, false, ctx.Err()
	case res := <-ch:
		return res.Val, res.Shared, res.Err
	}
}
