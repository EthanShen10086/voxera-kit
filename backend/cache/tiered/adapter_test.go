package tiered_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/EthanShen10086/voxera-kit/cache"
	"github.com/EthanShen10086/voxera-kit/cache/contract"
	"github.com/EthanShen10086/voxera-kit/cache/memory"
	"github.com/EthanShen10086/voxera-kit/cache/tiered"
)

func TestTieredCacheContract(t *testing.T) {
	contract.RunCacheContract(t, func(t *testing.T) (cache.Cache, func()) {
		l1 := memory.New()
		l2 := memory.New()
		c, err := tiered.New(l1, l2)
		if err != nil {
			t.Fatalf("New: %v", err)
		}
		return c, func() {
			_ = l1.Close()
			_ = l2.Close()
		}
	})
}

func TestNewValidation(t *testing.T) {
	if _, err := tiered.New(memory.New()); err == nil {
		t.Fatal("expected error for single layer")
	}
	if _, err := tiered.New(memory.New(), nil); err == nil {
		t.Fatal("expected error for nil layer")
	}
}

func TestGetBackfillsUpperLayer(t *testing.T) {
	ctx := context.Background()
	l1 := memory.New()
	l2 := memory.New()
	defer func() { _ = l1.Close(); _ = l2.Close() }()

	if err := l2.Set(ctx, "k", []byte("from-l2")); err != nil {
		t.Fatal(err)
	}

	c, err := tiered.New(l1, l2)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = c.Close() }()

	val, err := c.Get(ctx, "k")
	if err != nil || string(val) != "from-l2" {
		t.Fatalf("Get() = %q, %v", val, err)
	}

	l1Val, err := l1.Get(ctx, "k")
	if err != nil || string(l1Val) != "from-l2" {
		t.Fatalf("L1 backfill = %q, %v", l1Val, err)
	}
}

func TestDeleteEvictsAllLayers(t *testing.T) {
	ctx := context.Background()
	l1 := memory.New()
	l2 := memory.New()
	defer func() { _ = l1.Close(); _ = l2.Close() }()

	c, err := tiered.New(l1, l2)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = c.Close() }()

	if err := c.Set(ctx, "k", []byte("v")); err != nil {
		t.Fatal(err)
	}
	if err := c.Delete(ctx, "k"); err != nil {
		t.Fatal(err)
	}
	if _, err := l1.Get(ctx, "k"); err != cache.ErrNotFound {
		t.Fatalf("L1 after delete: %v", err)
	}
	if _, err := l2.Get(ctx, "k"); err != cache.ErrNotFound {
		t.Fatalf("L2 after delete: %v", err)
	}
}

func TestSetWithTTLExistsFlush(t *testing.T) {
	ctx := context.Background()
	l1 := memory.New()
	l2 := memory.New()
	defer func() { _ = l1.Close(); _ = l2.Close() }()

	c, err := tiered.New(l1, l2)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = c.Close() }()

	if err := c.SetWithTTL(ctx, "k", []byte("v"), time.Minute); err != nil {
		t.Fatal(err)
	}
	ok, err := c.Exists(ctx, "k")
	if err != nil || !ok {
		t.Fatalf("Exists() = %v, %v", ok, err)
	}
	if err := c.Flush(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := c.Get(ctx, "k"); !errors.Is(err, cache.ErrNotFound) {
		t.Fatalf("after flush: %v", err)
	}
}

func TestGetCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	c, err := tiered.New(memory.New(), memory.New())
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = c.Close() }()

	if _, err := c.Get(ctx, "k"); !errors.Is(err, context.Canceled) {
		t.Fatalf("Get: %v", err)
	}
}

type failCache struct {
	inner    cache.Cache
	getErr   error
	setErr   error
	delErr   error
	existErr error
	flushErr error
	closeErr error
}

func (f *failCache) Get(ctx context.Context, key string) ([]byte, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	return f.inner.Get(ctx, key)
}
func (f *failCache) Set(ctx context.Context, key string, value []byte) error {
	if f.setErr != nil {
		return f.setErr
	}
	return f.inner.Set(ctx, key, value)
}
func (f *failCache) SetWithTTL(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	if f.setErr != nil {
		return f.setErr
	}
	return f.inner.SetWithTTL(ctx, key, value, ttl)
}
func (f *failCache) Delete(ctx context.Context, key string) error {
	if f.delErr != nil {
		return f.delErr
	}
	return f.inner.Delete(ctx, key)
}
func (f *failCache) Exists(ctx context.Context, key string) (bool, error) {
	if f.existErr != nil {
		return false, f.existErr
	}
	return f.inner.Exists(ctx, key)
}
func (f *failCache) Flush(ctx context.Context) error {
	if f.flushErr != nil {
		return f.flushErr
	}
	return f.inner.Flush(ctx)
}
func (f *failCache) Close() error {
	if f.closeErr != nil {
		return f.closeErr
	}
	return f.inner.Close()
}

func TestTieredPropagatesLayerErrors(t *testing.T) {
	ctx := context.Background()
	boom := errors.New("layer failed")

	t.Run("Get", func(t *testing.T) {
		l1 := memory.New()
		defer func() { _ = l1.Close() }()
		c, err := tiered.New(l1, &failCache{inner: memory.New(), getErr: boom})
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = c.Close() }()
		if _, err := c.Get(ctx, "k"); !errors.Is(err, boom) {
			t.Fatalf("Get: %v", err)
		}
	})

	t.Run("Set", func(t *testing.T) {
		l1 := memory.New()
		l2 := &failCache{inner: memory.New(), setErr: boom}
		defer func() { _ = l1.Close(); _ = l2.Close() }()
		c, err := tiered.New(l1, l2)
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = c.Close() }()
		if err := c.Set(ctx, "k", []byte("v")); !errors.Is(err, boom) {
			t.Fatalf("Set: %v", err)
		}
	})

	t.Run("Close", func(t *testing.T) {
		l1 := memory.New()
		l2 := &failCache{inner: memory.New(), closeErr: boom}
		defer func() { _ = l1.Close() }()
		c, err := tiered.New(l1, l2)
		if err != nil {
			t.Fatal(err)
		}
		if err := c.Close(); !errors.Is(err, boom) {
			t.Fatalf("Close: %v", err)
		}
	})
}
