package sync_test

import (
	"context"
	"errors"
	"testing"
	"time"

	sfsync "github.com/EthanShen10086/voxera-kit/singleflight/sync"
)

func TestDoReturnsResult(t *testing.T) {
	a := sfsync.New()
	val, shared, err := a.Do(context.Background(), "k", func() (any, error) {
		return 42, nil
	})
	if err != nil || val != 42 || shared {
		t.Fatalf("Do() = %v shared=%v err=%v", val, shared, err)
	}
}

func TestDoPropagatesError(t *testing.T) {
	a := sfsync.New()
	_, _, err := a.Do(context.Background(), "k", func() (any, error) {
		return nil, errors.New("boom")
	})
	if err == nil || err.Error() != "boom" {
		t.Fatalf("Do() = %v", err)
	}
}

func TestDoContextCancel(t *testing.T) {
	a := sfsync.New()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, _, err := a.Do(ctx, "k", func() (any, error) {
		time.Sleep(50 * time.Millisecond)
		return nil, nil
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Do() = %v", err)
	}
}
