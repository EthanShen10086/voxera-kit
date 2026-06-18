package tiered_test

import (
	"context"
	"testing"

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
