// Package memory provides tests for the in-memory task queue adapter.
package memory_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/EthanShen10086/voxera-kit/task"
	"github.com/EthanShen10086/voxera-kit/task/contract"
	"github.com/EthanShen10086/voxera-kit/task/memory"
)

func TestAdapterEnqueueAndRun(t *testing.T) {
	var ran atomic.Bool
	q := memory.New(memory.Config{
		Handler: func(_ context.Context, tk task.Task) error {
			if tk.ID == "t1" {
				ran.Store(true)
			}
			return nil
		},
	})

	ctx := context.Background()
	if err := q.Enqueue(ctx, task.Task{ID: "t1", Name: "job"}); err != nil {
		t.Fatalf("Enqueue() = %v", err)
	}

	deadline := time.Now().Add(2 * time.Second)
	for !ran.Load() && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if !ran.Load() {
		t.Fatal("task handler was not invoked")
	}
}

func TestAdapterScheduleCancel(t *testing.T) {
	var ran atomic.Bool
	q := memory.New(memory.Config{
		Handler: func(_ context.Context, _ task.Task) error {
			ran.Store(true)
			return nil
		},
	})

	ctx := context.Background()
	runAt := time.Now().Add(500 * time.Millisecond)
	if err := q.Schedule(ctx, task.Task{ID: "t2"}, runAt); err != nil {
		t.Fatalf("Schedule() = %v", err)
	}
	if q.Pending() == 0 {
		t.Fatal("expected pending scheduled task")
	}
	if err := q.Cancel(ctx, "t2"); err != nil {
		t.Fatalf("Cancel() = %v", err)
	}
	if q.Pending() != 0 {
		t.Fatal("expected no pending after cancel")
	}

	time.Sleep(700 * time.Millisecond)
	if ran.Load() {
		t.Fatal("canceled task should not run")
	}
}

func TestAdapterStopAndDeadLetter(t *testing.T) {
	q := memory.New(memory.Config{
		Handler: func(_ context.Context, _ task.Task) error {
			return errors.New("fail")
		},
	})
	ctx := context.Background()
	if err := q.Enqueue(ctx, task.Task{
		ID: "dlq1", Name: "job",
		Retry: task.RetryPolicy{MaxAttempts: 1, Backoff: time.Millisecond},
	}); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(2 * time.Second)
	for q.DeadLetterLen() == 0 && time.Now().Before(deadline) {
		time.Sleep(20 * time.Millisecond)
	}
	if q.DeadLetterLen() == 0 {
		t.Fatal("expected dead letter task")
	}
	tasks := q.DeadLetterTasks()
	if len(tasks) != 1 || tasks[0].ID != "dlq1" {
		t.Fatalf("DeadLetterTasks() = %#v", tasks)
	}
	q.Stop()
}

func TestTaskContract_Memory(t *testing.T) {
	contract.RunTaskContract(t, func(t *testing.T, handler task.Handler) (task.TaskQueue, func()) {
		q := memory.New(memory.Config{Handler: handler})
		return q, nil
	})
}

func TestTaskAdvancedContract_Memory(t *testing.T) {
	contract.RunTaskAdvancedContract(t, func(t *testing.T, handler task.Handler) (task.TaskQueue, func()) {
		q := memory.New(memory.Config{Handler: handler})
		return q, nil
	})
}
