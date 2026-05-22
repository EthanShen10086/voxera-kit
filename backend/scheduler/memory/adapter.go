// Package memory provides a lightweight in-memory implementation of the scheduler.Scheduler
// interface using time.Timer and goroutines for job scheduling.
//
// This adapter is intended as a fallback when the full cron adapter is not available.
// It supports "@every <duration>" syntax (e.g. "@every 30s", "@every 5m") for interval-based
// scheduling. Standard five-field cron expressions are not fully parsed; they fall back to a
// default interval of one minute. For production use with real cron expressions, prefer the
// cron adapter in the sibling "cron" package.
package memory

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/EthanShen10086/voxera-kit/scheduler"
)

type jobEntry struct {
	job      scheduler.Job
	cronExpr string
	info     scheduler.JobInfo
	timer    *time.Timer
	cancel   context.CancelFunc
}

// Adapter is an in-memory cron-style scheduler.
type Adapter struct {
	mu        sync.RWMutex
	jobs      map[string]*jobEntry
	running   atomic.Bool
	config    scheduler.Config
	cancel    context.CancelFunc
	semaphore chan struct{}
}

// New creates a new in-memory scheduler with the given configuration.
func New(config scheduler.Config) *Adapter {
	maxConcurrent := config.MaxConcurrent
	if maxConcurrent <= 0 {
		maxConcurrent = 10
	}
	return &Adapter{
		jobs:      make(map[string]*jobEntry),
		config:    config,
		semaphore: make(chan struct{}, maxConcurrent),
	}
}

// Register adds a job with the given cron expression.
func (a *Adapter) Register(name string, cronExpr string, job scheduler.Job) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if _, exists := a.jobs[name]; exists {
		return fmt.Errorf("scheduler: job %q already registered", name)
	}

	entry := &jobEntry{
		job:      job,
		cronExpr: cronExpr,
		info: scheduler.JobInfo{
			ID:       name,
			Name:     job.Name(),
			CronExpr: cronExpr,
			Status:   scheduler.Pending,
		},
	}
	a.jobs[name] = entry
	return nil
}

// Unregister removes a previously registered job by name.
func (a *Adapter) Unregister(name string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	entry, exists := a.jobs[name]
	if !exists {
		return fmt.Errorf("scheduler: job %q not found", name)
	}

	if entry.timer != nil {
		entry.timer.Stop()
	}
	if entry.cancel != nil {
		entry.cancel()
	}
	delete(a.jobs, name)
	return nil
}

// Start begins the scheduler's event loop.
func (a *Adapter) Start(ctx context.Context) error {
	if a.running.Load() {
		return fmt.Errorf("scheduler: already running")
	}

	ctx, cancel := context.WithCancel(ctx)
	a.cancel = cancel
	a.running.Store(true)

	a.mu.RLock()
	for _, entry := range a.jobs {
		a.scheduleNext(ctx, entry)
	}
	a.mu.RUnlock()

	return nil
}

// Stop gracefully shuts down the scheduler.
func (a *Adapter) Stop(_ context.Context) error {
	if !a.running.Load() {
		return fmt.Errorf("scheduler: not running")
	}

	a.cancel()
	a.running.Store(false)

	a.mu.Lock()
	for _, entry := range a.jobs {
		if entry.timer != nil {
			entry.timer.Stop()
		}
	}
	a.mu.Unlock()

	return nil
}

// RunNow triggers immediate execution of the named job.
func (a *Adapter) RunNow(name string) error {
	a.mu.RLock()
	entry, exists := a.jobs[name]
	a.mu.RUnlock()

	if !exists {
		return fmt.Errorf("scheduler: job %q not found", name)
	}

	go a.executeJob(context.Background(), entry)
	return nil
}

// List returns information about all registered jobs.
func (a *Adapter) List() []scheduler.JobInfo {
	a.mu.RLock()
	defer a.mu.RUnlock()

	infos := make([]scheduler.JobInfo, 0, len(a.jobs))
	for _, entry := range a.jobs {
		infos = append(infos, entry.info)
	}
	return infos
}

// IsRunning reports whether the scheduler is currently active.
func (a *Adapter) IsRunning() bool {
	return a.running.Load()
}

// nextRunDuration parses the cron expression to determine the interval between runs.
// It supports "@every <duration>" syntax (e.g. "@every 30s", "@every 5m"). For standard
// cron expressions that cannot be parsed as an interval, it falls back to one minute.
func (a *Adapter) nextRunDuration(cronExpr string) time.Duration {
	const prefix = "@every "
	if len(cronExpr) > len(prefix) && cronExpr[:len(prefix)] == prefix {
		d, err := time.ParseDuration(cronExpr[len(prefix):])
		if err == nil && d > 0 {
			return d
		}
	}
	return time.Minute
}

func (a *Adapter) scheduleNext(ctx context.Context, entry *jobEntry) {
	dur := a.nextRunDuration(entry.cronExpr)
	entry.info.NextRun = time.Now().Add(dur)

	entry.timer = time.AfterFunc(dur, func() {
		select {
		case <-ctx.Done():
			return
		default:
			a.executeJob(ctx, entry)
			if a.running.Load() {
				a.mu.RLock()
				a.scheduleNext(ctx, entry)
				a.mu.RUnlock()
			}
		}
	})
}

func (a *Adapter) executeJob(ctx context.Context, entry *jobEntry) {
	a.semaphore <- struct{}{}
	defer func() { <-a.semaphore }()

	if a.config.RecoverPanic {
		defer func() {
			if r := recover(); r != nil {
				a.mu.Lock()
				entry.info.Status = scheduler.Failed
				a.mu.Unlock()
			}
		}()
	}

	a.mu.Lock()
	entry.info.Status = scheduler.Running
	entry.info.LastRun = time.Now()
	a.mu.Unlock()

	err := entry.job.Execute(ctx)

	a.mu.Lock()
	if err != nil {
		entry.info.Status = scheduler.Failed
	} else {
		entry.info.Status = scheduler.Completed
	}
	entry.info.RunCount++
	a.mu.Unlock()
}
