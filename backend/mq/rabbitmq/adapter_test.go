package rabbitmq_test

import (
	"testing"

	"github.com/EthanShen10086/voxera-kit/mq"
	rabbitmqmq "github.com/EthanShen10086/voxera-kit/mq/rabbitmq"
)

func TestRabbitMQConnectValidation(t *testing.T) {
	_, err := rabbitmqmq.NewPublisher(mq.Config{})
	if err == nil {
		t.Fatal("expected error for empty brokers")
	}
	_, err = rabbitmqmq.NewSubscriber(mq.Config{})
	if err == nil {
		t.Fatal("expected error for empty brokers")
	}
}

func TestRabbitMQDialUnreachable(t *testing.T) {
	_, err := rabbitmqmq.NewPublisher(mq.Config{Brokers: []string{"127.0.0.1:1"}})
	if err == nil {
		t.Fatal("expected dial error")
	}
}
