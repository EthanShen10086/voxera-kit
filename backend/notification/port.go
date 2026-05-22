// Package notification provides a pluggable notification delivery abstraction
// supporting multiple channels like WeChat Work, Feishu, Email, Slack, etc.
package notification

import (
	"context"
	"time"
)

// ChannelType identifies the notification channel type.
type ChannelType string

const (
	// ChannelWecom represents WeChat Work (企业微信) webhook channel.
	ChannelWecom ChannelType = "wecom"
	// ChannelFeishu represents Feishu (飞书) webhook channel.
	ChannelFeishu ChannelType = "feishu"
	// ChannelEmail represents email (SMTP) channel.
	ChannelEmail ChannelType = "email"
	// ChannelSlack represents Slack webhook channel.
	ChannelSlack ChannelType = "slack"
	// ChannelTelegram represents Telegram bot channel.
	ChannelTelegram ChannelType = "telegram"
)

// MessageFormat defines the format of notification messages.
type MessageFormat string

const (
	// FormatMarkdown represents markdown formatted messages.
	FormatMarkdown MessageFormat = "markdown"
	// FormatHTML represents HTML formatted messages.
	FormatHTML MessageFormat = "html"
	// FormatPlainText represents plain text messages.
	FormatPlainText MessageFormat = "text"
	// FormatCard represents interactive card messages.
	FormatCard MessageFormat = "card"
)

// Message represents a notification message to be delivered.
type Message struct {
	Title    string
	Content  string
	Format   MessageFormat
	URL      string
	ImageURL string
	Metadata map[string]any
}

// Config holds configuration for a notification channel.
type Config struct {
	Type       ChannelType
	WebhookURL string
	Token      string
	Secret     string
	Recipient  string
	Extra      map[string]any
}

// DeliveryStatus indicates the result of a notification delivery.
type DeliveryStatus string

const (
	// StatusDelivered indicates successful delivery.
	StatusDelivered DeliveryStatus = "delivered"
	// StatusFailed indicates delivery failure.
	StatusFailed DeliveryStatus = "failed"
	// StatusRateLimited indicates the channel is rate limiting.
	StatusRateLimited DeliveryStatus = "rate_limited"
	// StatusRetrying indicates the delivery is being retried.
	StatusRetrying DeliveryStatus = "retrying"
)

// DeliveryResult contains the outcome of a notification delivery attempt.
type DeliveryResult struct {
	Status    DeliveryStatus
	Channel   ChannelType
	Timestamp time.Time
	Error     string
	RetryAt   *time.Time
}

// Notifier sends notifications through a specific channel.
type Notifier interface {
	Send(ctx context.Context, msg *Message) (*DeliveryResult, error)
	Channel() ChannelType
}

// Router dispatches messages to multiple notifiers based on rules.
type Router interface {
	Register(notifier Notifier)
	Send(ctx context.Context, channelType ChannelType, msg *Message) (*DeliveryResult, error)
	SendAll(ctx context.Context, msg *Message) ([]*DeliveryResult, error)
}
