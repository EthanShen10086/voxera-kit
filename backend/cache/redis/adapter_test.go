package redis_test

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/EthanShen10086/voxera-kit/cache"
	"github.com/EthanShen10086/voxera-kit/cache/redis"
)

func newAdapter(t *testing.T) (*redis.Adapter, *miniredis.Miniredis) {
	t.Helper()
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	return redis.New(cache.Config{Address: mr.Addr()}), mr
}

func TestGetSetDelete(t *testing.T) {
	ctx := context.Background()
	a, mr := newAdapter(t)
	defer mr.Close()

	if err := a.Set(ctx, "k", []byte("v")); err != nil {
		t.Fatal(err)
	}
	val, err := a.Get(ctx, "k")
	if err != nil || string(val) != "v" {
		t.Fatalf("Get: %q err=%v", val, err)
	}
	ok, err := a.Exists(ctx, "k")
	if err != nil || !ok {
		t.Fatalf("Exists: %v err=%v", ok, err)
	}
	if err := a.Delete(ctx, "k"); err != nil {
		t.Fatal(err)
	}
	_, err = a.Get(ctx, "k")
	if err != cache.ErrNotFound {
		t.Fatalf("Get missing: %v", err)
	}
}

func TestSetWithTTLAndFlush(t *testing.T) {
	ctx := context.Background()
	a, mr := newAdapter(t)
	defer mr.Close()

	if err := a.SetWithTTL(ctx, "ttl", []byte("x"), time.Minute); err != nil {
		t.Fatal(err)
	}
	if err := a.Flush(ctx); err != nil {
		t.Fatal(err)
	}
	_, err := a.Get(ctx, "ttl")
	if err != cache.ErrNotFound {
		t.Fatalf("Get after flush: %v", err)
	}
}

func TestContextCancel(t *testing.T) {
	a, mr := newAdapter(t)
	defer mr.Close()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := a.Get(ctx, "k"); err == nil {
		t.Fatal("expected context error")
	}
}
