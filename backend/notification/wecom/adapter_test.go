package wecom_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/EthanShen10086/voxera-kit/notification"
	"github.com/EthanShen10086/voxera-kit/notification/wecom"
)

func TestSendSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	a := wecom.New(notification.Config{WebhookURL: srv.URL})
	res, err := a.Send(context.Background(), &notification.Message{Title: "t", Content: "c"})
	if err != nil || res.Status != notification.StatusDelivered {
		t.Fatalf("Send: %+v err=%v", res, err)
	}
	if a.Channel() != notification.ChannelWecom {
		t.Fatalf("Channel = %v", a.Channel())
	}
}

func TestSendHTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer srv.Close()

	a := wecom.New(notification.Config{WebhookURL: srv.URL})
	res, err := a.Send(context.Background(), &notification.Message{Content: "c"})
	if err != nil || res.Status != notification.StatusFailed {
		t.Fatalf("Send: %+v err=%v", res, err)
	}
}
