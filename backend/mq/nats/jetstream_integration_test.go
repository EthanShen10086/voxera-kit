//go:build integration

package nats_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/EthanShen10086/voxera-kit/mq"
	"github.com/EthanShen10086/voxera-kit/mq/contract"
	natsmq "github.com/EthanShen10086/voxera-kit/mq/nats"
	"github.com/EthanShen10086/voxera-kit/testkit/containers"
)

func TestJetStreamContract(t *testing.T) {
	ctx := context.Background()
	url, cleanup := containers.StartNATS(ctx, t)
	defer cleanup()

	cfg := mq.Config{
		Brokers:   []string{url},
		JetStream: true,
		Stream:    "TEST_VOXERA",
		Durable:   "test-durable",
		GroupID:   "test-group",
	}

	contract.RunMQContract(t, func(t *testing.T) (mq.Publisher, mq.Subscriber, func()) {
		pub, err := natsmq.NewPublisher(cfg)
		if err != nil {
			t.Fatalf("NewPublisher: %v", err)
		}
		sub, err := natsmq.NewSubscriber(cfg)
		if err != nil {
			t.Fatalf("NewSubscriber: %v", err)
		}
		return pub, sub, func() {}
	})
}

func TestJetStreamManualAck(t *testing.T) {
	ctx := context.Background()
	url, cleanup := containers.StartNATS(ctx, t)
	defer cleanup()

	cfg := mq.Config{
		Brokers:   []string{url},
		JetStream: true,
		Stream:    "TEST_ACK",
		Durable:   "ack-durable",
	}

	pub, err := natsmq.NewPublisher(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = pub.Close() }()

	sub, err := natsmq.NewSubscriber(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = sub.Close() }()

	topic := "js.ack.test"
	var (
		mu   sync.Mutex
		got  *mq.Message
		done = make(chan struct{}, 1)
	)

	if err := sub.Subscribe(ctx, topic, func(_ context.Context, msg *mq.Message) error {
		mu.Lock()
		got = msg
		mu.Unlock()
		done <- struct{}{}
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	msg := &mq.Message{ID: "msg-1", Payload: []byte("js-payload")}
	if err := pub.Publish(ctx, topic, msg); err != nil {
		t.Fatal(err)
	}

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("timeout waiting for jetstream message")
	}

	mu.Lock()
	defer mu.Unlock()
	if got == nil || string(got.Payload) != "js-payload" {
		t.Fatalf("got = %+v", got)
	}
	if err := sub.Ack(ctx, got); err != nil {
		t.Fatalf("Ack: %v", err)
	}
}
