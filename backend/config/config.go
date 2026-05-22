// Package config defines a provider-agnostic configuration interface that
// supports multiple backends (file, env, remote) and dynamic watching.
package config

import "time"

// Source identifies where configuration values originate.
type Source string

const (
	// SourceFile represents a local file source (YAML, TOML, JSON, etc.).
	SourceFile Source = "file"
	// SourceEnv represents environment variables.
	SourceEnv Source = "env"
	// SourceRemote represents a remote config center (Consul, etcd, Apollo, Nacos, etc.).
	SourceRemote Source = "remote"
)

// Options controls how the config provider is initialized.
type Options struct {
	Source   Source
	Path     string
	Endpoint string
	Format   string
	Watch    bool
}

// Config is the top-level interface every configuration provider must implement.
type Config interface {
	// Get returns the raw value for a dot-separated key path.
	Get(key string) any

	// GetString returns the string value for a key or an empty string.
	GetString(key string) string

	// GetInt returns the integer value for a key or 0.
	GetInt(key string) int

	// GetBool returns the boolean value for a key or false.
	GetBool(key string) bool

	// GetDuration returns the time.Duration value for a key or 0.
	GetDuration(key string) time.Duration

	// Sub returns a Config scoped to a key prefix, or nil when the prefix does
	// not exist.
	Sub(key string) Config

	// Unmarshal decodes the configuration tree (or a sub-tree) into dst,
	// which must be a pointer to a struct.
	Unmarshal(key string, dst any) error

	// Watch starts observing configuration changes. Implementations that do not
	// support watching may return nil immediately.
	Watch() error

	// OnChange registers a callback that fires whenever a watched key changes.
	OnChange(key string, fn func(key string, value any))
}
