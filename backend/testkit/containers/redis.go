// Package containers provides testcontainers helpers for integration tests.
package containers

import (
	"context"
	"fmt"

	"github.com/testcontainers/testcontainers-go/modules/redis"
)

// Redis holds a running Redis testcontainer endpoint.
type Redis struct {
	Address   string
	terminate func(context.Context) error
}

// StartRedis launches redis:7-alpine and returns host:port.
func StartRedis(ctx context.Context) (*Redis, error) {
	c, err := redis.Run(ctx, "redis:7-alpine")
	if err != nil {
		return nil, fmt.Errorf("containers: start redis: %w", err)
	}
	addr, err := c.Endpoint(ctx, "")
	if err != nil {
		_ = c.Terminate(ctx)
		return nil, fmt.Errorf("containers: redis endpoint: %w", err)
	}
	return &Redis{
		Address: addr,
		terminate: func(ctx context.Context) error { return c.Terminate(ctx) },
	}, nil
}

// Terminate stops the container.
func (r *Redis) Terminate(ctx context.Context) error {
	if r == nil || r.terminate == nil {
		return nil
	}
	return r.terminate(ctx)
}
