// Package memory provides an in-memory implementation of the messaging.MessagingService
// and messaging.PresenceService interfaces using maps and channel-based subscriptions.
package memory

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/EthanShen10086/voxera-kit/messaging"
)

type subscription struct {
	handler messaging.MessageHandler
}

// Adapter is an in-memory messaging and presence service.
type Adapter struct {
	mu            sync.RWMutex
	channels      map[string]*messaging.Channel
	messages      map[string][]*messaging.Message
	subscriptions map[string]map[uint64]*subscription
	presence      map[string]bool
	subCounter    uint64
	msgCounter    uint64
	chanCounter   uint64
}

// New creates a new in-memory messaging adapter.
func New() *Adapter {
	return &Adapter{
		channels:      make(map[string]*messaging.Channel),
		messages:      make(map[string][]*messaging.Message),
		subscriptions: make(map[string]map[uint64]*subscription),
		presence:      make(map[string]bool),
	}
}

// CreateChannel creates a new messaging channel with the given type, members, and name.
func (a *Adapter) CreateChannel(_ context.Context, channelType messaging.ChannelType, members []string, name string) (*messaging.Channel, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.chanCounter++
	ch := &messaging.Channel{
		ID:        fmt.Sprintf("ch_%d", a.chanCounter),
		Name:      name,
		Type:      channelType,
		Members:   members,
		CreatedAt: time.Now(),
	}
	a.channels[ch.ID] = ch
	return ch, nil
}

// SendMessage sends a message to the specified channel.
func (a *Adapter) SendMessage(_ context.Context, channelID string, msg *messaging.Message) error {
	a.mu.Lock()

	if _, exists := a.channels[channelID]; !exists {
		a.mu.Unlock()
		return fmt.Errorf("messaging: channel %q not found", channelID)
	}

	a.msgCounter++
	msg.ID = fmt.Sprintf("msg_%d", a.msgCounter)
	msg.ChannelID = channelID
	if msg.CreatedAt.IsZero() {
		msg.CreatedAt = time.Now()
	}
	a.messages[channelID] = append(a.messages[channelID], msg)

	subs := make([]*subscription, 0, len(a.subscriptions[channelID]))
	for _, sub := range a.subscriptions[channelID] {
		subs = append(subs, sub)
	}
	a.mu.Unlock()

	for _, sub := range subs {
		sub.handler(msg)
	}
	return nil
}

// Subscribe registers a handler for messages on the specified channel.
func (a *Adapter) Subscribe(channelID string, handler messaging.MessageHandler) (func(), error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if _, exists := a.channels[channelID]; !exists {
		return nil, fmt.Errorf("messaging: channel %q not found", channelID)
	}

	a.subCounter++
	id := a.subCounter

	if a.subscriptions[channelID] == nil {
		a.subscriptions[channelID] = make(map[uint64]*subscription)
	}
	a.subscriptions[channelID][id] = &subscription{handler: handler}

	unsubscribe := func() {
		a.mu.Lock()
		defer a.mu.Unlock()
		delete(a.subscriptions[channelID], id)
	}
	return unsubscribe, nil
}

// GetHistory retrieves messages from a channel before the given timestamp.
func (a *Adapter) GetHistory(_ context.Context, channelID string, before time.Time, limit int) ([]*messaging.Message, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	msgs := a.messages[channelID]
	var result []*messaging.Message
	for i := len(msgs) - 1; i >= 0 && len(result) < limit; i-- {
		if msgs[i].CreatedAt.Before(before) {
			result = append(result, msgs[i])
		}
	}
	return result, nil
}

// GetChannels returns all channels the given user is a member of.
func (a *Adapter) GetChannels(_ context.Context, userID string) ([]*messaging.Channel, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	var result []*messaging.Channel
	for _, ch := range a.channels {
		for _, member := range ch.Members {
			if member == userID {
				result = append(result, ch)
				break
			}
		}
	}
	return result, nil
}

// SetOnline marks the given user as online.
func (a *Adapter) SetOnline(_ context.Context, userID string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.presence[userID] = true
	return nil
}

// SetOffline marks the given user as offline.
func (a *Adapter) SetOffline(_ context.Context, userID string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	delete(a.presence, userID)
	return nil
}

// IsOnline reports whether the given user is currently online.
func (a *Adapter) IsOnline(_ context.Context, userID string) (bool, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.presence[userID], nil
}

// GetOnlineUsers returns the IDs of online users in the given channel.
func (a *Adapter) GetOnlineUsers(_ context.Context, channelID string) ([]string, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	ch, exists := a.channels[channelID]
	if !exists {
		return nil, fmt.Errorf("messaging: channel %q not found", channelID)
	}

	var online []string
	for _, member := range ch.Members {
		if a.presence[member] {
			online = append(online, member)
		}
	}
	return online, nil
}
