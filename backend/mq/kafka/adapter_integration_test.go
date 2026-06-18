//go:build integration

package kafka_test

import (
	"context"
	"testing"

	"github.com/EthanShen10086/voxera-kit/mq"
	"github.com/EthanShen10086/voxera-kit/mq/contract"
	kafkamq "github.com/EthanShen10086/voxera-kit/mq/kafka"
	"github.com/EthanShen10086/voxera-kit/testkit/containers"
)

func TestKafkaMQContract(t *testing.T) {
	ctx := context.Background()
	c, err := containers.StartKafka(ctx)
	if err != nil {
		t.Fatalf("StartKafka: %v", err)
	}
	t.Cleanup(func() { _ = c.Terminate(context.Background()) })

	cfg := mq.Config{Brokers: c.Brokers, GroupID: "voxera-kafka-test"}

	contract.RunMQContract(t, func(t *testing.T) (mq.Publisher, mq.Subscriber, func()) {
		pub, err := kafkamq.NewPublisher(cfg)
		if err != nil {
			t.Fatalf("NewPublisher: %v", err)
		}
		sub, err := kafkamq.NewSubscriber(cfg)
		if err != nil {
			_ = pub.Close()
			t.Fatalf("NewSubscriber: %v", err)
		}
		return pub, sub, func() {}
	})
}
