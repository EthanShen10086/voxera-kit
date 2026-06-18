package rabbitmq

import (
	"context"
	"testing"
	"time"

	"github.com/EthanShen10086/voxera-kit/mq"
	amqp "github.com/rabbitmq/amqp091-go"
)

func TestDeliveryToMessage(t *testing.T) {
	ts := time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)
	d := amqp.Delivery{
		MessageId: "mid",
		Body:      []byte("payload"),
		Timestamp: ts,
		Headers: amqp.Table{
			"id":        "hdr-id",
			"timestamp": ts.Format(time.RFC3339Nano),
			"trace":     "abc",
		},
	}
	msg := deliveryToMessage("queue", d)
	if msg.ID != "hdr-id" {
		t.Fatalf("ID = %q", msg.ID)
	}
	if string(msg.Payload) != "payload" || msg.Topic != "queue" {
		t.Fatalf("msg = %+v", msg)
	}
	if msg.Headers["trace"] != "abc" {
		t.Fatalf("headers = %+v", msg.Headers)
	}
	if !msg.Timestamp.Equal(ts) {
		t.Fatalf("timestamp = %v", msg.Timestamp)
	}
}

func TestDeliveryToMessageFallbackID(t *testing.T) {
	d := amqp.Delivery{DeliveryTag: 42, Body: []byte("x")}
	msg := deliveryToMessage("q", d)
	if msg.ID != "amqp-42" {
		t.Fatalf("ID = %q", msg.ID)
	}
	if msg.Timestamp.IsZero() {
		t.Fatal("expected generated timestamp")
	}
}

func TestDialBuildsURL(t *testing.T) {
	_, err := dial(mq.Config{Brokers: []string{}})
	if err == nil {
		t.Fatal("expected broker error")
	}
	// unreachable host; we only assert dial attempts URL construction.
	_, err = dial(mq.Config{Brokers: []string{"127.0.0.1:59999"}, Username: "u", Password: "p"})
	if err == nil {
		t.Fatal("expected dial error")
	}
}

func TestSubscriberAckValidation(t *testing.T) {
	s := &Subscriber{}
	if err := s.Ack(context.Background(), nil); err == nil {
		t.Fatal("expected nil message error")
	}
	if err := s.Ack(context.Background(), &mq.Message{}); err == nil {
		t.Fatal("expected empty id error")
	}
}

func TestSubscriberAckMissingPending(t *testing.T) {
	s := &Subscriber{}
	if err := s.Ack(context.Background(), &mq.Message{ID: "missing"}); err != nil {
		t.Fatalf("Ack() = %v", err)
	}
}

func TestPublisherPublishValidation(t *testing.T) {
	p := &Publisher{ch: nil}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := p.Publish(ctx, "t", &mq.Message{ID: "1", Payload: []byte("x")}); err == nil {
		t.Fatal("expected ctx error")
	}
	if err := p.Publish(context.Background(), "t", nil); err == nil {
		t.Fatal("expected nil message error")
	}
}

func TestPublisherCloseNil(t *testing.T) {
	p := &Publisher{}
	if err := p.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestSubscriberAckInvalidPending(t *testing.T) {
	s := &Subscriber{}
	s.pending.Store("bad", "not-delivery")
	err := s.Ack(context.Background(), &mq.Message{ID: "bad"})
	if err == nil {
		t.Fatal("expected invalid pending error")
	}
}

func TestSubscriberUnsubscribeMissing(t *testing.T) {
	s := &Subscriber{}
	if err := s.Unsubscribe("missing"); err != nil {
		t.Fatalf("Unsubscribe() = %v", err)
	}
}

func TestSubscriberCloseEmpty(t *testing.T) {
	s := &Subscriber{}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestSubscribeValidation(t *testing.T) {
	s := &Subscriber{}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := s.Subscribe(ctx, "q", func(context.Context, *mq.Message) error { return nil }); err == nil {
		t.Fatal("expected context error")
	}
	if err := s.Subscribe(context.Background(), "q", nil); err == nil {
		t.Fatal("expected nil handler error")
	}
}
