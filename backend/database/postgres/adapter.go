// Package postgres provides a PostgreSQL implementation of the database.Database interface.
// It is intended to use github.com/jackc/pgx/v5 as the underlying driver.
package postgres

import (
	"context"

	"github.com/EthanShen10086/voxera-kit/database"
)

// Adapter implements the database.Database interface using PostgreSQL via pgx.
//
// Intended dependency: github.com/jackc/pgx/v5
type Adapter struct {
	// pool *pgxpool.Pool // TODO: uncomment when pgx dependency is added
}

// New creates a new PostgreSQL Adapter instance.
func New() *Adapter {
	return &Adapter{}
}

// Connect establishes a connection pool to the PostgreSQL database.
func (a *Adapter) Connect(ctx context.Context, cfg database.DatabaseConfig) error {
	// TODO: implement using pgx
	return nil
}

// Close shuts down the connection pool gracefully.
func (a *Adapter) Close() error {
	// TODO: implement using pgx
	return nil
}

// Ping verifies the database connection is alive.
func (a *Adapter) Ping(ctx context.Context) error {
	// TODO: implement using pgx
	return nil
}

// Transaction returns a new Transaction backed by a pgx transaction.
func (a *Adapter) Transaction() database.Transaction {
	// TODO: implement using pgx
	return nil
}
