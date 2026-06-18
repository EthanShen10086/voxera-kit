// Package postgres provides a PostgreSQL implementation of the database.Database interface.
// It uses github.com/jackc/pgx/v5 as the underlying driver.
package postgres

import (
	"context"
	"fmt"
	"net/url"

	"github.com/EthanShen10086/voxera-kit/database"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Adapter implements the database.Database interface using PostgreSQL via pgx.
type Adapter struct {
	pool *pgxpool.Pool
}

// New creates a new PostgreSQL Adapter instance.
func New() *Adapter {
	return &Adapter{}
}

// Connect establishes a connection pool to the PostgreSQL database.
func (a *Adapter) Connect(ctx context.Context, cfg database.Config) error {
	poolConfig, err := pgxpool.ParseConfig(dsn(cfg))
	if err != nil {
		return fmt.Errorf("postgres: parse config: %w", err)
	}

	if cfg.MaxOpenConns > 0 {
		poolConfig.MaxConns = int32(cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns > 0 {
		poolConfig.MinConns = int32(cfg.MaxIdleConns)
	}
	if cfg.ConnMaxLifetime > 0 {
		poolConfig.MaxConnLifetime = cfg.ConnMaxLifetime
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return fmt.Errorf("postgres: connect: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return fmt.Errorf("postgres: ping: %w", err)
	}

	if a.pool != nil {
		a.pool.Close()
	}
	a.pool = pool
	return nil
}

// Close shuts down the connection pool gracefully.
func (a *Adapter) Close() error {
	if a.pool == nil {
		return nil
	}
	a.pool.Close()
	a.pool = nil
	return nil
}

// Ping verifies the database connection is alive.
func (a *Adapter) Ping(ctx context.Context) error {
	if a.pool == nil {
		return fmt.Errorf("postgres: not connected")
	}
	return a.pool.Ping(ctx)
}

// Transaction returns a new Transaction backed by a pgx transaction.
func (a *Adapter) Transaction() database.Transaction {
	return &Transaction{pool: a.pool}
}

func dsn(cfg database.Config) string {
	sslmode := cfg.SSLMode
	if sslmode == "" {
		sslmode = "disable"
	}
	port := cfg.Port
	if port == 0 {
		port = 5432
	}

	u := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(cfg.User, cfg.Password),
		Host:   fmt.Sprintf("%s:%d", cfg.Host, port),
		Path:   cfg.Database,
	}
	q := u.Query()
	q.Set("sslmode", sslmode)
	u.RawQuery = q.Encode()
	return u.String()
}
