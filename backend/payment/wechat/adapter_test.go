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
	if err := a.Close(); err != nil {
		t.Fatal(err)
	}
}
