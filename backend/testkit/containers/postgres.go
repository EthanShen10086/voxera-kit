package containers

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/EthanShen10086/voxera-kit/database"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

// Postgres holds a running PostgreSQL testcontainer.
type Postgres struct {
	Config    database.Config
	DSN       string
	terminate func(context.Context) error
}

// StartPostgres launches postgres:16-alpine with a test database.
func StartPostgres(ctx context.Context) (*Postgres, error) {
	c, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("voxera_test"),
		postgres.WithUsername("voxera"),
		postgres.WithPassword("voxera"),
	)
	if err != nil {
		return nil, fmt.Errorf("containers: start postgres: %w", err)
	}
	dsn, err := c.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		_ = c.Terminate(ctx)
		return nil, fmt.Errorf("containers: postgres dsn: %w", err)
	}
	cfg, err := parsePostgresDSN(dsn)
	if err != nil {
		_ = c.Terminate(ctx)
		return nil, err
	}
	return &Postgres{
		Config:    cfg,
		DSN:       dsn,
		terminate: func(ctx context.Context) error { return c.Terminate(ctx) },
	}, nil
}

// Terminate stops the container.
func (p *Postgres) Terminate(ctx context.Context) error {
	if p == nil || p.terminate == nil {
		return nil
	}
	return p.terminate(ctx)
}

func parsePostgresDSN(dsn string) (database.Config, error) {
	u, err := url.Parse(dsn)
	if err != nil {
		return database.Config{}, fmt.Errorf("containers: parse postgres dsn: %w", err)
	}
	user := ""
	pass := ""
	if u.User != nil {
		user = u.User.Username()
		pass, _ = u.User.Password()
	}
	host := u.Hostname()
	port := 5432
	if p := u.Port(); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil {
			port = parsed
		}
	}
	dbName := strings.TrimPrefix(u.Path, "/")
	sslmode := u.Query().Get("sslmode")
	return database.Config{
		Host:     host,
		Port:     port,
		User:     user,
		Password: pass,
		Database: dbName,
		SSLMode:  sslmode,
	}, nil
}
