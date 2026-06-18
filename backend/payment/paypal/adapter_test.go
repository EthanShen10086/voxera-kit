package paypal_test

import (
	"context"
	"testing"

	"github.com/EthanShen10086/voxera-kit/payment"
	"github.com/EthanShen10086/voxera-kit/payment/paypal"
)

func TestPayPalGatewayStub(t *testing.T) {
	a := paypal.New(payment.Config{APIKey: "client"})
	order, err := a.CreateOrder(context.Background(), &payment.Order{ID: "p1"})
	if err != nil || order.ID != "p1" {
		t.Fatalf("CreateOrder: %+v err=%v", order, err)
	}
	if err := a.Refund(context.Background(), &payment.RefundRequest{OrderID: "p1"}); err != nil {
		t.Fatal(err)
	}
	got, err := a.QueryOrder(context.Background(), "p1")
	if err != nil || got != nil {
		t.Fatalf("QueryOrder: %+v err=%v", got, err)
	}
	if err := a.Close(); err != nil {
		t.Fatal(err)
	}
}
