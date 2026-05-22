// Package stripe provides a Stripe implementation of the payment.Gateway interface.
// It is intended to use github.com/stripe/stripe-go/v78 as the underlying SDK.
package stripe

import (
	"context"

	"github.com/EthanShen10086/voxera-kit/payment"
)

// Adapter implements the payment.Gateway interface using Stripe.
//
// Intended dependency: github.com/stripe/stripe-go/v78
type Adapter struct {
	// No external client field; Stripe SDK uses package-level configuration.
	cfg payment.Config
}

// New creates a new Stripe Adapter with the provided configuration.
func New(cfg payment.Config) *Adapter {
	return &Adapter{cfg: cfg}
}

// CreateOrder creates a Stripe PaymentIntent for the given order.
func (a *Adapter) CreateOrder(ctx context.Context, order *payment.Order) (*payment.Order, error) {
	// TODO: implement using stripe-go
	return order, nil
}

// QueryOrder retrieves the current status of a PaymentIntent from Stripe.
func (a *Adapter) QueryOrder(ctx context.Context, orderID string) (*payment.Order, error) {
	// TODO: implement using stripe-go
	return nil, nil
}

// Refund initiates a refund via the Stripe Refunds API.
func (a *Adapter) Refund(ctx context.Context, req *payment.RefundRequest) error {
	// TODO: implement using stripe-go
	return nil
}

// HandleCallback processes a Stripe webhook event.
func (a *Adapter) HandleCallback(ctx context.Context, payload *payment.CallbackPayload) (*payment.Order, error) {
	// TODO: implement using stripe-go webhook verification
	return nil, nil
}

// Close releases all resources (no-op for Stripe as it uses stateless HTTP).
func (a *Adapter) Close() error {
	return nil
}
