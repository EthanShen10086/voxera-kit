package tiered_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/EthanShen10086/voxera-kit/cache"
	"github.com/EthanShen10086/voxera-kit/cache/local"
	"github.com/EthanShen10086/voxera-kit/cache/memory"
	"github.com/EthanShen10086/voxera-kit/cache/redis"
	"github.com/EthanShen10086/voxera-kit/cache/tiered"
	"github.com/alicebob/miniredis/v2"
)

const benchValueSize = 256

func benchValue() []byte {
	b := make([]byte, benchValueSize)
	for i := range b {
		b[i] = byte('a' + (i % 26))
	}
	return b
}

func setupRedis(t testing.TB) (*redis.Adapter, *miniredis.Miniredis) {
	t.Helper()
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(mr.Close)
	return redis.New(cache.Config{Address: mr.Addr()}), mr
}

func setupLocal(t testing.TB) *local.Adapter {
	t.Helper()
	l1, err := local.New(cache.Config{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = l1.Close() })
	return l1
}

// BenchmarkMemory_Get is the in-process baseline (no network).
func BenchmarkMemory_Get(b *testing.B) {
	ctx := context.Background()
	m := memory.New()
	key := "bench-key"
	val := benchValue()
	_ = m.Set(ctx, key, val)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := m.Get(ctx, key); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkLocal_Get measures ristretto L1 (process-local, concurrent-safe).
func BenchmarkLocal_Get(b *testing.B) {
	ctx := context.Background()
	l1 := setupLocal(b)
	key := "bench-key"
	val := benchValue()
	_ = l1.Set(ctx, key, val)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := l1.Get(ctx, key); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkRedis_Get measures a single network hop (miniredis ≈ local TCP).
func BenchmarkRedis_Get(b *testing.B) {
	ctx := context.Background()
	r, _ := setupRedis(b)
	key := "bench-key"
	val := benchValue()
	_ = r.Set(ctx, key, val)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := r.Get(ctx, key); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkTiered_L1Hit: tiered local+redis with key warmed in L1.
func BenchmarkTiered_L1Hit(b *testing.B) {
	ctx := context.Background()
	l1 := setupLocal(b)
	r, _ := setupRedis(b)
	c, err := tiered.New(l1, r)
	if err != nil {
		b.Fatal(err)
	}
	key := "bench-key"
	val := benchValue()
	_ = c.Set(ctx, key, val)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := c.Get(ctx, key); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkTiered_L2Hit: L1 cold, L2 hit — includes backfill to L1 on first read.
func BenchmarkTiered_L2Hit(b *testing.B) {
	ctx := context.Background()
	l1 := setupLocal(b)
	r, _ := setupRedis(b)
	key := "bench-key"
	val := benchValue()
	_ = r.Set(ctx, key, val)

	c, err := tiered.New(l1, r)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = l1.Flush(ctx)
		if _, err := c.Get(ctx, key); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkTiered_Set measures write-through to both layers.
func BenchmarkTiered_Set(b *testing.B) {
	ctx := context.Background()
	l1 := setupLocal(b)
	r, _ := setupRedis(b)
	c, err := tiered.New(l1, r)
	if err != nil {
		b.Fatal(err)
	}
	val := benchValue()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("k-%d", i%1024)
		if err := c.Set(ctx, key, val); err != nil {
			b.Fatal(err)
		}
	}
}
