package stripe_test

import (
	"context"
	"testing"

	"github.com/EthanShen10086/voxera-kit/payment"
	"github.com/EthanShen10086/voxera-kit/payment/stripe"
)

func TestStripeGatewayStub(t *testing.T) {
	a := stripe.New(payment.Config{APIKey: "sk_test"})
	ctx := context.Background()
	order := &payment.Order{ID: "o1"}
	got, err := a.CreateOrder(ctx, order)
	if err != nil || got.ID != "o1" {
		t.Fatalf("CreateOrder: %+v err=%v", got, err)
	}
	if err := a.Refund(ctx, &payment.RefundRequest{OrderID: "o1"}); err != nil {
		t.Fatal(err)
	}
	q, err := a.QueryOrder(ctx, "o1")
	if err != nil || q != nil {
		t.Fatalf("QueryOrder: %+v err=%v", q, err)
	}
	cb, err := a.HandleCallback(ctx, &payment.CallbackPayload{Raw: []byte("{}")})
	if err != nil || cb != nil {
		t.Fatalf("HandleCallback: %+v err=%v", cb, err)
	}
	if err := a.Close(); err != nil {
		t.Fatal(err)
	}
}
