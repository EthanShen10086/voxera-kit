package stub_test

import (
	"context"
	"testing"

	"github.com/EthanShen10086/voxera-kit/notification"
	"github.com/EthanShen10086/voxera-kit/notification/stub"
)

func TestStubNotifier(t *testing.T) {
	a := stub.New(notification.ChannelEmail)
	res, err := a.Send(context.Background(), &notification.Message{Title: "hi", Content: "body"})
	if err != nil || res.Status != notification.StatusDelivered {
		t.Fatalf("Send: %+v err=%v", res, err)
	}
	if a.Channel() != notification.ChannelEmail {
		t.Fatalf("Channel = %v", a.Channel())
	}
	if len(a.Messages) != 1 {
		t.Fatalf("messages = %d", len(a.Messages))
	}
}
