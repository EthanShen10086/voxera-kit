package containers

import (
	"context"
	"fmt"
	"strings"

	"github.com/testcontainers/testcontainers-go/modules/nats"
)

// NATS holds a running NATS testcontainer connection URL.
type NATS struct {
	URL       string
	terminate func(context.Context) error
}

// StartNATS launches nats:2-alpine and returns the client URL.
func StartNATS(ctx context.Context) (*NATS, error) {
	c, err := nats.Run(ctx, "nats:2.10-alpine")
	if err != nil {
		return nil, fmt.Errorf("containers: start nats: %w", err)
	}
	url, err := c.ConnectionString(ctx)
	if err != nil {
		_ = c.Terminate(ctx)
		return nil, fmt.Errorf("containers: nats connection string: %w", err)
	}
	url = strings.Replace(url, "nats://localhost:", "nats://127.0.0.1:", 1)
	return &NATS{
		URL:       url,
		terminate: func(ctx context.Context) error { return c.Terminate(ctx) },
	}, nil
}

// Terminate stops the container.
func (n *NATS) Terminate(ctx context.Context) error {
	if n == nil || n.terminate == nil {
		return nil
	}
	return n.terminate(ctx)
}
