// Package scheduler defines the port interface for cron/scheduled task operations.
// It abstracts away the underlying scheduler implementation, allowing different
// backends to be used interchangeably.
package scheduler

import (
	"context"
	"time"
)

// JobStatus represents the current state of a scheduled job.
type JobStatus int

const (
	// Pending indicates the job is registered but has not yet run.
	Pending JobStatus = iota
	// Running indicates the job is currently executing.
	Running
	// Completed indicates the job finished successfully.
	Completed
	// Failed indicates the job finished with an error.
	Failed
	// Cancelled indicates the job was cancelled before completion.
	Cancelled
)

// JobInfo holds runtime metadata about a scheduled job.
type JobInfo struct {
	ID       string
	Name     string
	CronExpr string
	Status   JobStatus
	LastRun  time.Time
	NextRun  time.Time
	RunCount int64
}

// Job is the interface that scheduled tasks must implement.
type Job interface {
	// Execute runs the job's logic within the given context.
	Execute(ctx context.Context) error
	// Name returns a human-readable name for the job.
	Name() string
}

// Scheduler is the interface for managing cron/scheduled jobs.
// Implementations must be safe for concurrent use.
type Scheduler interface {
	// Register adds a job with the given cron expression.
	Register(name string, cronExpr string, job Job) error
	// Unregister removes a previously registered job by name.
	Unregister(name string) error
	// Start begins the scheduler's event loop.
	Start(ctx context.Context) error
	// Stop gracefully shuts down the scheduler.
	Stop(ctx context.Context) error
	// RunNow triggers immediate execution of the named job.
	RunNow(name string) error
	// List returns information about all registered jobs.
	List() []JobInfo
	// IsRunning reports whether the scheduler is currently active.
	IsRunning() bool
}

// SchedulerConfig holds configuration parameters for a scheduler backend.
type SchedulerConfig struct {
	// MaxConcurrent is the maximum number of jobs that can run simultaneously.
	MaxConcurrent int
	// Location is the time zone used for cron expression evaluation.
	Location *time.Location
	// RecoverPanic controls whether the scheduler recovers from panicking jobs.
	RecoverPanic bool
	// Logger is an optional logger instance (typed as any for flexibility).
	Logger any
}
