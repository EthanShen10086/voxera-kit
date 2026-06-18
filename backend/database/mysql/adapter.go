// Package mysql provides a MySQL implementation of the database.Database interface.
// It uses database/sql with github.com/go-sql-driver/mysql.
package mysql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/EthanShen10086/voxera-kit/database"
	_ "github.com/go-sql-driver/mysql"
)

// Adapter implements the database.Database interface using MySQL.
type Adapter struct {
	db *sql.DB
}

// New creates a new MySQL Adapter instance.
func New() *Adapter {
	return &Adapter{}
}

// Connect establishes a connection pool to the MySQL database.
func (a *Adapter) Connect(ctx context.Context, cfg database.Config) error {
	db, err := sql.Open("mysql", dsn(cfg))
	if err != nil {
		return fmt.Errorf("mysql: open: %w", err)
	}

	if cfg.MaxOpenConns > 0 {
		db.SetMaxOpenConns(cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns > 0 {
		db.SetMaxIdleConns(cfg.MaxIdleConns)
	}
	if cfg.ConnMaxLifetime > 0 {
		db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	}

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return fmt.Errorf("mysql: ping: %w", err)
	}

	if a.db != nil {
		_ = a.db.Close()
	}
	a.db = db
	return nil
}

// Close shuts down the MySQL connection pool.
func (a *Adapter) Close() error {
	if a.db == nil {
		return nil
	}
	err := a.db.Close()
	a.db = nil
	return err
}

// Ping verifies the MySQL connection is alive.
func (a *Adapter) Ping(ctx context.Context) error {
	if a.db == nil {
		return fmt.Errorf("mysql: not connected")
	}
	return a.db.PingContext(ctx)
}

// Transaction returns a new Transaction backed by a sql.Tx.
func (a *Adapter) Transaction() database.Transaction {
	return &Transaction{db: a.db}
}

func dsn(cfg database.Config) string {
	port := cfg.Port
	if port == 0 {
		port = 3306
	}

	tls := "false"
	switch cfg.SSLMode {
	case "require", "verify-ca", "verify-full", "true", "preferred":
		tls = "true"
	}

	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&tls=%s",
		cfg.User, cfg.Password, cfg.Host, port, cfg.Database, tls)
}
