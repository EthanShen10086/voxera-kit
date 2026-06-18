// Package task defines the port interface for delayed and immediate task queues.
// It abstracts away the underlying queue implementation, allowing different
// backends (in-memory, Redis) to be used interchangeably.
package task

import (
	"context"
	"time"
)

// RetryPolicy controls automatic retries when a handler returns an error.
type RetryPolicy struct {
	// MaxAttempts is the total number of execution attempts (including the first).
	// Zero defaults to 3.
	MaxAttempts int
	// Backoff is the delay before each retry. Zero defaults to 100ms.
	Backoff time.Duration
}

// Task represents a unit of work to be executed by a queue consumer.
type Task struct {
	// ID is the unique identifier for the task.
	ID string
	// Name is a human-readable label for the task.
	Name string
	// Payload holds opaque task data for the consumer.
	Payload []byte
	// IdempotencyKey deduplicates enqueue when non-empty.
	IdempotencyKey string
	// Retry configures handler failure retries. Zero value uses defaults in the worker.
	Retry RetryPolicy
	// Attempt is the 1-based execution attempt (set by the worker on retry).
	Attempt int
}

// Handler processes a task when it becomes due.
type Handler func(ctx context.Context, t Task) error

// TaskQueue manages enqueueing, scheduling, and cancellation of tasks.
//
//nolint:revive // TaskQueue is the established port name across adapters.
type TaskQueue interface {
	// Enqueue adds a task for immediate execution.
	Enqueue(ctx context.Context, t Task) error
	// Schedule adds a task to run at the specified time.
	Schedule(ctx context.Context, t Task, runAt time.Time) error
	// Cancel removes a pending task by ID.
	Cancel(ctx context.Context, id string) error
}

// DeadLetterQueue exposes tasks that exhausted retries (testing and ops).
type DeadLetterQueue interface {
	// DeadLetterLen returns the number of tasks in the dead-letter queue.
	DeadLetterLen() int
}
