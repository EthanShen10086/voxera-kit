package contract

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/EthanShen10086/voxera-kit/cache"
)

// Factory creates a cache.Cache instance and an optional cleanup function.
type Factory func(t *testing.T) (cache.Cache, func())

// RunCacheContract exercises the cache.Cache interface against the given factory.
func RunCacheContract(t *testing.T, factory Factory) {
	t.Helper()
	ctx := context.Background()

	c, cleanup := factory(t)
	if cleanup != nil {
		defer cleanup()
	}
	defer func() { _ = c.Close() }()

	t.Run("GetMissing", func(t *testing.T) {
		_, err := c.Get(ctx, "missing-key")
		if !errors.Is(err, cache.ErrNotFound) {
			t.Fatalf("Get missing: got %v, want %v", err, cache.ErrNotFound)
		}
	})

	t.Run("SetGet", func(t *testing.T) {
		key := "set-get"
		val := []byte("hello")
		if err := c.Set(ctx, key, val); err != nil {
			t.Fatalf("Set: %v", err)
		}
		got, err := c.Get(ctx, key)
		if err != nil {
			t.Fatalf("Get: %v", err)
		}
		if string(got) != string(val) {
			t.Fatalf("Get = %q, want %q", got, val)
		}
	})

	t.Run("SetWithTTL", func(t *testing.T) {
		key := "ttl-key"
		if err := c.SetWithTTL(ctx, key, []byte("ttl"), 50*time.Millisecond); err != nil {
			t.Fatalf("SetWithTTL: %v", err)
		}
		if _, err := c.Get(ctx, key); err != nil {
			t.Fatalf("Get before expiry: %v", err)
		}
		time.Sleep(60 * time.Millisecond)
		if _, err := c.Get(ctx, key); !errors.Is(err, cache.ErrNotFound) {
			t.Fatalf("Get after expiry: got %v, want %v", err, cache.ErrNotFound)
		}
	})

	t.Run("Delete", func(t *testing.T) {
		key := "delete-me"
		if err := c.Set(ctx, key, []byte("x")); err != nil {
			t.Fatalf("Set: %v", err)
		}
		if err := c.Delete(ctx, key); err != nil {
			t.Fatalf("Delete: %v", err)
		}
		if _, err := c.Get(ctx, key); !errors.Is(err, cache.ErrNotFound) {
			t.Fatalf("Get after delete: got %v, want %v", err, cache.ErrNotFound)
		}
	})

	t.Run("Exists", func(t *testing.T) {
		key := "exists-key"
		ok, err := c.Exists(ctx, key)
		if err != nil {
			t.Fatalf("Exists missing: %v", err)
		}
		if ok {
			t.Fatal("Exists missing: got true, want false")
		}
		if err := c.Set(ctx, key, []byte("1")); err != nil {
			t.Fatalf("Set: %v", err)
		}
		ok, err = c.Exists(ctx, key)
		if err != nil {
			t.Fatalf("Exists present: %v", err)
		}
		if !ok {
			t.Fatal("Exists present: got false, want true")
		}
	})

	t.Run("Flush", func(t *testing.T) {
		if err := c.Set(ctx, "flush-a", []byte("a")); err != nil {
			t.Fatalf("Set a: %v", err)
		}
		if err := c.Set(ctx, "flush-b", []byte("b")); err != nil {
			t.Fatalf("Set b: %v", err)
		}
		if err := c.Flush(ctx); err != nil {
			t.Fatalf("Flush: %v", err)
		}
		ok, err := c.Exists(ctx, "flush-a")
		if err != nil {
			t.Fatalf("Exists a: %v", err)
		}
		if ok {
			t.Fatal("expected flush-a to be gone")
		}
	})
}
