// Package database defines the port interfaces for database operations.
// It follows the hexagonal architecture pattern, allowing different database
// implementations (PostgreSQL, MongoDB, MySQL) to be swapped transparently.
package database

import (
	"context"
	"time"
)

// Config holds the connection parameters for a relational database.
type Config struct {
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
	Connect(ctx context.Context, cfg Config) error
	// Close terminates the database connection and releases resources.
	Close() error
	// Ping verifies that the database connection is still alive.
	Ping(ctx context.Context) error
	// Transaction returns a new Transaction for atomic multi-statement operations.
	Transaction() Transaction
}

// QueryCondition represents a single filter predicate for query building.
type QueryCondition struct {
	Field    string
	Operator string
	Value    any
}

// SortOrder defines the direction of an ORDER BY clause.
type SortOrder int

const (
	// Asc sorts in ascending order.
	Asc SortOrder = iota
	// Desc sorts in descending order.
	Desc
)

// OrderBy specifies a single sort field and its direction.
type OrderBy struct {
	Field string
	Order SortOrder
}

// Pagination holds paging metadata for paginated query results.
type Pagination struct {
	Page  int
	Size  int
	Total int64
}

// QueryBuilder provides a fluent API for constructing type-safe queries.
type QueryBuilder[T any] interface {
	Where(field string, op string, value any) QueryBuilder[T]
	And(field string, op string, value any) QueryBuilder[T]
	Or(field string, op string, value any) QueryBuilder[T]
	OrderBy(field string, order SortOrder) QueryBuilder[T]
	Limit(n int) QueryBuilder[T]
	Offset(n int) QueryBuilder[T]
	Select(fields ...string) QueryBuilder[T]
	Execute(ctx context.Context) ([]T, error)
	Count(ctx context.Context) (int64, error)
	First(ctx context.Context) (*T, error)
	Paginate(ctx context.Context, page, size int) ([]T, *Pagination, error)
}

// Migration represents a single versioned schema migration.
type Migration interface {
	Up(ctx context.Context) error
	Down(ctx context.Context) error
	Version() string
	Name() string
}

// MigrationStatus records the state of an applied migration.
type MigrationStatus struct {
	Version   string
	Name      string
	AppliedAt time.Time
	Status    string
}

// Migrator manages the execution and tracking of schema migrations.
type Migrator interface {
	Apply(ctx context.Context, migrations ...Migration) error
	Rollback(ctx context.Context, steps int) error
	Status(ctx context.Context) ([]MigrationStatus, error)
}

// DBCluster abstracts a master-slave database cluster topology.
type DBCluster interface {
	Master() Database
	Slave() Database
	Close() error
}

// DBClusterConfig holds connection parameters for a master-slave cluster.
type DBClusterConfig struct {
	MasterDSN       string
	SlaveDSNs       []string
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifetime time.Duration
}
