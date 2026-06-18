package semaphore_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/EthanShen10086/voxera-kit/bulkhead"
	"github.com/EthanShen10086/voxera-kit/bulkhead/semaphore"
)

func TestExecuteSuccess(t *testing.T) {
	a := semaphore.New(bulkhead.Config{Name: "api", MaxConcurrent: 2, MaxWaitTime: time.Second})
	err := a.Execute(context.Background(), func() error { return nil })
	if err != nil {
		t.Fatal(err)
	}
	if a.Name() != "api" {
		t.Fatalf("Name = %q", a.Name())
	}
	if a.Available() != 2 {
		t.Fatalf("Available = %d", a.Available())
	}
}

func TestExecuteFull(t *testing.T) {
	a := semaphore.New(bulkhead.Config{MaxConcurrent: 1, MaxWaitTime: 20 * time.Millisecond})
	block := make(chan struct{})
	release := make(chan struct{})
	go func() {
		_ = a.Execute(context.Background(), func() error {
			close(block)
			<-release
			return nil
		})
	}()
	<-block
	err := a.Execute(context.Background(), func() error { return nil })
	close(release)
	if !errors.Is(err, bulkhead.ErrBulkheadFull) {
		t.Fatalf("Execute() = %v", err)
	}
}

func TestExecuteContextCancel(t *testing.T) {
	a := semaphore.New(bulkhead.Config{MaxConcurrent: 1, MaxWaitTime: time.Second})
	block := make(chan struct{})
	release := make(chan struct{})
	go func() {
		_ = a.Execute(context.Background(), func() error {
			close(block)
			<-release
			return nil
		})
	}()
	<-block
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := a.Execute(ctx, func() error { return nil })
	close(release)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Execute() = %v", err)
	}
}

func TestActiveCountAndMaxWaitTime(t *testing.T) {
	wait := 50 * time.Millisecond
	a := semaphore.New(bulkhead.Config{MaxConcurrent: 2, MaxWaitTime: wait})
	if a.MaxWaitTime() != wait {
		t.Fatalf("MaxWaitTime = %v", a.MaxWaitTime())
	}
	if a.ActiveCount() != 0 {
		t.Fatalf("ActiveCount = %d", a.ActiveCount())
	}
}
