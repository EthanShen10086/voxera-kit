// Package vault provides a secret Manager backed by HashiCorp Vault KV v2.
package vault

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/EthanShen10086/voxera-kit/secret"
	vaultapi "github.com/hashicorp/vault/api"
)

// Config holds Vault connection settings.
type Config struct {
	// Address is the Vault server URL (e.g. https://127.0.0.1:8200).
	Address string
	// Token is the Vault authentication token.
	Token string
	// Mount is the KV v2 mount path. Defaults to "secret".
	Mount string
}

// Manager reads and writes secrets from HashiCorp Vault KV v2.
type Manager struct {
	client *vaultapi.Client
	mount  string
}

// NewManager creates a Vault-backed secret Manager.
func NewManager(cfg Config) (*Manager, error) {
	config := vaultapi.DefaultConfig()
	if cfg.Address != "" {
		config.Address = cfg.Address
	}

	client, err := vaultapi.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("vault: create client: %w", err)
	}
	if cfg.Token != "" {
		client.SetToken(cfg.Token)
	}

	mount := cfg.Mount
	if mount == "" {
		mount = "secret"
	}

	return &Manager{client: client, mount: mount}, nil
}

// Get retrieves a secret value from Vault KV v2.
func (m *Manager) Get(ctx context.Context, key string) (string, error) {
	s, err := m.client.KVv2(m.mount).Get(ctx, key)
	if err != nil {
		if isNotFound(err) {
			return "", secret.ErrNotFound
		}
		return "", err
	}
	if s == nil || s.Data == nil {
		return "", secret.ErrNotFound
	}

	if val, ok := s.Data["value"].(string); ok {
		return val, nil
	}
	for _, v := range s.Data {
		if str, ok := v.(string); ok {
			return str, nil
		}
	}
	return "", secret.ErrNotFound
}

// Set stores or updates a secret in Vault KV v2.
func (m *Manager) Set(ctx context.Context, key string, value string) error {
	_, err := m.client.KVv2(m.mount).Put(ctx, key, map[string]any{
		"value": value,
	})
	return err
}

// Delete removes secret metadata and all versions from Vault KV v2.
func (m *Manager) Delete(ctx context.Context, key string) error {
	return m.client.KVv2(m.mount).DeleteMetadata(ctx, key)
}

// List returns secret keys under the given prefix from Vault KV v2 metadata.
func (m *Manager) List(ctx context.Context, prefix string) ([]string, error) {
	path := m.mount + "/metadata"
	if prefix != "" {
		path += "/" + strings.Trim(prefix, "/")
	}

	s, err := m.client.Logical().ListWithContext(ctx, path)
	if err != nil {
		if isNotFound(err) {
			return nil, secret.ErrNotFound
		}
		return nil, err
	}
	if s == nil || s.Data == nil {
		return nil, nil
	}

	keysRaw, ok := s.Data["keys"].([]any)
	if !ok {
		return nil, nil
	}

	keys := make([]string, 0, len(keysRaw))
	for _, k := range keysRaw {
		if str, ok := k.(string); ok {
			keys = append(keys, str)
		}
	}
	return keys, nil
}

func isNotFound(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, vaultapi.ErrSecretNotFound) {
		return true
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "404") || strings.Contains(msg, "not found")
}
