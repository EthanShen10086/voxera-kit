// Package nats provides a NATS implementation of the mq.Publisher and mq.Subscriber interfaces.
// It is intended to use github.com/nats-io/nats.go as the underlying client.
package nats

import (
	"context"

	"github.com/EthanShen10086/voxera-kit/mq"
)

// Publisher implements the mq.Publisher interface using NATS.
//
// Intended dependency: github.com/nats-io/nats.go
type Publisher struct {
	// conn *nats.Conn // TODO: uncomment when nats.go dependency is added
	cfg mq.Config
}

// NewPublisher creates a new NATS Publisher with the provided configuration.
func NewPublisher(cfg mq.Config) *Publisher {
	return &Publisher{cfg: cfg}
}

// Publish sends a message to the specified NATS subject.
func (p *Publisher) Publish(ctx context.Context, topic string, msg *mq.Message) error {
	// TODO: implement using nats.go
	return nil
}

// Close disconnects the NATS publisher connection.
func (p *Publisher) Close() error {
	// TODO: implement using nats.go
	return nil
}

// Subscriber implements the mq.Subscriber interface using NATS.
//
// Intended dependency: github.com/nats-io/nats.go
type Subscriber struct {
	// conn *nats.Conn // TODO: uncomment when nats.go dependency is added
	cfg mq.Config
}

// NewSubscriber creates a new NATS Subscriber with the provided configuration.
func NewSubscriber(cfg mq.Config) *Subscriber {
	return &Subscriber{cfg: cfg}
}

// Subscribe registers a handler for messages on the given NATS subject.
func (s *Subscriber) Subscribe(ctx context.Context, topic string, handler mq.MessageHandler) error {
	// TODO: implement using nats.go
	return nil
}

// Unsubscribe removes the subscription for the given NATS subject.
func (s *Subscriber) Unsubscribe(topic string) error {
	// TODO: implement using nats.go
	return nil
}

// Ack acknowledges a message. NATS core does not require explicit acks;
// this is relevant for JetStream mode.
func (s *Subscriber) Ack(ctx context.Context, msg *mq.Message) error {
	// TODO: implement using nats.go JetStream
	return nil
}

// Close disconnects the NATS subscriber connection.
func (s *Subscriber) Close() error {
	// TODO: implement using nats.go
	return nil
}
