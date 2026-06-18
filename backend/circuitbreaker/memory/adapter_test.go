package memory

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/EthanShen10086/voxera-kit/circuitbreaker"
)

func TestCircuitBreakerOpensAfterFailures(t *testing.T) {
	cb := New(circuitbreaker.Config{MaxFailures: 2, Timeout: time.Hour})
	ctx := context.Background()
	fail := func() error { return errors.New("boom") }
	_ = cb.Execute(ctx, fail)
	_ = cb.Execute(ctx, fail)
	err := cb.Execute(ctx, fail)
	if !errors.Is(err, circuitbreaker.ErrCircuitOpen) {
		t.Fatalf("expected open circuit, got %v state=%v", err, cb.State())
	}
}

func TestCircuitBreakerClosesOnSuccessFromHalfOpen(t *testing.T) {
	cb := New(circuitbreaker.Config{MaxFailures: 1, Timeout: time.Millisecond, HalfOpenMaxCalls: 2})
	ctx := context.Background()
	_ = cb.Execute(ctx, func() error { return errors.New("x") })
	time.Sleep(2 * time.Millisecond)
	if err := cb.Execute(ctx, func() error { return nil }); err != nil {
		t.Fatal(err)
	}
	if cb.State() != circuitbreaker.Closed {
		t.Fatalf("state %v", cb.State())
	}
}

func TestCircuitBreakerResetAndCounts(t *testing.T) {
	cb := New(circuitbreaker.Config{MaxFailures: 5})
	ctx := context.Background()
	_ = cb.Execute(ctx, func() error { return errors.New("x") })
	_, failures := cb.Counts()
	if failures == 0 {
		t.Fatal("expected failure count")
	}
	cb.Reset()
	if cb.State() != circuitbreaker.Closed {
		t.Fatalf("state after reset = %v", cb.State())
	}
}
