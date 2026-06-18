package channel_test

import (
	"context"
	"errors"
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
	release := make(chan struct{})

	block := func(context.Context) error {
		<-release
		return nil
	}

	// Saturate workers first, then fill the queue buffer.
	for i := 0; i < 2; i++ {
		if err := pool.Submit(block); err != nil {
			t.Fatalf("submit worker %d: %v", i, err)
		}
	}
	deadline := time.Now().Add(time.Second)
	for pool.Running() < 2 && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	for i := 0; i < 2; i++ {
		if err := pool.Submit(block); err != nil {
			t.Fatalf("submit queued %d: %v", i, err)
		}
	}

	if err := pool.Submit(block); err == nil {
		t.Fatal("expected queue full error")
	}

	close(release)

	if err := pool.Shutdown(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := pool.Submit(block); !errors.Is(err, context.Canceled) {
		t.Fatalf("submit after shutdown: %v", err)
	}
}

func TestWorkerPoolSubmitWait(t *testing.T) {
	pool := channel.NewWorkerPool(concurrency.WorkerPoolConfig{MaxWorkers: 1, QueueSize: 1})
	defer func() { _ = pool.Shutdown(context.Background()) }()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	result, err := pool.SubmitWait(ctx, func(context.Context) error { return nil })
	if err != nil {
		t.Fatal(err)
	}
	if result.Duration < 0 {
		t.Fatal("negative duration")
	}
}
