package channel_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/EthanShen10086/voxera-kit/concurrency"
	"github.com/EthanShen10086/voxera-kit/concurrency/channel"
)

func TestSemaphore(t *testing.T) {
	sem := channel.NewSemaphore(1)
	ctx := context.Background()

	if !sem.TryAcquire() {
		t.Fatal("TryAcquire should succeed")
	}
	if sem.TryAcquire() {
		t.Fatal("TryAcquire should fail when full")
	}
	if sem.Available() != 0 {
		t.Fatalf("available = %d", sem.Available())
	}

	ctxCancel, cancel := context.WithCancel(ctx)
	cancel()
	if err := sem.Acquire(ctxCancel); !errors.Is(err, context.Canceled) {
		t.Fatalf("Acquire: %v", err)
	}

	sem.Release()
	if err := sem.Acquire(ctx); err != nil {
		t.Fatal(err)
	}
}

func TestWorkerPoolSubmitAndShutdown(t *testing.T) {
	pool := channel.NewWorkerPool(concurrency.WorkerPoolConfig{MaxWorkers: 2, QueueSize: 2})
	var ran atomic.Int32

	if err := pool.Submit(func(context.Context) error {
		ran.Add(1)
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	result, err := pool.SubmitWait(ctx, func(context.Context) error { return nil })
	if err != nil {
		t.Fatal(err)
	}
	if result.Duration < 0 {
		t.Fatal("negative duration")
	}

	deadline := time.Now().Add(time.Second)
	for ran.Load() < 1 && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}

	for i := 0; i < 3; i++ {
		_ = pool.Submit(func(context.Context) error { return nil })
	}
	if err := pool.Submit(func(context.Context) error { return nil }); err == nil {
		t.Fatal("expected queue full error")
	}

	if err := pool.Shutdown(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := pool.Submit(func(context.Context) error { return nil }); !errors.Is(err, context.Canceled) {
		t.Fatalf("submit after shutdown: %v", err)
	}
}
