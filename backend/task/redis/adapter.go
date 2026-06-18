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
	client        *redis.Client
	queueKey      string
	payloadPrefix string
}

// New creates a new Redis task queue adapter.
func New(cfg Config) *Adapter {
	prefix := cfg.KeyPrefix
	if prefix == "" {
		prefix = "task"
	}
	return &Adapter{
		client:        cfg.Client,
		queueKey:      prefix + ":queue",
		payloadPrefix: prefix + ":payload:",
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
