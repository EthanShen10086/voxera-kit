package email_test

import (
	"context"
	"testing"

	"github.com/EthanShen10086/voxera-kit/notification"
	"github.com/EthanShen10086/voxera-kit/notification/email"
)

func TestSendFailure(t *testing.T) {
	a := email.New(notification.Config{
		Recipient: "user@example.com",
		Extra: map[string]any{
			"host": "127.0.0.1", "port": "1", "from": "noreply@example.com",
		},
	})
	res, err := a.Send(context.Background(), &notification.Message{Title: "hi", Content: "body"})
	if err != nil || res.Status != notification.StatusFailed {
		t.Fatalf("Send: %+v err=%v", res, err)
	}
	if a.Channel() != notification.ChannelEmail {
		t.Fatalf("Channel = %v", a.Channel())
	}
}

func TestNewDefaults(t *testing.T) {
	a := email.New(notification.Config{Recipient: "to@example.com"})
	if a.Channel() != notification.ChannelEmail {
		t.Fatal("expected email channel")
	}
}
