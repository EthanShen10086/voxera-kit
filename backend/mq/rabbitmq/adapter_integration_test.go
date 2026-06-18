//go:build integration

package rabbitmq_test

import (
	"context"
	"testing"

	"github.com/EthanShen10086/voxera-kit/mq"
	"github.com/EthanShen10086/voxera-kit/mq/contract"
	rabbitmqmq "github.com/EthanShen10086/voxera-kit/mq/rabbitmq"
	"github.com/EthanShen10086/voxera-kit/testkit/containers"
)

func TestRabbitMQContract(t *testing.T) {
	ctx := context.Background()
	c, err := containers.StartRabbitMQ(ctx)
	if err != nil {
		t.Fatalf("StartRabbitMQ: %v", err)
	}
	t.Cleanup(func() { _ = c.Terminate(context.Background()) })

	cfg := mq.Config{Brokers: []string{c.URL}}

	contract.RunMQContract(t, func(t *testing.T) (mq.Publisher, mq.Subscriber, func()) {
		pub, err := rabbitmqmq.NewPublisher(cfg)
		if err != nil {
			t.Fatalf("NewPublisher: %v", err)
		}
		sub, err := rabbitmqmq.NewSubscriber(cfg)
		if err != nil {
			_ = pub.Close()
			t.Fatalf("NewSubscriber: %v", err)
		}
		return pub, sub, func() {}
	})
}
