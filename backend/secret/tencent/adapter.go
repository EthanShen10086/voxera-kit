// Package tencent provides a placeholder secret Manager for Tencent Cloud SSM.
// Full SDK integration is not yet implemented.
package tencent

import (
	"context"
	"fmt"

	"github.com/EthanShen10086/voxera-kit/secret"
)

// Manager is a stub Tencent Cloud secret Manager adapter.
type Manager struct{}

// NewManager creates a stub Tencent Cloud secret Manager.
func NewManager() *Manager {
	return &Manager{}
}

// Get returns ErrNotFound because Tencent Cloud integration is not implemented.
func (m *Manager) Get(_ context.Context, key string) (string, error) {
	return "", fmt.Errorf("tencent cloud secrets manager is not implemented for key %q: %w", key, secret.ErrNotFound)
}

// Set returns an error because Tencent Cloud integration is not implemented.
func (m *Manager) Set(_ context.Context, key string, _ string) error {
	return fmt.Errorf("tencent cloud secrets manager is not implemented for key %q", key)
}

// Delete returns an error because Tencent Cloud integration is not implemented.
func (m *Manager) Delete(_ context.Context, key string) error {
	return fmt.Errorf("tencent cloud secrets manager is not implemented for key %q", key)
}

// List returns an error because Tencent Cloud integration is not implemented.
func (m *Manager) List(_ context.Context, prefix string) ([]string, error) {
	return nil, fmt.Errorf("tencent cloud secrets manager is not implemented for prefix %q", prefix)
}
