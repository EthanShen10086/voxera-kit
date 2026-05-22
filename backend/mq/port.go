// Package mq defines the port interfaces for message queue operations.
// It abstracts message publishing and subscribing across different brokers
// (NATS, Kafka, RabbitMQ) using a unified interface.
package mq

import (
	"context"
	"time"
)

// Message represents a single message transmitted through the message queue.
type Message struct {
	// ID is the unique identifier of the message.
	ID string
	// Topic is the subject or channel the message belongs to.
	Topic string
	// Payload is the raw message body.
	Payload []byte
	// Headers contains optional key-value metadata for the message.
	Headers map[string]string
	// Timestamp is the time when the message was produced.
	Timestamp time.Time
}

// MessageHandler is a callback function type for processing received messages.
type MessageHandler func(ctx context.Context, msg *Message) error

// Publisher is the interface for publishing messages to a topic.
type Publisher interface {
	// Publish sends a message to the specified topic.
	Publish(ctx context.Context, topic string, msg *Message) error
	// Close releases all resources held by the publisher.
	Close() error
}

// Subscriber is the interface for consuming messages from topics.
type Subscriber interface {
	// Subscribe registers a handler for messages on the given topic.
	Subscribe(ctx context.Context, topic string, handler MessageHandler) error
	// Unsubscribe removes the subscription for the given topic.
	Unsubscribe(topic string) error
	// Ack acknowledges successful processing of a message.
	Ack(ctx context.Context, msg *Message) error
	// Close releases all resources held by the subscriber.
	Close() error
}

// Config holds the connection parameters for a message queue broker.
type Config struct {
	// Brokers is the list of broker addresses to connect to.
	Brokers []string
	// Username is the authentication username.
	Username string
	// Password is the authentication password.
	Password string
	// GroupID is the consumer group identifier for coordinated consumption.
	GroupID string
}
