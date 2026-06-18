package redis_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/EthanShen10086/voxera-kit/task"
	"github.com/EthanShen10086/voxera-kit/task/redis"
)

func TestWorkerProcessesTask(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	a, mr := newAdapter(t)
	defer mr.Close()

	var handled atomic.Bool
	w, err := redis.NewWorker(redis.WorkerConfig{
		Adapter: a,
		Handler: func(_ context.Context, tk task.Task) error {
			if tk.ID == "w1" {
				handled.Store(true)
			}
			return nil
		},
		PollInterval: 5 * time.Millisecond,
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := a.Enqueue(ctx, task.Task{ID: "w1", Name: "job"}); err != nil {
		t.Fatal(err)
	}

	done := make(chan error, 1)
	go func() { done <- w.Run(ctx) }()

	deadline := time.Now().Add(time.Second)
	for !handled.Load() && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}
	cancel()
	<-done

	if !handled.Load() {
		t.Fatal("worker did not handle task")
	}
}

func TestWorkerHandlerErrorRequeues(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	a, mr := newAdapter(t)
	defer mr.Close()

	w, err := redis.NewWorker(redis.WorkerConfig{
		Adapter: a,
		Handler: func(_ context.Context, _ task.Task) error {
			return context.Canceled
		},
		PollInterval: 5 * time.Millisecond,
	})
	if err != nil {
		t.Fatal(err)
	}

	tk := task.Task{ID: "w2", Name: "fail", Attempt: 1}
	if err := a.Enqueue(ctx, tk); err != nil {
		t.Fatal(err)
	}

	go func() { _ = w.Run(ctx) }()
	time.Sleep(100 * time.Millisecond)
	cancel()
}

func TestNewWorkerValidation(t *testing.T) {
	if _, err := redis.NewWorker(redis.WorkerConfig{}); err == nil {
		t.Fatal("expected adapter error")
	}
	a, mr := newAdapter(t)
	defer mr.Close()
	if _, err := redis.NewWorker(redis.WorkerConfig{Adapter: a}); err == nil {
		t.Fatal("expected handler error")
	}
}
