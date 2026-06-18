// Package redis provides a Redis ZSET-based implementation of the task.TaskQueue interface.
package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/EthanShen10086/voxera-kit/task"
	"github.com/redis/go-redis/v9"
)

// Config holds configuration for the Redis task queue.
type Config struct {
	// Client is the Redis client used for queue operations.
	Client *redis.Client
	// KeyPrefix prefixes all Redis keys. Defaults to "task".
	KeyPrefix string
}

// Adapter implements task.TaskQueue using a sorted set for scheduling.
type Adapter struct {
	client         *redis.Client
	queueKey       string
	payloadPrefix  string
	idempotencyKey string
	processedKey   string
	dlqKey         string
}

// New creates a new Redis task queue adapter.
func New(cfg Config) *Adapter {
	prefix := cfg.KeyPrefix
	if prefix == "" {
		prefix = "task"
	}
	return &Adapter{
		client:         cfg.Client,
		queueKey:       prefix + ":queue",
		payloadPrefix:  prefix + ":payload:",
		idempotencyKey: prefix + ":idem:",
		processedKey:   prefix + ":processed:",
		dlqKey:         prefix + ":dlq",
	}
}

// Enqueue adds a task for immediate execution.
func (a *Adapter) Enqueue(ctx context.Context, t task.Task) error {
	return a.Schedule(ctx, t, time.Now())
}

// Schedule adds a task to run at the specified time.
func (a *Adapter) Schedule(ctx context.Context, t task.Task, runAt time.Time) error {
	if t.ID == "" {
		return fmt.Errorf("task: id is required")
	}
	if a.client == nil {
		return fmt.Errorf("redis: client is nil")
	}

	if t.IdempotencyKey != "" {
		processed := a.processedKey + t.IdempotencyKey
		exists, err := a.client.Exists(ctx, processed).Result()
		if err != nil {
			return err
		}
		if exists > 0 {
			return nil
		}
		idemKey := a.idempotencyKey + t.IdempotencyKey
		ok, err := a.client.SetNX(ctx, idemKey, t.ID, 0).Result()
		if err != nil {
			return err
		}
		if !ok {
			return nil
		}
	}

	data, err := json.Marshal(t)
	if err != nil {
		return fmt.Errorf("redis: marshal task: %w", err)
	}

	pipe := a.client.Pipeline()
	pipe.Set(ctx, a.payloadKey(t.ID), data, 0)
	pipe.ZAdd(ctx, a.queueKey, redis.Z{
		Score:  float64(runAt.UnixMilli()),
		Member: t.ID,
	})
	_, err = pipe.Exec(ctx)
	return err
}

func (a *Adapter) payloadKey(id string) string {
	return a.payloadPrefix + id
}

// Cancel removes a pending task by ID.
func (a *Adapter) Cancel(ctx context.Context, id string) error {
	if a.client == nil {
		return fmt.Errorf("redis: client is nil")
	}

	pipe := a.client.Pipeline()
	pipe.ZRem(ctx, a.queueKey, id)
	pipe.Del(ctx, a.payloadKey(id))
	_, err := pipe.Exec(ctx)
	return err
}

// DeadLetterLen returns the length of the dead-letter list.
func (a *Adapter) DeadLetterLen() int {
	ctx := context.Background()
	n, err := a.client.LLen(ctx, a.dlqKey).Result()
	if err != nil {
		return 0
	}
	return int(n)
}

// PopDue returns the next due task ID and payload, removing it from the queue.
func (a *Adapter) PopDue(ctx context.Context) (task.Task, bool, error) {
	now := float64(time.Now().UnixMilli())
	ids, err := a.client.ZRangeArgs(ctx, redis.ZRangeArgs{
		Key:     a.queueKey,
		Start:   "-inf",
		Stop:    fmt.Sprintf("%f", now),
		ByScore: true,
		Count:   1,
	}).Result()
	if err != nil {
		return task.Task{}, false, err
	}
	if len(ids) == 0 {
		return task.Task{}, false, nil
	}
	id := ids[0]

	removed, err := a.client.ZRem(ctx, a.queueKey, id).Result()
	if err != nil {
		return task.Task{}, false, err
	}
	if removed == 0 {
		return task.Task{}, false, nil
	}

	data, err := a.client.Get(ctx, a.payloadKey(id)).Bytes()
	if err != nil {
		return task.Task{}, false, err
	}
	var tk task.Task
	if err := json.Unmarshal(data, &tk); err != nil {
		return task.Task{}, false, err
	}
	return tk, true, nil
}

// Complete removes payload and marks idempotency processed after success.
func (a *Adapter) Complete(ctx context.Context, tk task.Task) error {
	pipe := a.client.Pipeline()
	pipe.Del(ctx, a.payloadKey(tk.ID))
	if tk.IdempotencyKey != "" {
		pipe.Set(ctx, a.processedKey+tk.IdempotencyKey, "1", 0)
	}
	_, err := pipe.Exec(ctx)
	return err
}

// RequeueOrDLQ reschedules on retry or pushes to DLQ when attempts exhausted.
func (a *Adapter) RequeueOrDLQ(ctx context.Context, tk task.Task, handlerErr error, retry task.RetryPolicy) error {
	if retry.MaxAttempts == 0 {
		retry.MaxAttempts = 3
	}
	if retry.Backoff == 0 {
		retry.Backoff = 100 * time.Millisecond
	}
	attempt := tk.Attempt
	if attempt == 0 {
		attempt = 1
	}

	if attempt >= retry.MaxAttempts {
		data, err := json.Marshal(tk)
		if err != nil {
			return err
		}
		return a.client.RPush(ctx, a.dlqKey, data).Err()
	}

	next := tk
	next.Attempt = attempt + 1
	return a.Schedule(ctx, next, time.Now().Add(retry.Backoff))
}

// PushDLQError wraps handler errors for worker logging.
func PushDLQError(err error) error {
	return fmt.Errorf("task handler: %w", err)
}
