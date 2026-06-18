// Package memory provides tests for the in-memory task queue adapter.
package memory_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/EthanShen10086/voxera-kit/task"
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
	if err := q.Cancel(ctx, "t2"); err != nil {
		t.Fatalf("Cancel() = %v", err)
	}

	time.Sleep(700 * time.Millisecond)
	if ran.Load() {
		t.Fatal("cancelled task should not run")
	}
}
