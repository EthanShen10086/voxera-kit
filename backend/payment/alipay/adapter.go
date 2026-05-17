// Package alipay provides an Alipay implementation of the payment.PaymentGateway interface.
// It is intended to use github.com/smartwalle/alipay/v3 as the underlying SDK.
package alipay

import (
	"context"

	"github.com/EthanShen10086/voxera-kit/payment"
)

// Adapter implements the payment.PaymentGateway interface using Alipay.
//
// Intended dependency: github.com/smartwalle/alipay/v3
type Adapter struct {
	// client *alipay.Client // TODO: uncomment when alipay SDK dependency is added
	cfg payment.PaymentConfig
}

// New creates a new Alipay Adapter with the provided configuration.
func New(cfg payment.PaymentConfig) *Adapter {
	return &Adapter{cfg: cfg}
}

// CreateOrder creates an Alipay trade order.
func (a *Adapter) CreateOrder(ctx context.Context, order *payment.Order) (*payment.Order, error) {
	// TODO: implement using alipay SDK
	return order, nil
}

// QueryOrder retrieves the current status of a trade from Alipay.
func (a *Adapter) QueryOrder(ctx context.Context, orderID string) (*payment.Order, error) {
	// TODO: implement using alipay SDK
	return nil, nil
}

// Refund initiates a refund via the Alipay refund API.
func (a *Adapter) Refund(ctx context.Context, req *payment.RefundRequest) error {
	// TODO: implement using alipay SDK
	return nil
}

// HandleCallback processes an Alipay asynchronous payment notification.
func (a *Adapter) HandleCallback(ctx context.Context, payload *payment.CallbackPayload) (*payment.Order, error) {
	// TODO: implement using alipay SDK notification verification
	return nil, nil
}

// Close releases all resources held by the Alipay client.
func (a *Adapter) Close() error {
	// TODO: implement using alipay SDK
	return nil
}
