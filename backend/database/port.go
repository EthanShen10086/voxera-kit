// Package database defines the port interfaces for database operations.
// It follows the hexagonal architecture pattern, allowing different database
// implementations (PostgreSQL, MongoDB, MySQL) to be swapped transparently.
package database

import (
	"context"
	"time"
)

// DatabaseConfig holds the connection parameters for a relational database.
type DatabaseConfig struct {
	// Host is the database server hostname or IP address.
	Host string
	// Port is the database server port number.
	Port int
	// User is the database authentication username.
	User string
	// Password is the database authentication password.
	Password string
	// Database is the name of the target database.
	Database string
	// SSLMode controls the SSL/TLS negotiation mode (e.g., "disable", "require").
	SSLMode string
	// MaxOpenConns sets the maximum number of open connections in the pool.
	MaxOpenConns int
	// MaxIdleConns sets the maximum number of idle connections in the pool.
	MaxIdleConns int
	// ConnMaxLifetime sets the maximum duration a connection may be reused.
	ConnMaxLifetime time.Duration
}

// Repository is a generic interface for CRUD operations on entities of type T.
// Implementations should handle mapping between the domain entity and the
// underlying storage format.
type Repository[T any] interface {
	// Create persists a new entity to the database.
	Create(ctx context.Context, entity *T) error
	// FindByID retrieves a single entity by its unique identifier.
	FindByID(ctx context.Context, id string) (*T, error)
	// FindAll retrieves all entities matching the given filter criteria.
	FindAll(ctx context.Context, filter map[string]any) ([]*T, error)
	// Update modifies an existing entity identified by id.
	Update(ctx context.Context, id string, entity *T) error
	// Delete removes an entity by its unique identifier.
	Delete(ctx context.Context, id string) error
	// Count returns the number of entities matching the given filter.
	Count(ctx context.Context, filter map[string]any) (int64, error)
}

// Transaction represents an active database transaction.
// Callers must either Commit or Rollback to release resources.
type Transaction interface {
	// Begin starts a new nested transaction or savepoint.
	Begin(ctx context.Context) (Transaction, error)
	// Commit finalizes the transaction, persisting all changes.
	Commit() error
	// Rollback aborts the transaction, discarding all changes.
	Rollback() error
}

// Database is the top-level interface for managing a database connection lifecycle.
type Database interface {
	// Connect establishes a connection to the database using the provided config.
	Connect(ctx context.Context, cfg DatabaseConfig) error
	// Close terminates the database connection and releases resources.
	Close() error
	// Ping verifies that the database connection is still alive.
	Ping(ctx context.Context) error
	// Transaction returns a new Transaction for atomic multi-statement operations.
	Transaction() Transaction
}
