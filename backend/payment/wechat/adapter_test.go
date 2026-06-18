package wechat_test

import (
	"context"
	"testing"

	"github.com/EthanShen10086/voxera-kit/payment"
	"github.com/EthanShen10086/voxera-kit/payment/wechat"
)

func TestWechatGatewayStub(t *testing.T) {
	a := wechat.New(payment.Config{APIKey: "mch"})
	order, err := a.CreateOrder(context.Background(), &payment.Order{ID: "w1"})
	if err != nil || order.ID != "w1" {
		t.Fatalf("CreateOrder: %+v err=%v", order, err)
	}
	if err := a.Refund(context.Background(), &payment.RefundRequest{OrderID: "w1"}); err != nil {
		t.Fatal(err)
	}
	got, err := a.QueryOrder(context.Background(), "w1")
	if err != nil || got != nil {
		t.Fatalf("QueryOrder: %+v err=%v", got, err)
	}
	if err := a.Close(); err != nil {
		t.Fatal(err)
	}
}
