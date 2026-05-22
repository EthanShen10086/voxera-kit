// Package concurrency defines the port interfaces for semaphore and worker pool primitives.
// It provides abstractions for controlling concurrent access to resources and
// managing pools of goroutines for parallel task execution.
package concurrency

import (
	"context"
	"time"
)

// Semaphore controls concurrent access to a finite number of resources.
// Implementations must be safe for concurrent use.
type Semaphore interface {
	// Acquire blocks until a resource slot is available or the context is canceled.
	Acquire(ctx context.Context) error
	// TryAcquire attempts to acquire a slot without blocking, returning true on success.
	TryAcquire() bool
	// Release returns one slot to the semaphore.
	Release()
	// Available returns the number of currently free slots.
	Available() int
}

// Task is a unit of work submitted to a worker pool.
type Task func(ctx context.Context) error

// TaskResult holds the outcome of an executed task.
type TaskResult struct {
	// Error is the error returned by the task, or nil on success.
	Error error
	// Duration is the wall-clock time the task took to execute.
	Duration time.Duration
}

// WorkerPool manages a pool of goroutines that execute submitted tasks.
// Implementations must be safe for concurrent use.
type WorkerPool interface {
	// Submit enqueues a task for asynchronous execution.
	// Returns an error if the pool is shut down or the queue is full.
	Submit(task Task) error
	// SubmitWait enqueues a task and blocks until it completes or the context is canceled.
	SubmitWait(ctx context.Context, task Task) (TaskResult, error)
	// Running returns the number of tasks currently being executed.
	Running() int
	// Pending returns the number of tasks waiting in the queue.
	Pending() int
	// Shutdown gracefully stops the pool, waiting for in-flight tasks to finish
	// or the context to be canceled.
	Shutdown(ctx context.Context) error
}

// WorkerPoolConfig holds the parameters for constructing a worker pool.
type WorkerPoolConfig struct {
	// MaxWorkers is the maximum number of concurrent worker goroutines.
	MaxWorkers int
	// QueueSize is the capacity of the pending task queue.
	QueueSize int
	// IdleTimeout is how long an idle worker stays alive before being reclaimed.
	IdleTimeout time.Duration
}
