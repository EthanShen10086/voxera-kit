// Package assert provides shared test helpers for kit contract and integration tests.
package assert

import (
	"errors"
	"testing"
)

// ErrorIs fails when err does not wrap target.
func ErrorIs(t *testing.T, err, target error) {
	t.Helper()
	if !errors.Is(err, target) {
		t.Fatalf("error = %v, want %v", err, target)
	}
}
