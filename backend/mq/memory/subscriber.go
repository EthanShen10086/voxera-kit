package memory

import (
	"context"
	"fmt"

	"github.com/EthanShen10086/voxera-kit/mq"
)

// Subscriber implements mq.Subscriber using an in-process bus.
type Subscriber struct {
	bus *Bus
}

// NewSubscriber returns a subscriber connected to the given bus.
func NewSubscriber(bus *Bus) *Subscriber {
	return &Subscriber{bus: bus}
}

// Subscribe registers a handler for messages on the given topic.
func (s *Subscriber) Subscribe(ctx context.Context, topic string, handler mq.MessageHandler) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if handler == nil {
		return fmt.Errorf("mq/memory: handler is nil")
	}

	subCtx, cancel := context.WithCancel(ctx)
	sub := &subscription{handler: handler, cancel: cancel}

	s.bus.mu.Lock()
	s.bus.subscriptions[topic] = append(s.bus.subscriptions[topic], sub)
	s.bus.mu.Unlock()

	go func() {
		<-subCtx.Done()
		s.removeSubscription(topic, sub)
	}()
	return nil
}

// Unsubscribe removes all handlers registered for the topic.
func (s *Subscriber) Unsubscribe(topic string) error {
	s.bus.mu.Lock()
	defer s.bus.mu.Unlock()

	for _, sub := range s.bus.subscriptions[topic] {
		if sub.cancel != nil {
			sub.cancel()
		}
	}
	delete(s.bus.subscriptions, topic)
	return nil
}

// Ack is a no-op for the in-process subscriber.
func (s *Subscriber) Ack(_ context.Context, _ *mq.Message) error {
	return nil
}

// Close releases subscriber resources.
func (s *Subscriber) Close() error {
	s.bus.mu.Lock()
	defer s.bus.mu.Unlock()

	for topic, subs := range s.bus.subscriptions {
		for _, sub := range subs {
			if sub.cancel != nil {
				sub.cancel()
			}
		}
		delete(s.bus.subscriptions, topic)
	}
	return nil
}

func (s *Subscriber) removeSubscription(topic string, target *subscription) {
	s.bus.mu.Lock()
	defer s.bus.mu.Unlock()

	subs := s.bus.subscriptions[topic]
	filtered := subs[:0]
	for _, sub := range subs {
		if sub != target {
			filtered = append(filtered, sub)
		}
	}
	if len(filtered) == 0 {
		delete(s.bus.subscriptions, topic)
		return
	}
	s.bus.subscriptions[topic] = filtered
}
