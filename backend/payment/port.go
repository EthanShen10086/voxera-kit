// Package payment defines the port interface for payment gateway operations.
// It abstracts order creation, querying, refunds, and callback handling
// across different payment providers (Stripe, WeChat Pay, Alipay, PayPal).
package payment

import (
	"context"
	"time"
)

// OrderStatus represents the current state of a payment order.
type OrderStatus string

const (
	// OrderStatusPending indicates the order is awaiting payment.
	OrderStatusPending OrderStatus = "pending"
	// OrderStatusPaid indicates the order has been successfully paid.
	OrderStatusPaid OrderStatus = "paid"
	// OrderStatusRefunded indicates the order has been refunded.
	OrderStatusRefunded OrderStatus = "refunded"
	// OrderStatusFailed indicates the payment attempt failed.
	OrderStatusFailed OrderStatus = "failed"
	// OrderStatusCancelled indicates the order was cancelled before payment.
	OrderStatusCancelled OrderStatus = "cancelled"
)

// Order represents a payment order with its full lifecycle metadata.
type Order struct {
	// ID is the unique order identifier.
	ID string
	// Amount is the payment amount in the smallest currency unit (e.g., cents).
	Amount int64
	// Currency is the ISO 4217 currency code (e.g., "USD", "CNY").
	Currency string
	// Status is the current order status.
	Status OrderStatus
	// Metadata contains arbitrary key-value data associated with the order.
	Metadata map[string]string
	// CreatedAt is the order creation timestamp.
	CreatedAt time.Time
	// PaidAt is the timestamp when payment was confirmed (zero if unpaid).
	PaidAt time.Time
	// ExpiresAt is the deadline for completing payment.
	ExpiresAt time.Time
}

// RefundRequest contains the parameters needed to initiate a refund.
type RefundRequest struct {
	// OrderID is the identifier of the order to refund.
	OrderID string
	// Amount is the refund amount in the smallest currency unit.
	// Use 0 or the full order amount for a full refund.
	Amount int64
	// Reason is a human-readable explanation for the refund.
	Reason string
}

// CallbackPayload represents the raw webhook/callback data from the payment provider.
type CallbackPayload struct {
	// Raw is the unprocessed request body bytes.
	Raw []byte
	// Headers contains the HTTP headers from the callback request.
	Headers map[string]string
}

// PaymentGateway is the interface for interacting with payment providers.
type PaymentGateway interface {
	// CreateOrder initiates a new payment order with the provider.
	CreateOrder(ctx context.Context, order *Order) (*Order, error)
	// QueryOrder retrieves the current status of an order from the provider.
	QueryOrder(ctx context.Context, orderID string) (*Order, error)
	// Refund initiates a refund for a previously paid order.
	Refund(ctx context.Context, req *RefundRequest) error
	// HandleCallback processes an asynchronous payment notification from the provider.
	HandleCallback(ctx context.Context, payload *CallbackPayload) (*Order, error)
	// Close releases all resources held by the gateway client.
	Close() error
}

// PaymentConfig holds the configuration parameters for a payment gateway.
type PaymentConfig struct {
	// AppID is the application identifier assigned by the payment provider.
	AppID string
	// MerchantID is the merchant account identifier.
	MerchantID string
	// APIKey is the API key or secret for authentication.
	APIKey string
	// APISecret is the secondary secret used for signing or verification.
	APISecret string
	// NotifyURL is the webhook URL for receiving payment notifications.
	NotifyURL string
	// ReturnURL is the URL to redirect users after payment completion.
	ReturnURL string
	// Sandbox enables sandbox/test mode when true.
	Sandbox bool
}
