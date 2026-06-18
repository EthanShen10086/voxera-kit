package kafka

import (
	"context"
	"testing"
	"time"

	"github.com/EthanShen10086/voxera-kit/mq"
	kafkago "github.com/segmentio/kafka-go"
)

func TestDialerWithSCRAM(t *testing.T) {
	d := dialer(mq.Config{Username: "user", Password: "pass"})
	if d.SASLMechanism == nil {
		t.Fatal("expected SASL mechanism")
	}
}

func TestSubscriberAckValidation(t *testing.T) {
	s := &Subscriber{readers: make(map[string]*kafkaSubscription)}
	if err := s.Ack(context.Background(), nil); err == nil {
		t.Fatal("expected nil message error")
	}
	if err := s.Ack(context.Background(), &mq.Message{}); err == nil {
		t.Fatal("expected empty id error")
	}
}

func TestSubscriberAckMissingPending(t *testing.T) {
	s := &Subscriber{readers: make(map[string]*kafkaSubscription)}
	if err := s.Ack(context.Background(), &mq.Message{ID: "x", Topic: "t"}); err != nil {
		t.Fatalf("Ack() = %v", err)
	}
}

func TestSubscriberAckInvalidPending(t *testing.T) {
	s := &Subscriber{readers: map[string]*kafkaSubscription{"t": {}}}
	s.pending.Store("bad", "not-a-message")
	err := s.Ack(context.Background(), &mq.Message{ID: "bad", Topic: "t"})
	if err == nil {
		t.Fatal("expected invalid pending error")
	}
}

func TestPublishContextCancel(t *testing.T) {
	p := &Publisher{writer: &kafkago.Writer{Addr: kafkago.TCP("127.0.0.1:9092")}}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := p.Publish(ctx, "topic", &mq.Message{ID: "1", Payload: []byte("x"), Timestamp: time.Now()})
	if err == nil {
		t.Fatal("expected context error")
	}
}

func TestSubscriberUnsubscribeMissing(t *testing.T) {
	s := &Subscriber{readers: make(map[string]*kafkaSubscription)}
	if err := s.Unsubscribe("missing"); err != nil {
		t.Fatalf("Unsubscribe() = %v", err)
	}
}

func TestPublisherCloseNilWriter(t *testing.T) {
	p := &Publisher{}
	if err := p.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestSubscribeValidation(t *testing.T) {
	s := &Subscriber{cfg: mq.Config{Brokers: []string{"127.0.0.1:9092"}}, readers: make(map[string]*kafkaSubscription)}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := s.Subscribe(ctx, "t", func(context.Context, *mq.Message) error { return nil }); err == nil {
		t.Fatal("expected context error")
	}
	if err := s.Subscribe(context.Background(), "t", nil); err == nil {
		t.Fatal("expected nil handler error")
	}
}

func TestSubscriberCloseEmpty(t *testing.T) {
	s := &Subscriber{readers: make(map[string]*kafkaSubscription)}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestSubscribeDuplicateTopic(t *testing.T) {
	s := &Subscriber{cfg: mq.Config{Brokers: []string{"127.0.0.1:9092"}}, readers: make(map[string]*kafkaSubscription)}
	handler := func(context.Context, *mq.Message) error { return nil }
	if err := s.Subscribe(context.Background(), "t", handler); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = s.Unsubscribe("t") }()
	if err := s.Subscribe(context.Background(), "t", handler); err == nil {
		t.Fatal("expected duplicate subscribe error")
	}
}

func TestPublishMessageFields(t *testing.T) {
	p := &Publisher{writer: &kafkago.Writer{Addr: kafkago.TCP("127.0.0.1:9092")}}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	ts := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	err := p.Publish(ctx, "topic", &mq.Message{
		ID: "mid", Payload: []byte("p"), Timestamp: ts,
		Headers: map[string]string{"trace": "abc"},
	})
	if err == nil {
		t.Fatal("expected context error")
	}
}

func TestSubscriberAckNoReader(t *testing.T) {
	s := &Subscriber{readers: make(map[string]*kafkaSubscription)}
	s.pending.Store("id1", kafkago.Message{Topic: "t"})
	err := s.Ack(context.Background(), &mq.Message{ID: "id1", Topic: "t"})
	if err == nil {
		t.Fatal("expected no active reader error")
	}
}
