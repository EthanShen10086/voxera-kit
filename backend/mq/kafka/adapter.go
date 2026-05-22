// Package kafka provides a Kafka implementation of the mq.Publisher and mq.Subscriber interfaces.
// It is intended to use github.com/segmentio/kafka-go as the underlying client.
package kafka

import (
	"context"

	"github.com/EthanShen10086/voxera-kit/mq"
)

// Publisher implements the mq.Publisher interface using Apache Kafka.
//
// Intended dependency: github.com/segmentio/kafka-go
type Publisher struct {
	// writer *kafka.Writer // TODO: uncomment when kafka-go dependency is added
	cfg mq.Config
}

// NewPublisher creates a new Kafka Publisher with the provided configuration.
func NewPublisher(cfg mq.Config) *Publisher {
	return &Publisher{cfg: cfg}
}

// Publish sends a message to the specified Kafka topic.
func (p *Publisher) Publish(ctx context.Context, topic string, msg *mq.Message) error {
	// TODO: implement using kafka-go
	return nil
}

// Close shuts down the Kafka writer.
func (p *Publisher) Close() error {
	// TODO: implement using kafka-go
	return nil
}

// Subscriber implements the mq.Subscriber interface using Apache Kafka.
//
// Intended dependency: github.com/segmentio/kafka-go
type Subscriber struct {
	// reader *kafka.Reader // TODO: uncomment when kafka-go dependency is added
	cfg mq.Config
}

// NewSubscriber creates a new Kafka Subscriber with the provided configuration.
func NewSubscriber(cfg mq.Config) *Subscriber {
	return &Subscriber{cfg: cfg}
}

// Subscribe registers a handler for messages on the given Kafka topic.
func (s *Subscriber) Subscribe(ctx context.Context, topic string, handler mq.MessageHandler) error {
	// TODO: implement using kafka-go
	return nil
}

// Unsubscribe stops consuming from the given Kafka topic.
func (s *Subscriber) Unsubscribe(topic string) error {
	// TODO: implement using kafka-go
	return nil
}

// Ack commits the offset for the processed message.
func (s *Subscriber) Ack(ctx context.Context, msg *mq.Message) error {
	// TODO: implement using kafka-go
	return nil
}

// Close shuts down the Kafka reader.
func (s *Subscriber) Close() error {
	// TODO: implement using kafka-go
	return nil
}
