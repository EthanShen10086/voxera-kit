// Package cron provides a production-ready implementation of the scheduler.Scheduler
// interface backed by github.com/robfig/cron/v3. It supports standard five-field cron
// expressions, optional seconds, and descriptor shortcuts (@every, @hourly, etc.).
package cron

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	robfigcron "github.com/robfig/cron/v3"

	"github.com/EthanShen10086/voxera-kit/scheduler"
)

// Parser is the cron expression parser configured with optional seconds and descriptors.
var Parser = robfigcron.NewParser(
	robfigcron.SecondOptional | robfigcron.Minute | robfigcron.Hour |
		robfigcron.Dom | robfigcron.Month | robfigcron.Dow | robfigcron.Descriptor,
)

type jobEntry struct {
	job      scheduler.Job
	cronExpr string
	entryID  robfigcron.EntryID
	info     scheduler.JobInfo
}

// Scheduler is a cron-based implementation of the scheduler.Scheduler interface.
// It delegates scheduling to robfig/cron/v3 and adds concurrency control via a semaphore.
type Scheduler struct {
	mu        sync.RWMutex
	cron      *robfigcron.Cron
	jobs      map[string]*jobEntry
	running   atomic.Bool
	config    scheduler.Config
	cancel    context.CancelFunc
	semaphore chan struct{}
}

// NewScheduler creates a new cron-based scheduler with the given configuration.
func NewScheduler(cfg scheduler.Config) *Scheduler {
	maxConcurrent := cfg.MaxConcurrent
	if maxConcurrent <= 0 {
		maxConcurrent = 10
	}

	opts := []robfigcron.Option{
		robfigcron.WithParser(Parser),
	}

	if cfg.Location != nil {
		opts = append(opts, robfigcron.WithLocation(cfg.Location))
	}

	if cfg.RecoverPanic {
		opts = append(opts, robfigcron.WithChain(robfigcron.Recover(robfigcron.DefaultLogger)))
	}

	return &Scheduler{
		cron:      robfigcron.New(opts...),
		jobs:      make(map[string]*jobEntry),
		config:    cfg,
		semaphore: make(chan struct{}, maxConcurrent),
	}
}

// Register adds a job with the given cron expression. It supports standard five-field
// cron syntax (e.g. "0 */5 * * *"), optional leading seconds field, and robfig
// descriptors (e.g. "@every 30s", "@hourly").
func (s *Scheduler) Register(name string, cronExpr string, job scheduler.Job) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.jobs[name]; exists {
		return fmt.Errorf("scheduler: job %q already registered", name)
	}

	if _, err := Parser.Parse(cronExpr); err != nil {
		return fmt.Errorf("scheduler: invalid cron expression %q: %w", cronExpr, err)
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

	entryID, err := s.cron.AddFunc(cronExpr, func() {
		s.executeJob(entry)
	})
	if err != nil {
		return fmt.Errorf("scheduler: failed to register job %q: %w", name, err)
	}

	entry.entryID = entryID
	s.jobs[name] = entry
	return nil
}

// Unregister removes a previously registered job by name.
func (s *Scheduler) Unregister(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, exists := s.jobs[name]
	if !exists {
		return fmt.Errorf("scheduler: job %q not found", name)
	}

	s.cron.Remove(entry.entryID)
	delete(s.jobs, name)
	return nil
}

// Start begins the cron scheduler's event loop. It stops automatically when ctx is canceled.
func (s *Scheduler) Start(ctx context.Context) error {
	if s.running.Load() {
		return fmt.Errorf("scheduler: already running")
	}

	ctx, cancel := context.WithCancel(ctx)
	s.cancel = cancel
	s.running.Store(true)

	s.cron.Start()

	go func() {
		<-ctx.Done()
		s.cron.Stop()
		s.running.Store(false)
	}()

	return nil
}

// Stop gracefully shuts down the scheduler, waiting for running jobs to complete.
func (s *Scheduler) Stop(_ context.Context) error {
	if !s.running.Load() {
		return fmt.Errorf("scheduler: not running")
	}

	if s.cancel != nil {
		s.cancel()
	}

	stopCtx := s.cron.Stop()
	<-stopCtx.Done()
	s.running.Store(false)
	return nil
}

// RunNow triggers immediate execution of the named job outside of its regular schedule.
func (s *Scheduler) RunNow(name string) error {
	s.mu.RLock()
	entry, exists := s.jobs[name]
	s.mu.RUnlock()

	if !exists {
		return fmt.Errorf("scheduler: job %q not found", name)
	}

	go s.executeJob(entry)
	return nil
}

// List returns information about all registered jobs, including next scheduled run times.
func (s *Scheduler) List() []scheduler.JobInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	infos := make([]scheduler.JobInfo, 0, len(s.jobs))
	for _, entry := range s.jobs {
		info := entry.info
		cronEntry := s.cron.Entry(entry.entryID)
		if cronEntry.Valid() {
			info.NextRun = cronEntry.Next
		}
		infos = append(infos, info)
	}
	return infos
}

// IsRunning reports whether the scheduler is currently active.
func (s *Scheduler) IsRunning() bool {
	return s.running.Load()
}

func (s *Scheduler) executeJob(entry *jobEntry) {
	s.semaphore <- struct{}{}
	defer func() { <-s.semaphore }()

	s.mu.Lock()
	entry.info.Status = scheduler.Running
	entry.info.LastRun = time.Now()
	s.mu.Unlock()

	err := entry.job.Execute(context.Background())

	s.mu.Lock()
	if err != nil {
		entry.info.Status = scheduler.Failed
	} else {
		entry.info.Status = scheduler.Completed
	}
	entry.info.RunCount++
	s.mu.Unlock()
}
