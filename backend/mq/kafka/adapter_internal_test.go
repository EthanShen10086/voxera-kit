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
