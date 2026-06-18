// Package fixture provides lightweight test data builders shared across kit modules and products.
package fixture

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync/atomic"
)

var seq atomic.Uint64

// NewID returns a random 32-character hex identifier suitable for tests.
func NewID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		panic("fixture: crypto/rand failed: " + err.Error())
	}
	return hex.EncodeToString(b)
}

// NewPrefixedID returns prefix joined with a random hex identifier.
func NewPrefixedID(prefix string) string {
	return fmt.Sprintf("%s-%s", prefix, NewID())
}

// NewSequentialID returns a deterministic identifier for stable assertions.
func NewSequentialID(prefix string) string {
	n := seq.Add(1)
	return fmt.Sprintf("%s-%d", prefix, n)
}

// ResetSequentialIDs resets the sequential counter used by NewSequentialID.
func ResetSequentialIDs() {
	seq.Store(0)
}
