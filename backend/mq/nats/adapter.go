// Package nats provides a NATS implementation of the mq.Publisher and mq.Subscriber interfaces.
// It uses github.com/nats-io/nats.go as the underlying client.
package nats

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/EthanShen10086/voxera-kit/mq"
	"github.com/nats-io/nats.go"
)

func connect(cfg mq.Config) (*nats.Conn, error) {
	if len(cfg.Brokers) == 0 {
		return nil, fmt.Errorf("nats: at least one broker address is required")
	}
	opts := []nats.Option{}
	if cfg.Username != "" {
		opts = append(opts, nats.UserInfo(cfg.Username, cfg.Password))
	}
	return nats.Connect(cfg.Brokers[0], opts...)
}

// Publisher implements the mq.Publisher interface using NATS.
type Publisher struct {
	conn *nats.Conn
}

// NewPublisher creates a new NATS Publisher with the provided configuration.
func NewPublisher(cfg mq.Config) (*Publisher, error) {
	conn, err := connect(cfg)
	if err != nil {
		return nil, err
	}
	return &Publisher{conn: conn}, nil
}

// Publish sends a message to the specified NATS subject.
func (p *Publisher) Publish(ctx context.Context, topic string, msg *mq.Message) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if msg == nil {
		return fmt.Errorf("nats: message is nil")
	}

	data := msg.Payload
	hdr := nats.Header{}
	for k, v := range msg.Headers {
		hdr.Set(k, v)
	}
	if msg.ID != "" {
		hdr.Set("id", msg.ID)
	}
	if !msg.Timestamp.IsZero() {
		hdr.Set("timestamp", msg.Timestamp.Format(time.RFC3339Nano))
	}

	m := &nats.Msg{Subject: topic, Data: data, Header: hdr}
	return p.conn.PublishMsg(m)
}

// Close disconnects the NATS publisher connection.
func (p *Publisher) Close() error {
	if p.conn == nil {
		return nil
	}
	p.conn.Close()
	return nil
}

type natsSubscription struct {
	sub *nats.Subscription
}

// Subscriber implements the mq.Subscriber interface using NATS.
type Subscriber struct {
	conn *nats.Conn
	mu   sync.Mutex
	subs map[string]*natsSubscription
	cfg  mq.Config
}

// NewSubscriber creates a new NATS Subscriber with the provided configuration.
func NewSubscriber(cfg mq.Config) (*Subscriber, error) {
	conn, err := connect(cfg)
	if err != nil {
		return nil, err
	}
	return &Subscriber{
		conn: conn,
		subs: make(map[string]*natsSubscription),
		cfg:  cfg,
	}, nil
}

// Subscribe registers a handler for messages on the given NATS subject.
func (s *Subscriber) Subscribe(ctx context.Context, topic string, handler mq.MessageHandler) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if handler == nil {
		return fmt.Errorf("nats: handler is nil")
	}

	sub, err := s.conn.Subscribe(topic, func(m *nats.Msg) {
		msg := &mq.Message{
			Topic:     topic,
			Payload:   append([]byte(nil), m.Data...),
			Headers:   make(map[string]string, len(m.Header)),
			Timestamp: time.Now(),
		}
		for k, vals := range m.Header {
			if len(vals) > 0 {
				msg.Headers[k] = vals[0]
			}
		}
		if id := msg.Headers["id"]; id != "" {
			msg.ID = id
			delete(msg.Headers, "id")
		}
		if ts := msg.Headers["timestamp"]; ts != "" {
			if parsed, err := time.Parse(time.RFC3339Nano, ts); err == nil {
				msg.Timestamp = parsed
				delete(msg.Headers, "timestamp")
			}
		}
		_ = handler(ctx, msg)
	})
	if err != nil {
		return err
	}

	s.mu.Lock()
	s.subs[topic] = &natsSubscription{sub: sub}
	s.mu.Unlock()
	return nil
}

// Unsubscribe removes the subscription for the given NATS subject.
func (s *Subscriber) Unsubscribe(topic string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, ok := s.subs[topic]
	if !ok {
		return nil
	}
	err := entry.sub.Unsubscribe()
	delete(s.subs, topic)
	return err
}

// Ack acknowledges a message. NATS core does not require explicit acks.
func (s *Subscriber) Ack(_ context.Context, _ *mq.Message) error {
	return nil
}

// Close disconnects the NATS subscriber connection.
func (s *Subscriber) Close() error {
	s.mu.Lock()
	for topic, entry := range s.subs {
		_ = entry.sub.Unsubscribe()
		delete(s.subs, topic)
	}
	s.mu.Unlock()

	if s.conn == nil {
		return nil
	}
	s.conn.Close()
	return nil
}
