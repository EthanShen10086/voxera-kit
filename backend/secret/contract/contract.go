package contract

import (
	"context"
	"errors"
	"testing"

	"github.com/EthanShen10086/voxera-kit/secret"
)

// Factory creates a secret Manager and an optional cleanup function.
type Factory func(t *testing.T) (secret.Manager, func())

// RunSecretContract exercises Set/Get/Delete/List roundtrips.
func RunSecretContract(t *testing.T, factory Factory) {
	t.Helper()
	ctx := context.Background()

	mgr, cleanup := factory(t)
	if cleanup != nil {
		defer cleanup()
	}

	t.Run("SetGet", func(t *testing.T) {
		const key = "api-key"
		const value = "super-secret"
		if err := mgr.Set(ctx, key, value); err != nil {
			t.Fatalf("Set() = %v", err)
		}
		got, err := mgr.Get(ctx, key)
		if err != nil {
			t.Fatalf("Get() = %v", err)
		}
		if got != value {
			t.Fatalf("Get() = %q, want %q", got, value)
		}
	})

	t.Run("DeleteNotFound", func(t *testing.T) {
		const key = "delete-me"
		if err := mgr.Set(ctx, key, "temp"); err != nil {
			t.Fatalf("Set() = %v", err)
		}
		if err := mgr.Delete(ctx, key); err != nil {
			t.Fatalf("Delete() = %v", err)
		}
		_, err := mgr.Get(ctx, key)
		if !errors.Is(err, secret.ErrNotFound) {
			t.Fatalf("Get() after delete = %v, want %v", err, secret.ErrNotFound)
		}
	})

	t.Run("ListPrefix", func(t *testing.T) {
		if err := mgr.Set(ctx, "prefix/a", "1"); err != nil {
			t.Fatalf("Set prefix/a: %v", err)
		}
		if err := mgr.Set(ctx, "prefix/b", "2"); err != nil {
			t.Fatalf("Set prefix/b: %v", err)
		}
		if err := mgr.Set(ctx, "other/c", "3"); err != nil {
			t.Fatalf("Set other/c: %v", err)
		}

		keys, err := mgr.List(ctx, "prefix/")
		if err != nil {
			t.Fatalf("List() = %v", err)
		}
		if len(keys) < 2 {
			t.Fatalf("List(prefix/) = %v, want at least 2 keys", keys)
		}
	})
}
