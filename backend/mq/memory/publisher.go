package memory

import (
	"context"
	"fmt"
	"time"

	"github.com/EthanShen10086/voxera-kit/mq"
)

// Publisher implements mq.Publisher using an in-process bus.
type Publisher struct {
	bus *Bus
}

// NewPublisher returns a publisher connected to the given bus.
func NewPublisher(bus *Bus) *Publisher {
	return &Publisher{bus: bus}
}

// Publish delivers a message to all subscribers on the topic.
func (p *Publisher) Publish(ctx context.Context, topic string, msg *mq.Message) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if msg == nil {
		return fmt.Errorf("mq/memory: message is nil")
	}

	out := *msg
	if out.Topic == "" {
		out.Topic = topic
	}
	if out.Timestamp.IsZero() {
		out.Timestamp = time.Now()
	}
	if out.ID == "" {
		out.ID = fmt.Sprintf("mem-%d", time.Now().UnixNano())
	}

	p.bus.mu.RLock()
	subs := append([]*subscription(nil), p.bus.subscriptions[topic]...)
	p.bus.mu.RUnlock()

	for _, sub := range subs {
		if err := sub.handler(ctx, &out); err != nil {
			return err
		}
	}
	return nil
}

// Close releases publisher resources.
func (p *Publisher) Close() error {
	return nil
}
