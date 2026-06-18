package memory_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/EthanShen10086/voxera-kit/messaging"
	"github.com/EthanShen10086/voxera-kit/messaging/memory"
)

func TestMessagingFlow(t *testing.T) {
	ctx := context.Background()
	a := memory.New()

	ch, err := a.CreateChannel(ctx, messaging.Group, []string{"u1", "u2"}, "general")
	if err != nil {
		t.Fatal(err)
	}

	var received atomic.Int32
	unsub, err := a.Subscribe(ch.ID, func(msg *messaging.Message) {
		received.Add(1)
		if msg.Content != "hello" {
			t.Errorf("content = %q", msg.Content)
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	defer unsub()

	if err := a.SendMessage(ctx, ch.ID, &messaging.Message{SenderID: "u1", Content: "hello"}); err != nil {
		t.Fatal(err)
	}
	time.Sleep(20 * time.Millisecond)
	if received.Load() != 1 {
		t.Fatalf("received = %d", received.Load())
	}

	channels, err := a.GetChannels(ctx, "u1")
	if err != nil || len(channels) != 1 {
		t.Fatalf("GetChannels: %v %v", channels, err)
	}

	before := time.Now().Add(time.Second)
	history, err := a.GetHistory(ctx, ch.ID, before, 10)
	if err != nil || len(history) != 1 {
		t.Fatalf("GetHistory: %v %v", history, err)
	}
}

func TestPresence(t *testing.T) {
	ctx := context.Background()
	a := memory.New()
	ch, err := a.CreateChannel(ctx, messaging.Direct, []string{"a", "b"}, "")
	if err != nil {
		t.Fatal(err)
	}
	if err := a.SetOnline(ctx, "a"); err != nil {
		t.Fatal(err)
	}
	online, err := a.IsOnline(ctx, "a")
	if err != nil || !online {
		t.Fatalf("IsOnline: %v %v", online, err)
	}
	users, err := a.GetOnlineUsers(ctx, ch.ID)
	if err != nil || len(users) != 1 || users[0] != "a" {
		t.Fatalf("GetOnlineUsers: %v %v", users, err)
	}
	if err := a.SetOffline(ctx, "a"); err != nil {
		t.Fatal(err)
	}
}

func TestErrors(t *testing.T) {
	ctx := context.Background()
	a := memory.New()
	if _, err := a.Subscribe("missing", func(*messaging.Message) {}); err == nil {
		t.Fatal("expected channel not found")
	}
	if err := a.SendMessage(ctx, "missing", &messaging.Message{}); err == nil {
		t.Fatal("expected channel not found")
	}
	if _, err := a.GetOnlineUsers(ctx, "missing"); err == nil {
		t.Fatal("expected channel not found")
	}
}
