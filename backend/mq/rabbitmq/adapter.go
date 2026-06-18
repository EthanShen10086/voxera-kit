// Package rabbitmq provides a RabbitMQ implementation of the mq.Publisher and mq.Subscriber interfaces.
// It uses github.com/rabbitmq/amqp091-go as the underlying client.
package rabbitmq

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/EthanShen10086/voxera-kit/mq"
	amqp "github.com/rabbitmq/amqp091-go"
)

func dial(cfg mq.Config) (*amqp.Connection, error) {
	if len(cfg.Brokers) == 0 {
		return nil, fmt.Errorf("rabbitmq: at least one broker address is required")
	}
	broker := cfg.Brokers[0]
	if !strings.HasPrefix(broker, "amqp://") && !strings.HasPrefix(broker, "amqps://") {
		if cfg.Username != "" {
			broker = fmt.Sprintf("amqp://%s:%s@%s/", cfg.Username, cfg.Password, broker)
		} else {
			broker = fmt.Sprintf("amqp://%s/", broker)
		}
	}
	return amqp.Dial(broker)
}

func openChannel(conn *amqp.Connection) (*amqp.Channel, error) {
	return conn.Channel()
}

// Publisher implements the mq.Publisher interface using RabbitMQ.
type Publisher struct {
	conn *amqp.Connection
	ch   *amqp.Channel
}

// NewPublisher creates a new RabbitMQ Publisher with the provided configuration.
func NewPublisher(cfg mq.Config) (*Publisher, error) {
	conn, err := dial(cfg)
	if err != nil {
		return nil, err
	}
	ch, err := openChannel(conn)
	if err != nil {
		_ = conn.Close()
		return nil, err
	}
	return &Publisher{conn: conn, ch: ch}, nil
}

// Publish sends a message to the specified RabbitMQ exchange/routing key.
func (p *Publisher) Publish(ctx context.Context, topic string, msg *mq.Message) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if msg == nil {
		return fmt.Errorf("rabbitmq: message is nil")
	}

	headers := amqp.Table{}
	for k, v := range msg.Headers {
		headers[k] = v
	}
	if msg.ID != "" {
		headers["id"] = msg.ID
	}
	if !msg.Timestamp.IsZero() {
		headers["timestamp"] = msg.Timestamp.Format(time.RFC3339Nano)
	}

	return p.ch.PublishWithContext(ctx, "", topic, false, false, amqp.Publishing{
		ContentType: "application/octet-stream",
		Body:        msg.Payload,
		Headers:     headers,
		Timestamp:   msg.Timestamp,
		MessageId:   msg.ID,
	})
}

// Close shuts down the RabbitMQ publisher channel and connection.
func (p *Publisher) Close() error {
	if p.ch != nil {
		_ = p.ch.Close()
	}
	if p.conn != nil {
		return p.conn.Close()
	}
	return nil
}

type rabbitSubscription struct {
	tag     string
	cancel  context.CancelFunc
}

// Subscriber implements the mq.Subscriber interface using RabbitMQ.
type Subscriber struct {
	conn     *amqp.Connection
	ch       *amqp.Channel
	mu       sync.Mutex
	subs     map[string]*rabbitSubscription
	pending  sync.Map // msg.ID -> amqp.Delivery
}

// NewSubscriber creates a new RabbitMQ Subscriber with the provided configuration.
func NewSubscriber(cfg mq.Config) (*Subscriber, error) {
	conn, err := dial(cfg)
	if err != nil {
		return nil, err
	}
	ch, err := openChannel(conn)
	if err != nil {
		_ = conn.Close()
		return nil, err
	}
	return &Subscriber{
		conn: conn,
		ch:   ch,
		subs: make(map[string]*rabbitSubscription),
	}, nil
}

// Subscribe registers a handler for messages on the given RabbitMQ queue.
func (s *Subscriber) Subscribe(ctx context.Context, topic string, handler mq.MessageHandler) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if handler == nil {
		return fmt.Errorf("rabbitmq: handler is nil")
	}

	if _, err := s.ch.QueueDeclare(topic, true, false, false, false, nil); err != nil {
		return err
	}

	tag := fmt.Sprintf("sub-%s-%d", topic, time.Now().UnixNano())
	deliveries, err := s.ch.Consume(topic, tag, false, false, false, false, nil)
	if err != nil {
		return err
	}

	subCtx, cancel := context.WithCancel(ctx)
	s.mu.Lock()
	s.subs[topic] = &rabbitSubscription{tag: tag, cancel: cancel}
	s.mu.Unlock()

	go func() {
		for {
			select {
			case <-subCtx.Done():
				return
			case d, ok := <-deliveries:
				if !ok {
					return
				}
				msg := deliveryToMessage(topic, d)
				s.pending.Store(msg.ID, d)
				if err := handler(subCtx, msg); err == nil {
					s.pending.Delete(msg.ID)
				}
			}
		}
	}()
	return nil
}

func deliveryToMessage(topic string, d amqp.Delivery) *mq.Message {
	msg := &mq.Message{
		ID:        d.MessageId,
		Topic:     topic,
		Payload:   append([]byte(nil), d.Body...),
		Headers:   make(map[string]string),
		Timestamp: d.Timestamp,
	}
	if msg.ID == "" {
		msg.ID = fmt.Sprintf("amqp-%d", d.DeliveryTag)
	}
	for k, v := range d.Headers {
		if s, ok := v.(string); ok {
			switch k {
			case "id":
				msg.ID = s
			case "timestamp":
				if parsed, err := time.Parse(time.RFC3339Nano, s); err == nil {
					msg.Timestamp = parsed
				}
			default:
				msg.Headers[k] = s
			}
		}
	}
	if msg.Timestamp.IsZero() {
		msg.Timestamp = time.Now()
	}
	return msg
}

// Unsubscribe cancels the consumer for the given queue.
func (s *Subscriber) Unsubscribe(topic string) error {
	s.mu.Lock()
	sub, ok := s.subs[topic]
	if ok {
		if sub.cancel != nil {
			sub.cancel()
		}
		delete(s.subs, topic)
	}
	s.mu.Unlock()
	if !ok {
		return nil
	}
	return s.ch.Cancel(sub.tag, false)
}

// Ack acknowledges successful processing of a message.
func (s *Subscriber) Ack(_ context.Context, msg *mq.Message) error {
	if msg == nil || msg.ID == "" {
		return fmt.Errorf("rabbitmq: message id is required for ack")
	}
	raw, ok := s.pending.Load(msg.ID)
	if !ok {
		return nil
	}
	d, ok := raw.(amqp.Delivery)
	if !ok {
		return fmt.Errorf("rabbitmq: invalid pending message for id %q", msg.ID)
	}
	err := d.Ack(false)
	if err == nil {
		s.pending.Delete(msg.ID)
	}
	return err
}

// Close shuts down the RabbitMQ subscriber channel and connection.
func (s *Subscriber) Close() error {
	s.mu.Lock()
	for topic, sub := range s.subs {
		if sub.cancel != nil {
			sub.cancel()
		}
		delete(s.subs, topic)
	}
	s.mu.Unlock()

	if s.ch != nil {
		_ = s.ch.Close()
	}
	if s.conn != nil {
		return s.conn.Close()
	}
	return nil
}
