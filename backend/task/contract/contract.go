package contract

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/EthanShen10086/voxera-kit/task"
)

// Factory creates a task queue wired to the given handler.
type Factory func(t *testing.T, handler task.Handler) (task.TaskQueue, func())

// RunTaskContract exercises Enqueue, Schedule, and Cancel behavior.
func RunTaskContract(t *testing.T, factory Factory) {
	t.Helper()
	ctx := context.Background()

	t.Run("Enqueue", func(t *testing.T) {
		var ran atomic.Bool
		q, cleanup := factory(t, func(_ context.Context, tk task.Task) error {
			if tk.ID == "enqueue-1" {
				ran.Store(true)
			}
			return nil
		})
		if cleanup != nil {
			defer cleanup()
		}

		if err := q.Enqueue(ctx, task.Task{ID: "enqueue-1", Name: "job"}); err != nil {
			t.Fatalf("Enqueue() = %v", err)
		}
		waitUntil(t, func() bool { return ran.Load() }, 2*time.Second, "task handler")
	})

	t.Run("ScheduleCancel", func(t *testing.T) {
		var ran atomic.Bool
		q, cleanup := factory(t, func(_ context.Context, _ task.Task) error {
			ran.Store(true)
			return nil
		})
		if cleanup != nil {
			defer cleanup()
		}

		runAt := time.Now().Add(500 * time.Millisecond)
		if err := q.Schedule(ctx, task.Task{ID: "schedule-cancel"}, runAt); err != nil {
			t.Fatalf("Schedule() = %v", err)
		}
		if err := q.Cancel(ctx, "schedule-cancel"); err != nil {
			t.Fatalf("Cancel() = %v", err)
		}

		time.Sleep(700 * time.Millisecond)
		if ran.Load() {
			t.Fatal("canceled task should not run")
		}
	})

	t.Run("ScheduleRun", func(t *testing.T) {
		var ran atomic.Bool
		q, cleanup := factory(t, func(_ context.Context, tk task.Task) error {
			if tk.ID == "schedule-run" {
				ran.Store(true)
			}
			return nil
		})
		if cleanup != nil {
			defer cleanup()
		}

		runAt := time.Now().Add(100 * time.Millisecond)
		if err := q.Schedule(ctx, task.Task{ID: "schedule-run"}, runAt); err != nil {
			t.Fatalf("Schedule() = %v", err)
		}
		waitUntil(t, func() bool { return ran.Load() }, 2*time.Second, "scheduled task")
	})
}

func waitUntil(t *testing.T, cond func() bool, timeout time.Duration, label string) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for !cond() && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if !cond() {
		t.Fatalf("timed out waiting for %s", label)
	}
}
