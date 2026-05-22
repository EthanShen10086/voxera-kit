// Package adaptive implements an AIMD (Additive Increase Multiplicative
// Decrease) load shedder that adjusts its acceptance threshold based on the
// success/failure ratio within a sliding window.
package adaptive

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/EthanShen10086/voxera-kit/loadshed"
)

// Adapter implements [loadshed.Shedder] using an AIMD algorithm.
type Adapter struct {
	cfg       loadshed.Config
	mu        sync.Mutex
	successes atomic.Int64
	failures  atomic.Int64
	windowEnd time.Time
}

// New creates a load shedder with the given configuration.
func New(cfg loadshed.Config) *Adapter {
	return &Adapter{
		cfg:       cfg,
		windowEnd: time.Now().Add(cfg.Window),
	}
}

var _ loadshed.Shedder = (*Adapter)(nil)

// Allow checks the current load and returns a token if the request is
// admitted. Returns [loadshed.ErrOverloaded] when the load exceeds the
// configured threshold.
func (a *Adapter) Allow() (loadshed.Token, error) {
	if a.Load() >= a.cfg.MaxLoad {
		return nil, loadshed.ErrOverloaded
	}
	return &token{adapter: a}, nil
}

// Load returns the failure ratio in the current sliding window as a value
// between 0 and 1. A fresh window returns 0.
func (a *Adapter) Load() float64 {
	a.maybeResetWindow()

	s := a.successes.Load()
	f := a.failures.Load()
	total := s + f
	if total == 0 {
		return 0
	}

	return float64(f) / float64(total)
}

// maybeResetWindow advances the sliding window when the current one has
// expired.
func (a *Adapter) maybeResetWindow() {
	now := time.Now()
	a.mu.Lock()
	defer a.mu.Unlock()

	if now.After(a.windowEnd) {
		a.successes.Store(0)
		a.failures.Store(0)
		a.windowEnd = now.Add(a.cfg.Window)
	}
}

// recordSuccess increments the success counter (additive increase).
func (a *Adapter) recordSuccess() {
	a.successes.Add(1)
}

// recordFailure increments the failure counter (multiplicative decrease
// effect on the load ratio).
func (a *Adapter) recordFailure() {
	a.failures.Add(1)
}

// token implements [loadshed.Token].
type token struct {
	adapter *Adapter
	done    atomic.Bool
}

// Done reports the outcome of the request. It is safe to call multiple
// times; only the first call takes effect.
func (t *token) Done(success bool) {
	if !t.done.CompareAndSwap(false, true) {
		return
	}
	if success {
		t.adapter.recordSuccess()
	} else {
		t.adapter.recordFailure()
	}
}
