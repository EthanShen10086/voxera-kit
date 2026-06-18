package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/EthanShen10086/voxera-kit/task"
)

// Worker polls a Redis task queue and executes handlers with retry and DLQ.
type Worker struct {
	adapter *Adapter
	handler task.Handler
	poll    time.Duration
}

// WorkerConfig configures the Redis task worker.
type WorkerConfig struct {
	Adapter *Adapter
	Handler task.Handler
	// PollInterval defaults to 50ms.
	PollInterval time.Duration
}

// NewWorker creates a worker for the given Redis adapter.
func NewWorker(cfg WorkerConfig) (*Worker, error) {
	if cfg.Adapter == nil {
		return nil, fmt.Errorf("redis worker: adapter is nil")
	}
	if cfg.Handler == nil {
		return nil, fmt.Errorf("redis worker: handler is nil")
	}
	poll := cfg.PollInterval
	if poll == 0 {
		poll = 50 * time.Millisecond
	}
	return &Worker{
		adapter: cfg.Adapter,
		handler: cfg.Handler,
		poll:    poll,
	}, nil
}

// Run processes due tasks until ctx is canceled.
func (w *Worker) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		tk, ok, err := w.adapter.PopDue(ctx)
		if err != nil {
			return err
		}
		if !ok {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(w.poll):
				continue
			}
		}

		if err := w.handleOne(ctx, tk); err != nil {
			return err
		}
	}
}

func (w *Worker) handleOne(ctx context.Context, tk task.Task) error {
	attempt := tk.Attempt
	if attempt == 0 {
		attempt = 1
	}
	tk.Attempt = attempt

	if err := w.handler(ctx, tk); err != nil {
		retry := tk.Retry
		if requeueErr := w.adapter.RequeueOrDLQ(ctx, tk, err, retry); requeueErr != nil {
			return requeueErr
		}
		return nil
	}
	return w.adapter.Complete(ctx, tk)
}
