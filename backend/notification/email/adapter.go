// Package email provides an SMTP-based email notifier.
package email

import (
	"context"
	"fmt"
	"net/smtp"
	"time"

	"github.com/EthanShen10086/voxera-kit/notification"
)

// Adapter implements notification.Notifier for SMTP email delivery.
type Adapter struct {
	host     string
	port     string
	username string
	password string
	from     string
	to       string
}

// New creates a new email notifier from the provided config.
// Expected Extra keys: "host", "port", "username", "password", "from".
func New(cfg notification.Config) *Adapter {
	extra := cfg.Extra
	return &Adapter{
		host:     strOrDefault(extra, "host", "localhost"),
		port:     strOrDefault(extra, "port", "587"),
		username: strOrDefault(extra, "username", ""),
		password: strOrDefault(extra, "password", ""),
		from:     strOrDefault(extra, "from", ""),
		to:       cfg.Recipient,
	}
}

// Send delivers a message via SMTP email.
func (a *Adapter) Send(_ context.Context, msg *notification.Message) (*notification.DeliveryResult, error) {
	addr := fmt.Sprintf("%s:%s", a.host, a.port)

	subject := msg.Title
	if subject == "" {
		subject = "Notification"
	}

	body := fmt.Sprintf("Subject: %s\r\nFrom: %s\r\nTo: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
		subject, a.from, a.to, msg.Content)

	var auth smtp.Auth
	if a.username != "" {
		auth = smtp.PlainAuth("", a.username, a.password, a.host)
	}

	err := smtp.SendMail(addr, auth, a.from, []string{a.to}, []byte(body))
	if err != nil {
		return &notification.DeliveryResult{
			Status:    notification.StatusFailed,
			Channel:   notification.ChannelEmail,
			Timestamp: time.Now(),
			Error:     err.Error(),
		}, nil
	}

	return &notification.DeliveryResult{
		Status:    notification.StatusDelivered,
		Channel:   notification.ChannelEmail,
		Timestamp: time.Now(),
	}, nil
}

// Channel returns the channel type.
func (a *Adapter) Channel() notification.ChannelType {
	return notification.ChannelEmail
}

func strOrDefault(m map[string]any, key, fallback string) string {
	if m == nil {
		return fallback
	}
	v, ok := m[key]
	if !ok {
		return fallback
	}
	s, ok := v.(string)
	if !ok {
		return fallback
	}
	return s
}
