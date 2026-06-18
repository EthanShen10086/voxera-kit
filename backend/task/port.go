// Package task defines the port interface for delayed and immediate task queues.
// It abstracts away the underlying queue implementation, allowing different
// backends (in-memory, Redis) to be used interchangeably.
package task

import (
	"context"
	"time"
)

// Task represents a unit of work to be executed by a queue consumer.
type Task struct {
	// ID is the unique identifier for the task.
	ID string
	// Name is a human-readable label for the task.
	Name string
	// Payload holds opaque task data for the consumer.
	Payload []byte
}

// Handler processes a task when it becomes due.
type Handler func(ctx context.Context, t Task) error

// TaskQueue manages enqueueing, scheduling, and cancellation of tasks.
type TaskQueue interface {
	// Enqueue adds a task for immediate execution.
	Enqueue(ctx context.Context, t Task) error
	// Schedule adds a task to run at the specified time.
	Schedule(ctx context.Context, t Task, runAt time.Time) error
	// Cancel removes a pending task by ID.
	Cancel(ctx context.Context, id string) error
}
