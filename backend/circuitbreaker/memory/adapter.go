// Package memory provides an in-process implementation of the circuitbreaker.CircuitBreaker interface
// using failure counters and timed state transitions.
package memory

import (
	"context"
	"sync"
	"time"

	"github.com/EthanShen10086/voxera-kit/circuitbreaker"
)

// Adapter implements circuitbreaker.CircuitBreaker with counter-based state management.
type Adapter struct {
	mu            sync.Mutex
	cfg           circuitbreaker.Config
	state         circuitbreaker.State
	successes     int
	failures      int
	halfOpenCalls int
	lastFailure   time.Time
}

// New creates a new in-memory circuit breaker from the given configuration.
func New(cfg circuitbreaker.Config) *Adapter {
	return &Adapter{
		cfg:   cfg,
		state: circuitbreaker.Closed,
	}
}

// Execute runs fn if the circuit permits it, recording the outcome.
func (a *Adapter) Execute(_ context.Context, fn func() error) error {
	a.mu.Lock()
	if a.state == circuitbreaker.Open {
		if time.Since(a.lastFailure) > a.cfg.Timeout {
			a.setState(circuitbreaker.HalfOpen)
			a.halfOpenCalls = 0
		} else {
			a.mu.Unlock()
			return circuitbreaker.ErrCircuitOpen
		}
	}
	if a.state == circuitbreaker.HalfOpen && a.halfOpenCalls >= a.cfg.HalfOpenMaxCalls {
		a.mu.Unlock()
		return circuitbreaker.ErrTooManyCalls
	}
	if a.state == circuitbreaker.HalfOpen {
		a.halfOpenCalls++
	}
	a.mu.Unlock()

	err := fn()

	a.mu.Lock()
	defer a.mu.Unlock()
	if err != nil {
		a.failures++
		a.lastFailure = time.Now()
		if a.state == circuitbreaker.HalfOpen || a.failures >= a.cfg.MaxFailures {
			a.setState(circuitbreaker.Open)
		}
	} else {
		a.successes++
		if a.state == circuitbreaker.HalfOpen {
			a.setState(circuitbreaker.Closed)
			a.failures = 0
			a.halfOpenCalls = 0
		}
	}
	return err
}

// State returns the current state of the circuit.
func (a *Adapter) State() circuitbreaker.State {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.state == circuitbreaker.Open && time.Since(a.lastFailure) > a.cfg.Timeout {
		a.setState(circuitbreaker.HalfOpen)
		a.halfOpenCalls = 0
	}
	return a.state
}

// Reset forces the circuit back to the closed state and clears all counters.
func (a *Adapter) Reset() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.setState(circuitbreaker.Closed)
	a.successes = 0
	a.failures = 0
	a.halfOpenCalls = 0
}

// Counts returns the running success and failure tallies.
func (a *Adapter) Counts() (successes, failures int) {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.successes, a.failures
}

func (a *Adapter) setState(to circuitbreaker.State) {
	from := a.state
	a.state = to
	if a.cfg.OnStateChange != nil && from != to {
		a.cfg.OnStateChange(from, to)
	}
}
