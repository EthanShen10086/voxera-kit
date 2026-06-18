package contract

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/EthanShen10086/voxera-kit/mq"
	"github.com/EthanShen10086/voxera-kit/mq/memory"
)

// Factory creates a publisher, subscriber, and optional cleanup for contract tests.
type Factory func(t *testing.T) (mq.Publisher, mq.Subscriber, func())

// RunMQContract exercises publish/subscribe behavior for mq adapters.
func RunMQContract(t *testing.T, factory Factory) {
	t.Helper()
	ctx := context.Background()

	pub, sub, cleanup := factory(t)
	if cleanup != nil {
		defer cleanup()
	}
	defer func() { _ = pub.Close() }()
	defer func() { _ = sub.Close() }()

	t.Run("PublishSubscribeRoundtrip", func(t *testing.T) {
		topic := "contract-topic"
		payload := []byte("hello-mq")
		var (
			mu       sync.Mutex
			received *mq.Message
			done     = make(chan struct{})
		)

		err := sub.Subscribe(ctx, topic, func(_ context.Context, msg *mq.Message) error {
			mu.Lock()
			received = msg
			mu.Unlock()
			close(done)
			return nil
		})
		if err != nil {
			t.Fatalf("Subscribe: %v", err)
		}

		msg := &mq.Message{Payload: payload, Headers: map[string]string{"k": "v"}}
		if err := pub.Publish(ctx, topic, msg); err != nil {
			t.Fatalf("Publish: %v", err)
		}

		select {
		case <-done:
		case <-time.After(time.Second):
			t.Fatal("timed out waiting for message")
		}

		mu.Lock()
		defer mu.Unlock()
		if received == nil {
			t.Fatal("no message received")
		}
		if string(received.Payload) != string(payload) {
			t.Fatalf("payload = %q, want %q", received.Payload, payload)
		}
		if received.Topic != topic {
			t.Fatalf("topic = %q, want %q", received.Topic, topic)
		}
		if received.Headers["k"] != "v" {
			t.Fatalf("header k = %q, want %q", received.Headers["k"], "v")
		}
	})

	t.Run("PublishDoesNotNoOp", func(t *testing.T) {
		ch := make(chan struct{}, 1)
		topic := "noop-check"
		_ = sub.Subscribe(ctx, topic, func(_ context.Context, _ *mq.Message) error {
			select {
			case ch <- struct{}{}:
			default:
			}
			return nil
		})

		if err := pub.Publish(ctx, topic, &mq.Message{Payload: []byte("ping")}); err != nil {
			t.Fatalf("Publish: %v", err)
		}

		select {
		case <-ch:
		case <-time.After(time.Second):
			t.Fatal("Publish appears to be a no-op; subscriber received nothing")
		}
	})
}

func memoryFactory(t *testing.T) (mq.Publisher, mq.Subscriber, func()) {
	t.Helper()
	bus := memory.NewBus()
	return memory.NewPublisher(bus), memory.NewSubscriber(bus), nil
}
