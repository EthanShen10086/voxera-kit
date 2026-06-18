// Package kafka provides a Kafka implementation of the mq.Publisher and mq.Subscriber interfaces.
// It uses github.com/segmentio/kafka-go as the underlying client.
package kafka

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/EthanShen10086/voxera-kit/mq"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/scram"
)

func dialer(cfg mq.Config) *kafka.Dialer {
	d := &kafka.Dialer{Timeout: 10 * time.Second}
	if cfg.Username != "" {
		mechanism, err := scram.Mechanism(scram.SHA256, cfg.Username, cfg.Password)
		if err == nil {
			d.SASLMechanism = mechanism
		}
	}
	return d
}

// Publisher implements the mq.Publisher interface using Apache Kafka.
type Publisher struct {
	writer *kafka.Writer
}

// NewPublisher creates a new Kafka Publisher with the provided configuration.
func NewPublisher(cfg mq.Config) (*Publisher, error) {
	if len(cfg.Brokers) == 0 {
		return nil, fmt.Errorf("kafka: at least one broker address is required")
	}
	return &Publisher{
		writer: &kafka.Writer{
			Addr:     kafka.TCP(cfg.Brokers...),
			Balancer: &kafka.LeastBytes{},
			Transport: &kafka.Transport{
				Dial: dialer(cfg).DialFunc,
			},
		},
	}, nil
}

// Publish sends a message to the specified Kafka topic.
func (p *Publisher) Publish(ctx context.Context, topic string, msg *mq.Message) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if msg == nil {
		return fmt.Errorf("kafka: message is nil")
	}

	headers := make([]kafka.Header, 0, len(msg.Headers)+2)
	for k, v := range msg.Headers {
		headers = append(headers, kafka.Header{Key: k, Value: []byte(v)})
	}
	if msg.ID != "" {
		headers = append(headers, kafka.Header{Key: "id", Value: []byte(msg.ID)})
	}
	if !msg.Timestamp.IsZero() {
		headers = append(headers, kafka.Header{
			Key:   "timestamp",
			Value: []byte(msg.Timestamp.Format(time.RFC3339Nano)),
		})
	}

	kmsg := kafka.Message{
		Topic:   topic,
		Key:     []byte(msg.ID),
		Value:   msg.Payload,
		Headers: headers,
		Time:    msg.Timestamp,
	}
	return p.writer.WriteMessages(ctx, kmsg)
}

// Close shuts down the Kafka writer.
func (p *Publisher) Close() error {
	if p.writer == nil {
		return nil
	}
	return p.writer.Close()
}

type kafkaSubscription struct {
	reader *kafka.Reader
	cancel context.CancelFunc
}

// Subscriber implements the mq.Subscriber interface using Apache Kafka.
type Subscriber struct {
	cfg     mq.Config
	mu      sync.Mutex
	readers map[string]*kafkaSubscription
	pending sync.Map // msg.ID -> kafka.Message
}

// NewSubscriber creates a new Kafka Subscriber with the provided configuration.
func NewSubscriber(cfg mq.Config) (*Subscriber, error) {
	if len(cfg.Brokers) == 0 {
		return nil, fmt.Errorf("kafka: at least one broker address is required")
	}
	return &Subscriber{
		cfg:     cfg,
		readers: make(map[string]*kafkaSubscription),
	}, nil
}

// Subscribe registers a handler for messages on the given Kafka topic.
func (s *Subscriber) Subscribe(ctx context.Context, topic string, handler mq.MessageHandler) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if handler == nil {
		return fmt.Errorf("kafka: handler is nil")
	}

	s.mu.Lock()
	if _, exists := s.readers[topic]; exists {
		s.mu.Unlock()
		return fmt.Errorf("kafka: already subscribed to topic %q", topic)
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  s.cfg.Brokers,
		Topic:    topic,
		GroupID:  fmt.Sprintf("%s-%s", s.cfg.GroupID, topic),
		Dialer:   dialer(s.cfg),
		MinBytes: 1,
		MaxBytes: 10e6,
	})
	subCtx, cancel := context.WithCancel(ctx)
	s.readers[topic] = &kafkaSubscription{reader: reader, cancel: cancel}
	s.mu.Unlock()

	go s.consume(subCtx, topic, reader, handler)
	return nil
}

func (s *Subscriber) consume(ctx context.Context, topic string, reader *kafka.Reader, handler mq.MessageHandler) {
	for {
		kmsg, err := reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			continue
		}

		msg := &mq.Message{
			Topic:     topic,
			Payload:   append([]byte(nil), kmsg.Value...),
			Headers:   make(map[string]string, len(kmsg.Headers)),
			Timestamp: kmsg.Time,
		}
		for _, h := range kmsg.Headers {
			switch h.Key {
			case "id":
				msg.ID = string(h.Value)
			case "timestamp":
				if parsed, err := time.Parse(time.RFC3339Nano, string(h.Value)); err == nil {
					msg.Timestamp = parsed
				}
			default:
				msg.Headers[h.Key] = string(h.Value)
			}
		}
		if msg.ID == "" {
			msg.ID = fmt.Sprintf("kafka-%d-%d", kmsg.Partition, kmsg.Offset)
		}
		if msg.Timestamp.IsZero() {
			msg.Timestamp = time.Now()
		}

		s.pending.Store(msg.ID, kmsg)
		if err := handler(ctx, msg); err == nil {
			s.pending.Delete(msg.ID)
		}
	}
}

// Unsubscribe stops consuming from the given Kafka topic.
func (s *Subscriber) Unsubscribe(topic string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	sub, ok := s.readers[topic]
	if !ok {
		return nil
	}
	if sub.cancel != nil {
		sub.cancel()
	}
	err := sub.reader.Close()
	delete(s.readers, topic)
	return err
}

// Ack commits the offset for the processed message.
func (s *Subscriber) Ack(ctx context.Context, msg *mq.Message) error {
	if msg == nil || msg.ID == "" {
		return fmt.Errorf("kafka: message id is required for ack")
	}
	raw, ok := s.pending.Load(msg.ID)
	if !ok {
		return nil
	}
	kmsg, ok := raw.(kafka.Message)
	if !ok {
		return fmt.Errorf("kafka: invalid pending message for id %q", msg.ID)
	}

	s.mu.Lock()
	sub, ok := s.readers[msg.Topic]
	s.mu.Unlock()
	if !ok {
		return fmt.Errorf("kafka: no active reader for topic %q", msg.Topic)
	}
	return sub.reader.CommitMessages(ctx, kmsg)
}

// Close shuts down the Kafka reader.
func (s *Subscriber) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var firstErr error
	for topic, sub := range s.readers {
		if sub.cancel != nil {
			sub.cancel()
		}
		if err := sub.reader.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
		delete(s.readers, topic)
	}
	return firstErr
}
