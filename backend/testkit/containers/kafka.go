package containers

import (
	"context"
	"fmt"
	"strings"

	tckafka "github.com/testcontainers/testcontainers-go/modules/kafka"
)

// Kafka holds a running Kafka testcontainer.
type Kafka struct {
	Brokers   []string
	terminate func(context.Context) error
}

// StartKafka launches a KRaft Kafka broker for integration tests.
func StartKafka(ctx context.Context) (*Kafka, error) {
	c, err := tckafka.Run(ctx, "confluentinc/confluent-local:7.5.0")
	if err != nil {
		return nil, fmt.Errorf("containers: start kafka: %w", err)
	}
	brokers, err := c.Brokers(ctx)
	if err != nil {
		_ = c.Terminate(ctx)
		return nil, fmt.Errorf("containers: kafka brokers: %w", err)
	}
	for i, b := range brokers {
		brokers[i] = strings.Replace(b, "localhost:", "127.0.0.1:", 1)
	}
	return &Kafka{
		Brokers:   brokers,
		terminate: func(ctx context.Context) error { return c.Terminate(ctx) },
	}, nil
}

// Terminate stops the container.
func (k *Kafka) Terminate(ctx context.Context) error {
	if k == nil || k.terminate == nil {
		return nil
	}
	return k.terminate(ctx)
}
