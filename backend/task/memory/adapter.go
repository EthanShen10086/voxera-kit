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

// Adapter is an in-memory delayed task queue.
type Adapter struct {
	mu      sync.Mutex
	pending map[string]*pendingEntry
	handler task.Handler
	stopped bool
}

// New creates a new in-memory task queue.
func New(cfg Config) *Adapter {
	return &Adapter{
		pending: make(map[string]*pendingEntry),
		handler: cfg.Handler,
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
		a.mu.Lock()
		delete(a.pending, taskCopy.ID)
		a.mu.Unlock()
		cancel()

		if a.handler != nil {
			_ = a.handler(runCtx, taskCopy)
		}
	})
	a.pending[t.ID] = entry
	return nil
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
