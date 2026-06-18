package memory_test

import (
	"testing"

	"github.com/EthanShen10086/voxera-kit/mq"
	"github.com/EthanShen10086/voxera-kit/mq/contract"
	"github.com/EthanShen10086/voxera-kit/mq/memory"
)

func TestMQContract_Memory(t *testing.T) {
	contract.RunMQContract(t, func(t *testing.T) (mq.Publisher, mq.Subscriber, func()) {
		bus := memory.NewBus()
		return memory.NewPublisher(bus), memory.NewSubscriber(bus), nil
	})
}
