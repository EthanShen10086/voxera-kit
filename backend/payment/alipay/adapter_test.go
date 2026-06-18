package alipay_test

import (
	"context"
	"testing"

	"github.com/EthanShen10086/voxera-kit/payment"
	"github.com/EthanShen10086/voxera-kit/payment/alipay"
)

func TestAlipayGatewayStub(t *testing.T) {
	a := alipay.New(payment.Config{APIKey: "app"})
	order, err := a.CreateOrder(context.Background(), &payment.Order{ID: "a1"})
	if err != nil || order.ID != "a1" {
		t.Fatalf("CreateOrder: %+v err=%v", order, err)
	}
	if err := a.Close(); err != nil {
		t.Fatal(err)
	}
}
