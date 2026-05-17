// Package circuitbreaker defines the port interface for the circuit breaker pattern.
// It prevents cascading failures by short-circuiting calls to a failing dependency
// and allowing it time to recover.
package circuitbreaker

import (
	"context"
	"errors"
	"time"
)

// State represents the current state of a circuit breaker.
type State int

const (
	// Closed allows all calls through and monitors for failures.
	Closed State = iota
	// Open rejects all calls immediately without executing them.
	Open
	// HalfOpen allows a limited number of probe calls to test recovery.
	HalfOpen
)

// ErrCircuitOpen is returned when a call is rejected because the circuit is open.
var ErrCircuitOpen = errors.New("circuit breaker is open")

// ErrTooManyCalls is returned when a call is rejected because the half-open
// circuit has reached its maximum number of probe calls.
var ErrTooManyCalls = errors.New("too many calls in half-open state")

// CircuitBreakerConfig holds the parameters for constructing a circuit breaker.
type CircuitBreakerConfig struct {
	// MaxFailures is the number of consecutive failures before the circuit opens.
	MaxFailures int
	// Timeout is how long the circuit stays open before transitioning to half-open.
	Timeout time.Duration
	// HalfOpenMaxCalls is the maximum number of probe calls allowed in half-open state.
	HalfOpenMaxCalls int
	// OnStateChange is called whenever the circuit transitions between states.
	OnStateChange func(from, to State)
}

// CircuitBreaker wraps calls to an external dependency and manages failure detection
// and recovery. Implementations must be safe for concurrent use.
type CircuitBreaker interface {
	// Execute runs fn if the circuit permits it, recording the outcome.
	Execute(ctx context.Context, fn func() error) error
	// State returns the current state of the circuit.
	State() State
	// Reset forces the circuit back to the closed state.
	Reset()
	// Counts returns the running success and failure tallies.
	Counts() (successes, failures int)
}
