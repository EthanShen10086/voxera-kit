// Package loadshed provides adaptive load shedding to protect services from
// overload by rejecting excess traffic.
package loadshed

import (
	"errors"
	"time"
)

// ErrOverloaded is returned when the current load exceeds the configured
// threshold.
var ErrOverloaded = errors.New("loadshed: overloaded")

// Token represents an in-flight request. Callers must invoke Done when the
// request completes to update the load signal.
type Token interface {
	// Done reports the outcome of the request. Pass true when the request
	// succeeded and false on failure.
	Done(success bool)
}

// Shedder decides whether a new request should be admitted based on the
// current load level.
type Shedder interface {
	// Allow returns a Token if the request is admitted. If the system is
	// overloaded it returns ErrOverloaded.
	Allow() (Token, error)
	// Load returns the current load as a value between 0 and 1.
	Load() float64
}

// Config controls the AIMD load shedder behavior.
type Config struct {
	// MaxLoad is the threshold above which requests are rejected (0–1).
	MaxLoad float64
	// Window is the duration of the sliding observation window.
	Window time.Duration
	// CooldownStep is the multiplicative decrease factor applied on
	// failure (e.g. 0.5 halves the threshold).
	CooldownStep float64
	// IncreaseStep is the additive increase applied per successful
	// request.
	IncreaseStep float64
}
