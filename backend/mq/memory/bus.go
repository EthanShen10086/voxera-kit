package memory

import (
	"context"
	"sync"

	"github.com/EthanShen10086/voxera-kit/mq"
)

type subscription struct {
	handler mq.MessageHandler
	cancel  context.CancelFunc
}

// Bus is the shared message broker backing memory publishers and subscribers.
type Bus struct {
	mu            sync.RWMutex
	subscriptions map[string][]*subscription
}

// NewBus creates a new in-process message bus.
func NewBus() *Bus {
	return &Bus{subscriptions: make(map[string][]*subscription)}
}
