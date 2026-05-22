// Package env provides a secret Manager backed by environment variables with a
// configurable prefix.
package env

import (
	"context"
	"os"
	"strings"

	"github.com/EthanShen10086/voxera-kit/secret"
)

// EnvManager reads and writes secrets as environment variables using a
// configurable prefix (e.g., "APP_SECRET").
type EnvManager struct {
	prefix string
}

// NewEnvManager creates a new EnvManager that maps secret keys to environment
// variables formatted as PREFIX_KEY.
func NewEnvManager(prefix string) *EnvManager {
	return &EnvManager{prefix: prefix}
}

// Get retrieves a secret value from the corresponding environment variable.
// Returns ErrNotFound if the variable is not set.
func (m *EnvManager) Get(_ context.Context, key string) (string, error) {
	envKey := m.envKey(key)
	val, ok := os.LookupEnv(envKey)
	if !ok {
		return "", secret.ErrNotFound
	}
	return val, nil
}

// Set stores a secret by setting the corresponding environment variable.
func (m *EnvManager) Set(_ context.Context, key string, value string) error {
	return os.Setenv(m.envKey(key), value)
}

// Delete removes a secret by unsetting the corresponding environment variable.
func (m *EnvManager) Delete(_ context.Context, key string) error {
	return os.Unsetenv(m.envKey(key))
}

// List returns all secret keys (without the prefix) that match the given prefix
// filter within the environment.
func (m *EnvManager) List(_ context.Context, prefix string) ([]string, error) {
	fullPrefix := m.envKey(prefix)
	var keys []string
	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) < 2 {
			continue
		}
		name := parts[0]
		if strings.HasPrefix(name, fullPrefix) {
			key := strings.TrimPrefix(name, m.prefix+"_")
			keys = append(keys, key)
		}
	}
	return keys, nil
}

func (m *EnvManager) envKey(key string) string {
	return m.prefix + "_" + key
}
