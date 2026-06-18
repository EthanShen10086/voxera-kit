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

const defaultStream = "VOXERA"

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

func streamName(cfg mq.Config) string {
	if cfg.Stream != "" {
		return cfg.Stream
	}
	return defaultStream
}

func ensureStream(js nats.JetStreamContext, cfg mq.Config, subjects ...string) error {
	name := streamName(cfg)
	_, err := js.StreamInfo(name)
	if err == nil {
		return nil
	}
	_, err = js.AddStream(&nats.StreamConfig{
		Name:     name,
		Subjects: subjects,
		Storage:  nats.FileStorage,
	})
	return err
}

// Publisher implements the mq.Publisher interface using NATS.
type Publisher struct {
	conn *nats.Conn
	js   nats.JetStreamContext
	cfg  mq.Config
}

// NewPublisher creates a new NATS Publisher with the provided configuration.
func NewPublisher(cfg mq.Config) (*Publisher, error) {
	conn, err := connect(cfg)
	if err != nil {
		return nil, err
	}
	p := &Publisher{conn: conn, cfg: cfg}
	if cfg.JetStream {
		js, err := conn.JetStream()
		if err != nil {
			conn.Close()
			return nil, fmt.Errorf("nats: jetstream: %w", err)
		}
		p.js = js
	}
	return p, nil
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

	if p.js != nil {
		if err := ensureStream(p.js, p.cfg, topic, topic+".>"); err != nil {
			return fmt.Errorf("nats: ensure stream: %w", err)
		}
		nmsg := &nats.Msg{Subject: topic, Data: data, Header: hdr}
		opts := []nats.PubOpt{}
		if msg.ID != "" {
			opts = append(opts, nats.MsgId(msg.ID))
		}
		_, err := p.js.PublishMsg(nmsg, opts...)
		return err
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
	conn       *nats.Conn
	js         nats.JetStreamContext
	mu         sync.Mutex
	subs       map[string]*natsSubscription
	pendingAck map[string]*nats.Msg
	cfg        mq.Config
}

// NewSubscriber creates a new NATS Subscriber with the provided configuration.
func NewSubscriber(cfg mq.Config) (*Subscriber, error) {
	conn, err := connect(cfg)
	if err != nil {
		return nil, err
	}
	s := &Subscriber{
		conn:       conn,
		subs:       make(map[string]*natsSubscription),
		pendingAck: make(map[string]*nats.Msg),
		cfg:        cfg,
	}
	if cfg.JetStream {
		js, err := conn.JetStream()
		if err != nil {
			conn.Close()
			return nil, fmt.Errorf("nats: jetstream: %w", err)
		}
		s.js = js
	}
	return s, nil
}

// Subscribe registers a handler for messages on the given NATS subject.
func (s *Subscriber) Subscribe(ctx context.Context, topic string, handler mq.MessageHandler) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if handler == nil {
		return fmt.Errorf("nats: handler is nil")
	}

	if s.js != nil {
		if err := ensureStream(s.js, s.cfg, topic, topic+".>"); err != nil {
			return fmt.Errorf("nats: ensure stream: %w", err)
		}
		durable := s.cfg.Durable
		if durable == "" {
			durable = s.cfg.GroupID
		}
		if durable == "" {
			durable = "voxera-" + topic
		}

		sub, err := s.js.Subscribe(topic, func(m *nats.Msg) {
			msg := decodeMsg(topic, m)
			s.mu.Lock()
			s.pendingAck[msg.ID] = m
			s.mu.Unlock()

			if err := handler(ctx, msg); err != nil {
				_ = m.Nak()
				s.mu.Lock()
				delete(s.pendingAck, msg.ID)
				s.mu.Unlock()
				return
			}
		}, nats.BindStream(streamName(s.cfg)), nats.Durable(durable), nats.ManualAck())
		if err != nil {
			return err
		}
		s.mu.Lock()
		s.subs[topic] = &natsSubscription{sub: sub}
		s.mu.Unlock()
		if err := s.conn.Flush(); err != nil {
			return fmt.Errorf("nats: flush after jetstream subscribe: %w", err)
		}
		return nil
	}

	sub, err := s.conn.Subscribe(topic, func(m *nats.Msg) {
		msg := decodeMsg(topic, m)
		_ = handler(ctx, msg)
	})
	if err != nil {
		return err
	}

	s.mu.Lock()
	s.subs[topic] = &natsSubscription{sub: sub}
	s.mu.Unlock()

	if err := s.conn.Flush(); err != nil {
		return fmt.Errorf("nats: flush after subscribe: %w", err)
	}
	return nil
}

func decodeMsg(topic string, m *nats.Msg) *mq.Message {
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
	} else if m.Header != nil {
		if ids := m.Header.Get("Nats-Msg-Id"); ids != "" {
			msg.ID = ids
		}
	}
	if msg.ID == "" {
		meta, _ := m.Metadata()
		if meta != nil {
			msg.ID = fmt.Sprintf("%s-%d", meta.Stream, meta.Sequence)
		}
	}
	if ts := msg.Headers["timestamp"]; ts != "" {
		if parsed, err := time.Parse(time.RFC3339Nano, ts); err == nil {
			msg.Timestamp = parsed
			delete(msg.Headers, "timestamp")
		}
	}
	return msg
}

// Unsubscribe removes the subscription for the given topic.
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

// Ack acknowledges a message. Required for JetStream manual ack mode.
func (s *Subscriber) Ack(_ context.Context, msg *mq.Message) error {
	if msg == nil || s.js == nil {
		return nil
	}
	s.mu.Lock()
	m := s.pendingAck[msg.ID]
	delete(s.pendingAck, msg.ID)
	s.mu.Unlock()
	if m == nil {
		return nil
	}
	return m.Ack()
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
