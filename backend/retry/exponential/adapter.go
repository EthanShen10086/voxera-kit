// Package exponential implements exponential back-off with jitter.
package exponential

import (
	"context"
	"crypto/rand"
	"math/big"
	"time"

	"github.com/EthanShen10086/voxera-kit/retry"
)

// Adapter implements [retry.Retrier] using exponential back-off with optional
// jitter sourced from crypto/rand.
type Adapter struct {
	policy      retry.Policy
	isRetryable retry.IsRetryable
}

// New returns an Adapter configured with the given policy. If isRetryable is
// nil every non-nil error is considered retryable.
func New(p retry.Policy, isRetryable retry.IsRetryable) *Adapter {
	return &Adapter{
		policy:      p,
		isRetryable: isRetryable,
	}
}

var _ retry.Retrier = (*Adapter)(nil)

// Execute runs fn according to the configured policy. It returns the last
// error when all attempts are exhausted or immediately when the context is
// canceled.
func (a *Adapter) Execute(ctx context.Context, fn func(ctx context.Context) error) error {
	var lastErr error
	backoff := a.policy.InitialBackoff

	for attempt := 0; attempt < a.policy.MaxAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return err
		}

		lastErr = fn(ctx)
		if lastErr == nil {
			return nil
		}

		if a.isRetryable != nil && !a.isRetryable(lastErr) {
			return lastErr
		}

		if attempt == a.policy.MaxAttempts-1 {
			break
		}

		wait := a.backoffWithJitter(backoff)

		t := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			t.Stop()
			return ctx.Err()
		case <-t.C:
		}

		backoff = time.Duration(float64(backoff) * a.policy.Multiplier)
		if backoff > a.policy.MaxBackoff {
			backoff = a.policy.MaxBackoff
		}
	}

	return lastErr
}

// backoffWithJitter applies jitter to the base duration using crypto/rand.
func (a *Adapter) backoffWithJitter(base time.Duration) time.Duration {
	if a.policy.JitterFraction <= 0 {
		return base
	}

	jitterRange := int64(float64(base) * a.policy.JitterFraction)
	if jitterRange <= 0 {
		return base
	}

	n, err := rand.Int(rand.Reader, big.NewInt(jitterRange))
	if err != nil {
		return base
	}

	return base - time.Duration(jitterRange/2) + time.Duration(n.Int64())
}
