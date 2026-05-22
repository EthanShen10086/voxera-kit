// Package stub provides a test stub notifier that logs messages and always succeeds.
package stub

import (
	"context"
	"log"
	"time"

	"github.com/EthanShen10086/voxera-kit/notification"
)

// Adapter implements notification.Notifier as a no-op test stub.
type Adapter struct {
	channel  notification.ChannelType
	Messages []*notification.Message
}

// New creates a new stub notifier for the given channel type.
func New(channel notification.ChannelType) *Adapter {
	return &Adapter{
		channel:  channel,
		Messages: make([]*notification.Message, 0),
	}
}

// Send logs the message and returns a successful delivery result.
func (a *Adapter) Send(_ context.Context, msg *notification.Message) (*notification.DeliveryResult, error) {
	log.Printf("stub[%s]: title=%q content=%q", a.channel, msg.Title, msg.Content)
	a.Messages = append(a.Messages, msg)
	return &notification.DeliveryResult{
		Status:    notification.StatusDelivered,
		Channel:   a.channel,
		Timestamp: time.Now(),
	}, nil
}

// Channel returns the channel type.
func (a *Adapter) Channel() notification.ChannelType {
	return a.channel
}
