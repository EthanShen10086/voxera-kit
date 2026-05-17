// Package mysql provides a MySQL implementation of the database.Database interface.
// It uses database/sql with a MySQL driver such as github.com/go-sql-driver/mysql.
package mysql

import (
	"context"

	"github.com/EthanShen10086/voxera-kit/database"
)

// Adapter implements the database.Database interface using MySQL.
//
// Intended dependency: database/sql with github.com/go-sql-driver/mysql
type Adapter struct {
	// db *sql.DB // TODO: uncomment when MySQL driver dependency is added
}

// New creates a new MySQL Adapter instance.
func New() *Adapter {
	return &Adapter{}
}

// Connect establishes a connection pool to the MySQL database.
func (a *Adapter) Connect(ctx context.Context, cfg database.DatabaseConfig) error {
	// TODO: implement using database/sql with MySQL driver
	return nil
}

// Close shuts down the MySQL connection pool.
func (a *Adapter) Close() error {
	// TODO: implement using database/sql with MySQL driver
	return nil
}

// Ping verifies the MySQL connection is alive.
func (a *Adapter) Ping(ctx context.Context) error {
	// TODO: implement using database/sql with MySQL driver
	return nil
}

// Transaction returns a new Transaction backed by a sql.Tx.
func (a *Adapter) Transaction() database.Transaction {
	// TODO: implement using database/sql with MySQL driver
	return nil
}
