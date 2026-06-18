//go:build integration

package kafka_test

import (
	"context"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/EthanShen10086/voxera-kit/mq"
	"github.com/EthanShen10086/voxera-kit/mq/contract"
	kafkamq "github.com/EthanShen10086/voxera-kit/mq/kafka"
	"github.com/EthanShen10086/voxera-kit/testkit/containers"
	"github.com/segmentio/kafka-go"
)

func TestKafkaMQContract(t *testing.T) {
	ctx := context.Background()
	c, err := containers.StartKafka(ctx)
	if err != nil {
		t.Fatalf("StartKafka: %v", err)
	}
	t.Cleanup(func() { _ = c.Terminate(context.Background()) })

	ensureKafkaTopics(t, c.Brokers, "contract-topic", "noop-check")

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
	}, contract.WithPostSubscribeDelay(2*time.Second), contract.WithReceiveTimeout(5*time.Second))
}

func ensureKafkaTopics(t *testing.T, brokers []string, topics ...string) {
	t.Helper()
	if len(brokers) == 0 {
		t.Fatal("no kafka brokers")
	}
	ctx := context.Background()
	conn, err := kafka.DialContext(ctx, "tcp", brokers[0])
	if err != nil {
		t.Fatalf("dial kafka: %v", err)
	}
	defer conn.Close()

	controller, err := conn.Controller()
	if err != nil {
		t.Fatalf("kafka controller: %v", err)
	}
	ctrlConn, err := kafka.DialContext(ctx, "tcp", net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port)))
	if err != nil {
		t.Fatalf("dial kafka controller: %v", err)
	}
	defer ctrlConn.Close()

	configs := make([]kafka.TopicConfig, len(topics))
	for i, topic := range topics {
		configs[i] = kafka.TopicConfig{
			Topic:             topic,
			NumPartitions:     1,
			ReplicationFactor: 1,
		}
	}
	if err := ctrlConn.CreateTopics(configs...); err != nil {
		t.Fatalf("create topics: %v", err)
	}
}
