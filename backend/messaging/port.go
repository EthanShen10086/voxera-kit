// Package messaging defines the port interface for IM/real-time messaging channels.
// It abstracts away the underlying messaging transport and storage, allowing
// different backends to be used interchangeably.
package messaging

import (
	"context"
	"time"
)

// MessageType represents the kind of content in a message.
type MessageType int

const (
	// Text is a plain text message.
	Text MessageType = iota
	// Image is an image attachment message.
	Image
	// File is a file attachment message.
	File
	// System is an automated system notification.
	System
	// Custom is an application-defined message type.
	Custom
)

// ChannelType represents the kind of messaging channel.
type ChannelType int

const (
	// Direct is a one-on-one private channel.
	Direct ChannelType = iota
	// Group is a multi-member group channel.
	Group
	// Broadcast is a one-to-many broadcast channel.
	Broadcast
)

// Message represents a single message in a channel.
type Message struct {
	ID        string
	ChannelID string
	SenderID  string
	Type      MessageType
	Content   string
	Metadata  map[string]string
	CreatedAt time.Time
}

// Channel represents a messaging channel.
type Channel struct {
	ID        string
	Name      string
	Type      ChannelType
	Members   []string
	CreatedAt time.Time
}

// MessageHandler is a callback invoked when a message is received.
type MessageHandler func(msg *Message)

// MessagingService is the interface for channel and message operations.
// Implementations must be safe for concurrent use.
type MessagingService interface {
	// CreateChannel creates a new messaging channel with the given type, members, and name.
	CreateChannel(ctx context.Context, channelType ChannelType, members []string, name string) (*Channel, error)
	// SendMessage sends a message to the specified channel.
	SendMessage(ctx context.Context, channelID string, msg *Message) error
	// Subscribe registers a handler for messages on the specified channel.
	// Returns an unsubscribe function and any error encountered.
	Subscribe(channelID string, handler MessageHandler) (unsubscribe func(), err error)
	// GetHistory retrieves messages from a channel before the given timestamp.
	GetHistory(ctx context.Context, channelID string, before time.Time, limit int) ([]*Message, error)
	// GetChannels returns all channels the given user is a member of.
	GetChannels(ctx context.Context, userID string) ([]*Channel, error)
}

// PresenceService is the interface for tracking user online/offline status.
// Implementations must be safe for concurrent use.
type PresenceService interface {
	// SetOnline marks a user as online.
	SetOnline(ctx context.Context, userID string) error
	// SetOffline marks a user as offline.
	SetOffline(ctx context.Context, userID string) error
	// IsOnline reports whether the given user is currently online.
	IsOnline(ctx context.Context, userID string) (bool, error)
	// GetOnlineUsers returns the IDs of online users in the given channel.
	GetOnlineUsers(ctx context.Context, channelID string) ([]string, error)
}

// MessagingConfig holds configuration parameters for a messaging backend.
type MessagingConfig struct {
	// MaxMessageSize is the maximum allowed message size in bytes.
	MaxMessageSize int
	// MaxChannelMembers is the maximum number of members per channel.
	MaxChannelMembers int
	// HistoryRetention is how long messages are retained before expiry.
	HistoryRetention time.Duration
}
