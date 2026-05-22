// Package channel provides channel-based implementations of the concurrency.Semaphore
// and concurrency.WorkerPool interfaces.
package channel

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/EthanShen10086/voxera-kit/concurrency"
)

// Semaphore implements concurrency.Semaphore using a buffered channel.
type Semaphore struct {
	ch chan struct{}
}

// NewSemaphore creates a new channel-based semaphore with the given capacity.
func NewSemaphore(n int) *Semaphore {
	return &Semaphore{ch: make(chan struct{}, n)}
}

// Acquire blocks until a resource slot is available or the context is canceled.
func (s *Semaphore) Acquire(ctx context.Context) error {
	select {
	case s.ch <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// TryAcquire attempts to acquire a slot without blocking.
func (s *Semaphore) TryAcquire() bool {
	select {
	case s.ch <- struct{}{}:
		return true
	default:
		return false
	}
}

// Release returns one slot to the semaphore.
func (s *Semaphore) Release() {
	<-s.ch
}

// Available returns the number of currently free slots.
func (s *Semaphore) Available() int {
	return cap(s.ch) - len(s.ch)
}

// WorkerPool implements concurrency.WorkerPool using goroutines and channels.
type WorkerPool struct {
	cfg      concurrency.WorkerPoolConfig
	tasks    chan concurrency.Task
	running  atomic.Int64
	pending  atomic.Int64
	quit     chan struct{}
	once     sync.Once
	wg       sync.WaitGroup
	shutdown atomic.Bool
}

// NewWorkerPool creates a new channel-based worker pool and starts the worker goroutines.
func NewWorkerPool(cfg concurrency.WorkerPoolConfig) *WorkerPool {
	p := &WorkerPool{
		cfg:   cfg,
		tasks: make(chan concurrency.Task, cfg.QueueSize),
		quit:  make(chan struct{}),
	}
	for i := 0; i < cfg.MaxWorkers; i++ {
		p.wg.Add(1)
		go p.worker()
	}
	return p
}

func (p *WorkerPool) worker() {
	defer p.wg.Done()
	for {
		select {
		case task, ok := <-p.tasks:
			if !ok {
				return
			}
			p.pending.Add(-1)
			p.running.Add(1)
			_ = task(context.Background())
			p.running.Add(-1)
		case <-p.quit:
			return
		}
	}
}

// Submit enqueues a task for asynchronous execution.
func (p *WorkerPool) Submit(task concurrency.Task) error {
	if p.shutdown.Load() {
		return context.Canceled
	}
	p.pending.Add(1)
	select {
	case p.tasks <- task:
		return nil
	default:
		p.pending.Add(-1)
		return context.DeadlineExceeded
	}
}

// SubmitWait enqueues a task and blocks until it completes or the context is canceled.
func (p *WorkerPool) SubmitWait(ctx context.Context, task concurrency.Task) (concurrency.TaskResult, error) {
	done := make(chan concurrency.TaskResult, 1)
	wrapped := func(taskCtx context.Context) error {
		start := time.Now()
		err := task(taskCtx)
		done <- concurrency.TaskResult{Error: err, Duration: time.Since(start)}
		return err
	}
	if err := p.Submit(wrapped); err != nil {
		return concurrency.TaskResult{}, err
	}
	select {
	case result := <-done:
		return result, nil
	case <-ctx.Done():
		return concurrency.TaskResult{}, ctx.Err()
	}
}

// Running returns the number of tasks currently being executed.
func (p *WorkerPool) Running() int {
	return int(p.running.Load())
}

// Pending returns the number of tasks waiting in the queue.
func (p *WorkerPool) Pending() int {
	return int(p.pending.Load())
}

// Shutdown gracefully stops the pool, waiting for in-flight tasks to finish
// or the context to be canceled.
func (p *WorkerPool) Shutdown(ctx context.Context) error {
	p.shutdown.Store(true)
	p.once.Do(func() { close(p.quit) })

	done := make(chan struct{})
	go func() {
		p.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
