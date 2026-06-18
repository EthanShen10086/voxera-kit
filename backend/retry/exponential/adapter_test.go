package exponential_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/EthanShen10086/voxera-kit/retry"
	"github.com/EthanShen10086/voxera-kit/retry/exponential"
)

func TestExecute_SuccessFirstTry(t *testing.T) {
	a := exponential.New(retry.Policy{MaxAttempts: 3}, nil)
	var calls int
	err := a.Execute(context.Background(), func(context.Context) error {
		calls++
		return nil
	})
	if err != nil || calls != 1 {
		t.Fatalf("calls=%d err=%v", calls, err)
	}
}

func TestExecute_RetriesThenSucceeds(t *testing.T) {
	a := exponential.New(retry.Policy{
		MaxAttempts:    3,
		InitialBackoff: time.Millisecond,
		Multiplier:     1,
		MaxBackoff:     time.Millisecond,
	}, nil)
	var calls int32
	err := a.Execute(context.Background(), func(context.Context) error {
		if atomic.AddInt32(&calls, 1) < 3 {
			return errors.New("transient")
		}
		return nil
	})
	if err != nil || atomic.LoadInt32(&calls) != 3 {
		t.Fatalf("calls=%d err=%v", calls, err)
	}
}

func TestExecute_NonRetryableStops(t *testing.T) {
	permanent := errors.New("permanent")
	a := exponential.New(retry.Policy{MaxAttempts: 5, InitialBackoff: time.Millisecond}, func(err error) bool {
		return !errors.Is(err, permanent)
	})
	var calls int
	err := a.Execute(context.Background(), func(context.Context) error {
		calls++
		return permanent
	})
	if !errors.Is(err, permanent) || calls != 1 {
		t.Fatalf("calls=%d err=%v", calls, err)
	}
}

func TestExecute_ExhaustedAttempts(t *testing.T) {
	fail := errors.New("fail")
	a := exponential.New(retry.Policy{
		MaxAttempts:    2,
		InitialBackoff: time.Millisecond,
		Multiplier:     1,
		MaxBackoff:     time.Millisecond,
	}, nil)
	err := a.Execute(context.Background(), func(context.Context) error {
		return fail
	})
	if !errors.Is(err, fail) {
		t.Fatalf("err=%v", err)
	}
}

func TestExecute_ContextCanceled(t *testing.T) {
	a := exponential.New(retry.Policy{
		MaxAttempts:    5,
		InitialBackoff: time.Second,
	}, nil)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := a.Execute(ctx, func(context.Context) error {
		return errors.New("nope")
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err=%v", err)
	}
}

func TestExecute_WithJitterAndMaxBackoff(t *testing.T) {
	fail := errors.New("fail")
	a := exponential.New(retry.Policy{
		MaxAttempts:    3,
		InitialBackoff: 2 * time.Millisecond,
		Multiplier:     10,
		MaxBackoff:     3 * time.Millisecond,
		JitterFraction: 0.5,
	}, nil)
	err := a.Execute(context.Background(), func(context.Context) error {
		return fail
	})
	if !errors.Is(err, fail) {
		t.Fatalf("err=%v", err)
	}
}
