package cron_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/EthanShen10086/voxera-kit/scheduler"
	"github.com/EthanShen10086/voxera-kit/scheduler/cron"
)

type countingJob struct {
	name  string
	count *atomic.Int32
}

func (j countingJob) Name() string { return j.name }

func (j countingJob) Execute(context.Context) error {
	j.count.Add(1)
	return nil
}

func TestRegisterStartStop(t *testing.T) {
	s := cron.NewScheduler(scheduler.Config{MaxConcurrent: 2})
	var runs atomic.Int32
	job := countingJob{name: "tick", count: &runs}

	if err := s.Register("tick", "@every 1h", job); err != nil {
		t.Fatal(err)
	}
	if err := s.Register("tick", "@every 1h", job); err == nil {
		t.Fatal("expected duplicate")
	}
	if err := s.Register("bad", "not a cron", job); err == nil {
		t.Fatal("expected invalid cron")
	}

	ctx := context.Background()
	if err := s.Start(ctx); err != nil {
		t.Fatal(err)
	}
	if err := s.Start(ctx); err == nil {
		t.Fatal("expected already running")
	}
	if !s.IsRunning() {
		t.Fatal("expected running")
	}
	if err := s.RunNow("tick"); err != nil {
		t.Fatal(err)
	}
	time.Sleep(50 * time.Millisecond)
	if err := s.Stop(ctx); err != nil {
		t.Fatal(err)
	}
	if runs.Load() < 1 {
		t.Fatal("job did not run")
	}
}

func TestRunNowAndList(t *testing.T) {
	s := cron.NewScheduler(scheduler.Config{})
	var runs atomic.Int32
	if err := s.Register("now", "0 0 * * *", countingJob{name: "now", count: &runs}); err != nil {
		t.Fatal(err)
	}
	if err := s.RunNow("now"); err != nil {
		t.Fatal(err)
	}
	time.Sleep(50 * time.Millisecond)
	if runs.Load() < 1 {
		t.Fatal("RunNow did not execute")
	}
	infos := s.List()
	if len(infos) != 1 || infos[0].Name != "now" {
		t.Fatalf("List = %+v", infos)
	}
	if err := s.Unregister("missing"); err == nil {
		t.Fatal("expected not found")
	}
}
