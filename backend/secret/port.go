// Package secret provides an abstraction for managing application secrets such
// as API keys, tokens, and credentials.
package secret

import (
	"context"
	"errors"
)

// ErrNotFound is returned when a requested secret does not exist.
var ErrNotFound = errors.New("secret not found")

// ErrAccessDenied is returned when the caller lacks permission to access a secret.
var ErrAccessDenied = errors.New("access denied")

// Manager defines the interface for secret storage operations.
type Manager interface {
	// Get retrieves the value of a secret by key.
	Get(ctx context.Context, key string) (string, error)
	// Set stores or updates a secret with the given key and value.
	Set(ctx context.Context, key string, value string) error
	// Delete removes a secret by key.
	Delete(ctx context.Context, key string) error
	// List returns all secret keys matching the given prefix.
	List(ctx context.Context, prefix string) ([]string, error)
}
