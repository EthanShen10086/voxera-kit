// Package rabbitmq provides a RabbitMQ implementation of the mq.Publisher and mq.Subscriber interfaces.
// It is intended to use github.com/rabbitmq/amqp091-go as the underlying client.
package rabbitmq

import (
	"context"

	"github.com/EthanShen10086/voxera-kit/mq"
)

// Publisher implements the mq.Publisher interface using RabbitMQ.
//
// Intended dependency: github.com/rabbitmq/amqp091-go
type Publisher struct {
	// conn *amqp091.Connection // TODO: uncomment when amqp091-go dependency is added
	// ch   *amqp091.Channel
	cfg mq.MQConfig
}

// NewPublisher creates a new RabbitMQ Publisher with the provided configuration.
func NewPublisher(cfg mq.MQConfig) *Publisher {
	return &Publisher{cfg: cfg}
}

// Publish sends a message to the specified RabbitMQ exchange/routing key.
func (p *Publisher) Publish(ctx context.Context, topic string, msg *mq.Message) error {
	// TODO: implement using amqp091-go
	return nil
}

// Close shuts down the RabbitMQ publisher channel and connection.
func (p *Publisher) Close() error {
	// TODO: implement using amqp091-go
	return nil
}

// Subscriber implements the mq.Subscriber interface using RabbitMQ.
//
// Intended dependency: github.com/rabbitmq/amqp091-go
type Subscriber struct {
	// conn *amqp091.Connection // TODO: uncomment when amqp091-go dependency is added
	// ch   *amqp091.Channel
	cfg mq.MQConfig
}

// NewSubscriber creates a new RabbitMQ Subscriber with the provided configuration.
func NewSubscriber(cfg mq.MQConfig) *Subscriber {
	return &Subscriber{cfg: cfg}
}

// Subscribe registers a handler for messages on the given RabbitMQ queue.
func (s *Subscriber) Subscribe(ctx context.Context, topic string, handler mq.MessageHandler) error {
	// TODO: implement using amqp091-go
	return nil
}

// Unsubscribe cancels the consumer for the given queue.
func (s *Subscriber) Unsubscribe(topic string) error {
	// TODO: implement using amqp091-go
	return nil
}

// Ack acknowledges successful processing of a message.
func (s *Subscriber) Ack(ctx context.Context, msg *mq.Message) error {
	// TODO: implement using amqp091-go
	return nil
}

// Close shuts down the RabbitMQ subscriber channel and connection.
func (s *Subscriber) Close() error {
	// TODO: implement using amqp091-go
	return nil
}
