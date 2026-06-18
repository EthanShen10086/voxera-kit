package memory_test

import (
	"context"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/EthanShen10086/voxera-kit/scheduler"
	"github.com/EthanShen10086/voxera-kit/scheduler/memory"
)

type countingJob struct {
	name  string
	count *atomic.Int32
	err   error
}

func (j countingJob) Name() string { return j.name }

func (j countingJob) Execute(context.Context) error {
	j.count.Add(1)
	return j.err
}

func TestRegisterAndRunNow(t *testing.T) {
	s := memory.New(scheduler.Config{})
	var runs atomic.Int32
	job := countingJob{name: "tick", count: &runs}

	if err := s.Register("tick", "@every 1h", job); err != nil {
		t.Fatal(err)
	}
	if err := s.Register("tick", "@every 1h", job); err == nil {
		t.Fatal("expected duplicate register error")
	}
	if err := s.RunNow("tick"); err != nil {
		t.Fatal(err)
	}
	time.Sleep(50 * time.Millisecond)
	if runs.Load() < 1 {
		t.Fatal("job did not run")
	}
}

func TestStartStopAndList(t *testing.T) {
	ctx := context.Background()
	s := memory.New(scheduler.Config{MaxConcurrent: 2})
	var runs atomic.Int32
	if err := s.Register("job", "@every 20ms", countingJob{name: "job", count: &runs}); err != nil {
		t.Fatal(err)
	}
	if err := s.Start(ctx); err != nil {
		t.Fatal(err)
	}
	if err := s.Start(ctx); err == nil {
		t.Fatal("expected already running error")
	}
	if !s.IsRunning() {
		t.Fatal("expected running")
	}
	time.Sleep(60 * time.Millisecond)
	if err := s.Stop(ctx); err != nil {
		t.Fatal(err)
	}
	if s.IsRunning() {
		t.Fatal("expected stopped")
	}
	if err := s.Stop(ctx); err == nil {
		t.Fatal("expected not running error")
	}
	infos := s.List()
	if len(infos) != 1 || infos[0].Name != "job" {
		t.Fatalf("List = %+v", infos)
	}
}

func TestUnregisterAndMissingJob(t *testing.T) {
	s := memory.New(scheduler.Config{})
	if err := s.Unregister("missing"); err == nil {
		t.Fatal("expected not found")
	}
	if err := s.RunNow("missing"); err == nil {
		t.Fatal("expected not found")
	}
	if err := s.Register("x", "@every 1s", countingJob{name: "x", count: &atomic.Int32{}}); err != nil {
		t.Fatal(err)
	}
	if err := s.Unregister("x"); err != nil {
		t.Fatal(err)
	}
}

func TestRecoverPanic(t *testing.T) {
	s := memory.New(scheduler.Config{RecoverPanic: true})
	panicJob := panicJob{name: "panic"}
	if err := s.Register("panic", "@every 1h", panicJob); err != nil {
		t.Fatal(err)
	}
	if err := s.RunNow("panic"); err != nil {
		t.Fatal(err)
	}
	time.Sleep(50 * time.Millisecond)
	infos := s.List()
	if len(infos) != 1 || infos[0].Status != scheduler.Failed {
		t.Fatalf("status = %+v", infos[0])
	}
}

type panicJob struct{ name string }

func (p panicJob) Name() string { return p.name }

func (p panicJob) Execute(context.Context) error { panic("boom") }

func TestNextRunDurationFallback(t *testing.T) {
	s := memory.New(scheduler.Config{})
	ctx := context.Background()
	if err := s.Register("cron", "0 * * * *", countingJob{name: "cron", count: &atomic.Int32{}}); err != nil {
		t.Fatal(err)
	}
	if err := s.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = s.Stop(ctx) }()
	infos := s.List()
	if len(infos) != 1 || !strings.Contains(infos[0].CronExpr, "*") {
		t.Fatalf("List = %+v", infos)
	}
}
