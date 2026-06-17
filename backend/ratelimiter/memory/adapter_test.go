package memory

import (
	"context"
	"testing"

	"github.com/EthanShen10086/voxera-kit/ratelimiter"
)

func TestRateLimiterBurst(t *testing.T) {
	rl := New(ratelimiter.Config{Rate: 1, Burst: 2})
	ctx := context.Background()
	for i := 0; i < 2; i++ {
		ok, err := rl.Allow(ctx, "user1")
		if err != nil || !ok {
			t.Fatalf("burst %d: ok=%v err=%v", i, ok, err)
		}
	}
	ok, _ := rl.Allow(ctx, "user1")
	if ok {
		t.Fatal("third call should be denied")
	}
}

func TestRateLimiterReset(t *testing.T) {
	rl := New(ratelimiter.Config{Rate: 0.01, Burst: 1})
	ctx := context.Background()
	_, _ = rl.Allow(ctx, "k")
	_ = rl.Reset(ctx, "k")
	ok, _ := rl.Allow(ctx, "k")
	if !ok {
		t.Fatal("expected allow after reset")
	}
}
