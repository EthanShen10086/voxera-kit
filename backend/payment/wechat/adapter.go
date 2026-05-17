// Package wechat provides a WeChat Pay implementation of the payment.PaymentGateway interface.
// It is intended to use github.com/wechatpay-apiv3/wechatpay-go as the underlying SDK.
package wechat

import (
	"context"

	"github.com/EthanShen10086/voxera-kit/payment"
)

// Adapter implements the payment.PaymentGateway interface using WeChat Pay.
//
// Intended dependency: github.com/wechatpay-apiv3/wechatpay-go
type Adapter struct {
	// client *core.Client // TODO: uncomment when wechatpay-go dependency is added
	cfg payment.PaymentConfig
}

// New creates a new WeChat Pay Adapter with the provided configuration.
func New(cfg payment.PaymentConfig) *Adapter {
	return &Adapter{cfg: cfg}
}

// CreateOrder creates a WeChat Pay unified order.
func (a *Adapter) CreateOrder(ctx context.Context, order *payment.Order) (*payment.Order, error) {
	// TODO: implement using wechatpay-go
	return order, nil
}

// QueryOrder retrieves the current status of an order from WeChat Pay.
func (a *Adapter) QueryOrder(ctx context.Context, orderID string) (*payment.Order, error) {
	// TODO: implement using wechatpay-go
	return nil, nil
}

// Refund initiates a refund via the WeChat Pay refund API.
func (a *Adapter) Refund(ctx context.Context, req *payment.RefundRequest) error {
	// TODO: implement using wechatpay-go
	return nil
}

// HandleCallback processes a WeChat Pay asynchronous payment notification.
func (a *Adapter) HandleCallback(ctx context.Context, payload *payment.CallbackPayload) (*payment.Order, error) {
	// TODO: implement using wechatpay-go notification verification
	return nil, nil
}

// Close releases all resources held by the WeChat Pay client.
func (a *Adapter) Close() error {
	// TODO: implement using wechatpay-go
	return nil
}
