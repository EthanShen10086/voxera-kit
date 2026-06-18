// Package memory provides an in-memory implementation of the task.TaskQueue interface
// using goroutines and time.AfterFunc for delayed execution.
package memory

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/EthanShen10086/voxera-kit/task"
)

type pendingEntry struct {
	task   task.Task
	timer  *time.Timer
	cancel context.CancelFunc
}

// Config holds configuration for the in-memory task queue.
type Config struct {
	// Handler is invoked when a task becomes due. Optional for producer-only usage.
	Handler task.Handler
}

// Adapter is an in-memory delayed task queue with idempotency, retry, and DLQ.
type Adapter struct {
	mu           sync.Mutex
	pending      map[string]*pendingEntry
	idempotency  map[string]string // key -> task ID
	processed    map[string]struct{}
	dlq          []task.Task
	handler      task.Handler
	stopped      bool
	defaultRetry task.RetryPolicy
}

// New creates a new in-memory task queue.
func New(cfg Config) *Adapter {
	return &Adapter{
		pending:     make(map[string]*pendingEntry),
		idempotency: make(map[string]string),
		processed:   make(map[string]struct{}),
		handler:     cfg.Handler,
		defaultRetry: task.RetryPolicy{
			MaxAttempts: 3,
			Backoff:     100 * time.Millisecond,
		},
	}
}

// Enqueue adds a task for immediate execution.
func (a *Adapter) Enqueue(ctx context.Context, t task.Task) error {
	return a.Schedule(ctx, t, time.Now())
}

// Schedule adds a task to run at the specified time.
func (a *Adapter) Schedule(_ context.Context, t task.Task, runAt time.Time) error {
	if t.ID == "" {
		return fmt.Errorf("task: id is required")
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	if a.stopped {
		return fmt.Errorf("task: queue stopped")
	}
	if t.IdempotencyKey != "" {
		if _, ok := a.processed[t.IdempotencyKey]; ok {
			return nil
		}
		if existing, ok := a.idempotency[t.IdempotencyKey]; ok && existing != t.ID {
			return nil
		}
		a.idempotency[t.IdempotencyKey] = t.ID
	}
	if _, exists := a.pending[t.ID]; exists {
		return fmt.Errorf("task: %q already scheduled", t.ID)
	}

	delay := time.Until(runAt)
	if delay < 0 {
		delay = 0
	}

	runCtx, cancel := context.WithCancel(context.Background())
	entry := &pendingEntry{
		task:   t,
		cancel: cancel,
	}

	taskCopy := t
	entry.timer = time.AfterFunc(delay, func() {
		a.runTask(runCtx, taskCopy)
	})
	a.pending[t.ID] = entry
	return nil
}

func (a *Adapter) runTask(ctx context.Context, t task.Task) {
	a.mu.Lock()
	delete(a.pending, t.ID)
	handler := a.handler
	retry := t.Retry
	if retry.MaxAttempts == 0 {
		retry = a.defaultRetry
	}
	if retry.Backoff == 0 {
		retry.Backoff = a.defaultRetry.Backoff
	}
	a.mu.Unlock()

	if handler == nil {
		return
	}

	attempt := t.Attempt
	if attempt == 0 {
		attempt = 1
	}
	t.Attempt = attempt

	err := handler(ctx, t)
	if err == nil {
		a.mu.Lock()
		if t.IdempotencyKey != "" {
			a.processed[t.IdempotencyKey] = struct{}{}
		}
		a.mu.Unlock()
		return
	}

	if attempt >= retry.MaxAttempts {
		a.mu.Lock()
		a.dlq = append(a.dlq, t)
		a.mu.Unlock()
		return
	}

	next := t
	next.Attempt = attempt + 1
	_ = a.Schedule(context.Background(), next, time.Now().Add(retry.Backoff))
}

// Cancel removes a pending task by ID.
func (a *Adapter) Cancel(_ context.Context, id string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	entry, ok := a.pending[id]
	if !ok {
		return fmt.Errorf("task: %q not found", id)
	}

	entry.timer.Stop()
	entry.cancel()
	delete(a.pending, id)
	return nil
}

// Stop cancels all pending tasks and prevents new scheduling.
func (a *Adapter) Stop() {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.stopped = true
	for id, entry := range a.pending {
		entry.timer.Stop()
		entry.cancel()
		delete(a.pending, id)
	}
}

// Pending returns the number of tasks waiting to run.
func (a *Adapter) Pending() int {
	a.mu.Lock()
	defer a.mu.Unlock()
	return len(a.pending)
}

// DeadLetterLen returns the number of tasks in the dead-letter queue.
func (a *Adapter) DeadLetterLen() int {
	a.mu.Lock()
	defer a.mu.Unlock()
	return len(a.dlq)
}

// DeadLetterTasks returns a copy of dead-letter tasks (for tests).
func (a *Adapter) DeadLetterTasks() []task.Task {
	a.mu.Lock()
	defer a.mu.Unlock()
	out := make([]task.Task, len(a.dlq))
	copy(out, a.dlq)
	return out
}
