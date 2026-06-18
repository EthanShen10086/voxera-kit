package redis_test

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/EthanShen10086/voxera-kit/task"
	"github.com/EthanShen10086/voxera-kit/task/redis"
	goredis "github.com/redis/go-redis/v9"
)

func newAdapter(t *testing.T) (*redis.Adapter, *miniredis.Miniredis) {
	t.Helper()
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	client := goredis.NewClient(&goredis.Options{Addr: mr.Addr()})
	return redis.New(redis.Config{Client: client, KeyPrefix: "test"}), mr
}

func TestEnqueuePopComplete(t *testing.T) {
	ctx := context.Background()
	a, mr := newAdapter(t)
	defer mr.Close()

	if err := a.Enqueue(ctx, task.Task{ID: "t1", Name: "job"}); err != nil {
		t.Fatal(err)
	}
	tk, ok, err := a.PopDue(ctx)
	if err != nil || !ok || tk.ID != "t1" {
		t.Fatalf("PopDue: %+v ok=%v err=%v", tk, ok, err)
	}
	if err := a.Complete(ctx, tk); err != nil {
		t.Fatal(err)
	}
}

func TestIdempotencyAndCancel(t *testing.T) {
	ctx := context.Background()
	a, mr := newAdapter(t)
	defer mr.Close()

	tk := task.Task{ID: "t2", Name: "idem", IdempotencyKey: "key-1"}
	if err := a.Schedule(ctx, tk, time.Now().Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	if err := a.Enqueue(ctx, tk); err != nil {
		t.Fatal(err)
	}
	if err := a.Cancel(ctx, "t2"); err != nil {
		t.Fatal(err)
	}
	_, ok, err := a.PopDue(ctx)
	if err != nil || ok {
		t.Fatalf("expected empty queue, ok=%v err=%v", ok, err)
	}
}

func TestRequeueOrDLQ(t *testing.T) {
	ctx := context.Background()
	a, mr := newAdapter(t)
	defer mr.Close()

	tk := task.Task{ID: "t3", Name: "retry", Attempt: 2}
	if err := a.RequeueOrDLQ(ctx, tk, context.Canceled, task.RetryPolicy{MaxAttempts: 3, Backoff: time.Millisecond}); err != nil {
		t.Fatal(err)
	}
	tk.Attempt = 3
	if err := a.RequeueOrDLQ(ctx, tk, context.Canceled, task.RetryPolicy{MaxAttempts: 3}); err != nil {
		t.Fatal(err)
	}
	if a.DeadLetterLen() != 1 {
		t.Fatalf("dlq len = %d", a.DeadLetterLen())
	}
}

func TestEnqueueValidation(t *testing.T) {
	a := redis.New(redis.Config{})
	if err := a.Enqueue(context.Background(), task.Task{}); err == nil {
		t.Fatal("expected id required")
	}
}
