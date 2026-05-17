// Package registry defines the port interface for service registry and configuration
// center operations. It abstracts away the underlying backend (etcd, Consul, ZooKeeper),
// allowing different implementations to be used interchangeably.
package registry

import (
	"context"
	"time"
)

// ServiceStatus represents the health state of a service instance.
type ServiceStatus int

const (
	// Up indicates the service instance is healthy and ready.
	Up ServiceStatus = iota
	// Down indicates the service instance is unreachable or unhealthy.
	Down
	// Starting indicates the service instance is initializing.
	Starting
	// Stopping indicates the service instance is shutting down.
	Stopping
)

// ServiceInstance represents a single running instance of a service.
type ServiceInstance struct {
	ID           string
	Name         string
	Host         string
	Port         int
	Metadata     map[string]string
	HealthCheck  string
	Status       ServiceStatus
	RegisteredAt time.Time
}

// ServiceRegistry is the interface for service registration and discovery.
// Implementations must be safe for concurrent use.
type ServiceRegistry interface {
	// Register adds a service instance to the registry.
	Register(ctx context.Context, instance *ServiceInstance) error
	// Deregister removes a service instance from the registry by its ID.
	Deregister(ctx context.Context, instanceID string) error
	// Discover returns all healthy instances of the named service.
	Discover(ctx context.Context, serviceName string) ([]*ServiceInstance, error)
	// Watch registers a callback that fires when the instance list for a service changes.
	Watch(ctx context.Context, serviceName string, callback func([]*ServiceInstance)) error
	// Heartbeat renews the registration TTL for the given instance.
	Heartbeat(ctx context.Context, instanceID string) error
}

// ConfigValue represents a versioned configuration entry.
type ConfigValue struct {
	Key       string
	Value     string
	Version   int64
	UpdatedAt time.Time
}

// ConfigCenter is the interface for distributed configuration management.
// Implementations must be safe for concurrent use.
type ConfigCenter interface {
	// Get retrieves a configuration value by key.
	Get(ctx context.Context, key string) (*ConfigValue, error)
	// Set creates or updates a configuration key-value pair.
	Set(ctx context.Context, key, value string) error
	// Delete removes a configuration entry by key.
	Delete(ctx context.Context, key string) error
	// Watch registers a callback that fires when the given key changes.
	Watch(ctx context.Context, key string, callback func(*ConfigValue)) error
	// List returns all configuration entries matching the given key prefix.
	List(ctx context.Context, prefix string) ([]*ConfigValue, error)
}

// RegistryConfig holds configuration parameters for a registry/config center backend.
type RegistryConfig struct {
	// Endpoints is the list of backend server addresses.
	Endpoints []string
	// Namespace is the logical grouping prefix for all entries.
	Namespace string
	// TTL is the default time-to-live for service registrations.
	TTL time.Duration
}
