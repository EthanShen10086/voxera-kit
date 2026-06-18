package containers

import (
	"context"
	"fmt"
	"strings"

	tcrabbit "github.com/testcontainers/testcontainers-go/modules/rabbitmq"
)

// RabbitMQ holds a running RabbitMQ testcontainer AMQP URL.
type RabbitMQ struct {
	URL       string
	terminate func(context.Context) error
}

// StartRabbitMQ launches rabbitmq:3.12 for integration tests.
func StartRabbitMQ(ctx context.Context) (*RabbitMQ, error) {
	c, err := tcrabbit.Run(ctx, "rabbitmq:3.12.11-management-alpine")
	if err != nil {
		return nil, fmt.Errorf("containers: start rabbitmq: %w", err)
	}
	url, err := c.AmqpURL(ctx)
	if err != nil {
		_ = c.Terminate(ctx)
		return nil, fmt.Errorf("containers: rabbitmq amqp url: %w", err)
	}
	url = strings.Replace(url, "localhost:", "127.0.0.1:", 1)
	return &RabbitMQ{
		URL:       url,
		terminate: func(ctx context.Context) error { return c.Terminate(ctx) },
	}, nil
}

// Terminate stops the container.
func (r *RabbitMQ) Terminate(ctx context.Context) error {
	if r == nil || r.terminate == nil {
		return nil
	}
	return r.terminate(ctx)
}
