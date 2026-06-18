package memcached_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/EthanShen10086/voxera-kit/cache"
	"github.com/EthanShen10086/voxera-kit/cache/memcached"
)

func TestNewAdapter(t *testing.T) {
	a := memcached.New(cache.Config{Address: "127.0.0.1:11211"})
	if a == nil {
		t.Fatal("nil adapter")
	}
	if err := a.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestContextCancel(t *testing.T) {
	a := memcached.New(cache.Config{Address: "127.0.0.1:11211"})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := a.Get(ctx, "k"); !errors.Is(err, context.Canceled) {
		t.Fatalf("Get() = %v", err)
	}
	if err := a.Set(ctx, "k", []byte("v")); !errors.Is(err, context.Canceled) {
		t.Fatalf("Set() = %v", err)
	}
	if err := a.SetWithTTL(ctx, "k", []byte("v"), time.Minute); !errors.Is(err, context.Canceled) {
		t.Fatalf("SetWithTTL() = %v", err)
	}
	if err := a.Delete(ctx, "k"); !errors.Is(err, context.Canceled) {
		t.Fatalf("Delete() = %v", err)
	}
	if _, err := a.Exists(ctx, "k"); !errors.Is(err, context.Canceled) {
		t.Fatalf("Exists() = %v", err)
	}
	if err := a.Flush(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("Flush() = %v", err)
	}
}

func TestSetWithTTLZeroDuration(t *testing.T) {
	a := memcached.New(cache.Config{Address: "127.0.0.1:1"})
	// exercises TTL clamp to 1 second before dial fails
	if err := a.SetWithTTL(context.Background(), "k", []byte("v"), 0); err == nil {
		t.Fatal("expected error from unreachable memcached")
	}
}

func TestGetMissOnUnreachable(t *testing.T) {
	a := memcached.New(cache.Config{Address: "127.0.0.1:1"})
	_, err := a.Get(context.Background(), "missing")
	if err == nil {
		t.Fatal("expected dial error")
	}
}
