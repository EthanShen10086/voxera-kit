package nats_test

import (
	"context"
	"testing"
	"time"

	"github.com/EthanShen10086/voxera-kit/mq"
	"github.com/EthanShen10086/voxera-kit/mq/contract"
	natsmq "github.com/EthanShen10086/voxera-kit/mq/nats"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats-server/v2/test"
)

func startNATSServer(t *testing.T, jetStream bool) string {
	t.Helper()
	opts := &server.Options{
		Host:      "127.0.0.1",
		Port:      -1,
		JetStream: jetStream,
	}
	s := test.RunServer(opts)
	t.Cleanup(s.Shutdown)
	return s.ClientURL()
}

func TestNATSContract(t *testing.T) {
	url := startNATSServer(t, false)
	cfg := mq.Config{Brokers: []string{url}}

	contract.RunMQContract(t, func(t *testing.T) (mq.Publisher, mq.Subscriber, func()) {
		pub, err := natsmq.NewPublisher(cfg)
		if err != nil {
			t.Fatalf("NewPublisher: %v", err)
		}
		sub, err := natsmq.NewSubscriber(cfg)
		if err != nil {
			_ = pub.Close()
			t.Fatalf("NewSubscriber: %v", err)
		}
		return pub, sub, func() {}
	})
}

func TestNATSJetStreamPublishSubscribe(t *testing.T) {
	ctx := context.Background()
	url := startNATSServer(t, true)
	cfg := mq.Config{
		Brokers:   []string{url},
		JetStream: true,
		Stream:    "VOXERA_JS_UNIT",
		Durable:   "unit-durable",
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

	topic := "js.unit.test"
	done := make(chan struct{}, 1)
	if err := sub.Subscribe(ctx, topic, func(_ context.Context, msg *mq.Message) error {
		done <- struct{}{}
		return sub.Ack(ctx, msg)
	}); err != nil {
		t.Fatal(err)
	}

	if err := pub.Publish(ctx, topic, &mq.Message{ID: "m1", Payload: []byte("js")}); err != nil {
		t.Fatal(err)
	}
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("timeout")
	}
}

func TestNATSConnectValidation(t *testing.T) {
	_, err := natsmq.NewPublisher(mq.Config{})
	if err == nil {
		t.Fatal("expected error for empty brokers")
	}
}
