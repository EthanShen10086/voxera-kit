package notification

import (
	"context"
	"errors"
	"testing"
)

type stubNotifier struct {
	channel ChannelType
	err     error
	result  *DeliveryResult
}

func (s *stubNotifier) Channel() ChannelType { return s.channel }
func (s *stubNotifier) Send(_ context.Context, _ *Message) (*DeliveryResult, error) {
	if s.err != nil {
		return nil, s.err
	}
	if s.result != nil {
		return s.result, nil
	}
	return &DeliveryResult{Status: StatusDelivered, Channel: s.channel}, nil
}

func TestDefaultRouterSend(t *testing.T) {
	r := NewRouter()
	r.Register(&stubNotifier{channel: ChannelEmail, result: &DeliveryResult{Status: StatusDelivered, Channel: ChannelEmail}})

	res, err := r.Send(context.Background(), ChannelEmail, &Message{Title: "hi"})
	if err != nil || res == nil || res.Status != StatusDelivered {
		t.Fatalf("Send() = %#v, %v", res, err)
	}

	_, err = r.Send(context.Background(), ChannelSlack, &Message{})
	if err == nil {
		t.Fatal("expected unregistered channel error")
	}
}

func TestDefaultRouterRegisterOverwrite(t *testing.T) {
	r := NewRouter()
	r.Register(&stubNotifier{channel: ChannelEmail})
	r.Register(&stubNotifier{channel: ChannelEmail, result: &DeliveryResult{Status: StatusDelivered, Channel: ChannelEmail}})
	res, err := r.Send(context.Background(), ChannelEmail, &Message{})
	if err != nil || res.Status != StatusDelivered {
		t.Fatalf("Send() = %#v, %v", res, err)
	}
}

func TestDefaultRouterSendAll(t *testing.T) {
	r := NewRouter()
	r.Register(&stubNotifier{channel: ChannelEmail})
	r.Register(&stubNotifier{channel: ChannelSlack})

	results, err := r.SendAll(context.Background(), &Message{Content: "all"})
	if err != nil || len(results) != 2 {
		t.Fatalf("SendAll() = %#v, %v", results, err)
	}

	r2 := NewRouter()
	r2.Register(&stubNotifier{channel: ChannelEmail, err: errors.New("fail")})
	_, err = r2.SendAll(context.Background(), &Message{})
	if err == nil {
		t.Fatal("expected error from SendAll")
	}
}
