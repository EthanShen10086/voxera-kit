// Package paypal provides a PayPal implementation of the payment.PaymentGateway interface.
// It is intended to use github.com/plutov/paypal/v4 as the underlying SDK.
package paypal

import (
	"context"

	"github.com/EthanShen10086/voxera-kit/payment"
)

// Adapter implements the payment.PaymentGateway interface using PayPal.
//
// Intended dependency: github.com/plutov/paypal/v4
type Adapter struct {
	// client *paypal.Client // TODO: uncomment when paypal SDK dependency is added
	cfg payment.PaymentConfig
}

// New creates a new PayPal Adapter with the provided configuration.
func New(cfg payment.PaymentConfig) *Adapter {
	return &Adapter{cfg: cfg}
}

// CreateOrder creates a PayPal order.
func (a *Adapter) CreateOrder(ctx context.Context, order *payment.Order) (*payment.Order, error) {
	// TODO: implement using paypal SDK
	return order, nil
}

// QueryOrder retrieves the current status of a PayPal order.
func (a *Adapter) QueryOrder(ctx context.Context, orderID string) (*payment.Order, error) {
	// TODO: implement using paypal SDK
	return nil, nil
}

// Refund initiates a refund for a captured PayPal payment.
func (a *Adapter) Refund(ctx context.Context, req *payment.RefundRequest) error {
	// TODO: implement using paypal SDK
	return nil
}

// HandleCallback processes a PayPal webhook notification.
func (a *Adapter) HandleCallback(ctx context.Context, payload *payment.CallbackPayload) (*payment.Order, error) {
	// TODO: implement using paypal SDK webhook verification
	return nil, nil
}

// Close releases all resources held by the PayPal client.
func (a *Adapter) Close() error {
	// TODO: implement using paypal SDK
	return nil
}
