package kafka_test

import (
	"context"
	"testing"

	"github.com/EthanShen10086/voxera-kit/mq"
	kafkamq "github.com/EthanShen10086/voxera-kit/mq/kafka"
)

func TestKafkaConnectValidation(t *testing.T) {
	_, err := kafkamq.NewPublisher(mq.Config{})
	if err == nil {
		t.Fatal("expected error for empty brokers")
	}
	_, err = kafkamq.NewSubscriber(mq.Config{})
	if err == nil {
		t.Fatal("expected error for empty brokers")
	}
}

func TestKafkaPublishNilMessage(t *testing.T) {
	p, err := kafkamq.NewPublisher(mq.Config{Brokers: []string{"127.0.0.1:9092"}})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = p.Close() }()
	if err := p.Publish(context.Background(), "t", nil); err == nil {
		t.Fatal("expected nil message error")
	}
}

func TestKafkaSubscribeValidation(t *testing.T) {
	s, err := kafkamq.NewSubscriber(mq.Config{Brokers: []string{"127.0.0.1:9092"}, GroupID: "g1"})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = s.Close() }()
	if err := s.Subscribe(context.Background(), "t", nil); err == nil {
		t.Fatal("expected nil handler error")
	}
}
