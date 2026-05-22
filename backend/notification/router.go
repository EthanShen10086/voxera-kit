package notification

import (
	"context"
	"fmt"
)

// DefaultRouter dispatches messages to registered notifiers.
type DefaultRouter struct {
	notifiers map[ChannelType]Notifier
}

// NewRouter creates a new default notification router.
func NewRouter() *DefaultRouter {
	return &DefaultRouter{notifiers: make(map[ChannelType]Notifier)}
}

// Register adds a notifier to the router.
func (r *DefaultRouter) Register(n Notifier) {
	r.notifiers[n.Channel()] = n
}

// Send dispatches to a specific channel.
func (r *DefaultRouter) Send(ctx context.Context, ct ChannelType, msg *Message) (*DeliveryResult, error) {
	n, ok := r.notifiers[ct]
	if !ok {
		return nil, fmt.Errorf("notification: no notifier registered for channel %q", ct)
	}
	return n.Send(ctx, msg)
}

// SendAll dispatches to all registered channels.
func (r *DefaultRouter) SendAll(ctx context.Context, msg *Message) ([]*DeliveryResult, error) {
	results := make([]*DeliveryResult, 0, len(r.notifiers))
	for _, n := range r.notifiers {
		result, err := n.Send(ctx, msg)
		if err != nil {
			return results, err
		}
		results = append(results, result)
	}
	return results, nil
}
