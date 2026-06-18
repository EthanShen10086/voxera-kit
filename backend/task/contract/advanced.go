package contract

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/EthanShen10086/voxera-kit/task"
)

// RunTaskAdvancedContract exercises idempotency, retry, and dead-letter behavior.
func RunTaskAdvancedContract(t *testing.T, factory Factory) {
	t.Helper()
	ctx := context.Background()

	t.Run("IdempotencyKey", func(t *testing.T) {
		var runs atomic.Int32
		q, cleanup := factory(t, func(_ context.Context, _ task.Task) error {
			runs.Add(1)
			return nil
		})
		if cleanup != nil {
			defer cleanup()
		}

		tk := task.Task{ID: "idem-1", IdempotencyKey: "order-42"}
		if err := q.Enqueue(ctx, tk); err != nil {
			t.Fatalf("Enqueue() = %v", err)
		}
		dup := task.Task{ID: "idem-2", IdempotencyKey: "order-42"}
		if err := q.Enqueue(ctx, dup); err != nil {
			t.Fatalf("duplicate Enqueue() = %v", err)
		}
		waitUntil(t, func() bool { return runs.Load() == 1 }, 2*time.Second, "idempotent handler")
	})

	t.Run("RetryThenSuccess", func(t *testing.T) {
		var attempts atomic.Int32
		q, cleanup := factory(t, func(_ context.Context, _ task.Task) error {
			if attempts.Add(1) < 2 {
				return errors.New("transient")
			}
			return nil
		})
		if cleanup != nil {
			defer cleanup()
		}

		if err := q.Enqueue(ctx, task.Task{
			ID:    "retry-success",
			Retry: task.RetryPolicy{MaxAttempts: 3, Backoff: 20 * time.Millisecond},
		}); err != nil {
			t.Fatalf("Enqueue() = %v", err)
		}
		waitUntil(t, func() bool { return attempts.Load() >= 2 }, 2*time.Second, "retry success")
	})

	t.Run("DeadLetter", func(t *testing.T) {
		q, cleanup := factory(t, func(_ context.Context, _ task.Task) error {
			return errors.New("permanent failure")
		})
		if cleanup != nil {
			defer cleanup()
		}

		dlq, ok := q.(task.DeadLetterQueue)
		if !ok {
			t.Skip("queue does not implement DeadLetterQueue")
		}

		if err := q.Enqueue(ctx, task.Task{
			ID:    "dlq-1",
			Retry: task.RetryPolicy{MaxAttempts: 2, Backoff: 10 * time.Millisecond},
		}); err != nil {
			t.Fatalf("Enqueue() = %v", err)
		}
		waitUntil(t, func() bool { return dlq.DeadLetterLen() >= 1 }, 2*time.Second, "dead letter")
	})
}
