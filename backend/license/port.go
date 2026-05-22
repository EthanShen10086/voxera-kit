// Package license provides deployment licensing validation and management,
// supporting both offline signature verification and online license servers.
package license

import (
	"context"
	"errors"
	"time"
)

// Type represents the license tier.
type Type string

const (
	// TypeTrial represents a time-limited trial license.
	TypeTrial Type = "trial"
	// TypeStandard represents a standard production license.
	TypeStandard Type = "standard"
	// TypeEnterprise represents an enterprise-grade license with extended features.
	TypeEnterprise Type = "enterprise"
)

// License represents a deployment authorization.
type License struct {
	ID        string
	TenantID  string
	Type      Type
	Features  []string
	MaxUsers  int
	IssuedAt  time.Time
	ExpiresAt time.Time
	Signature string // RSA signature for tamper detection
}

// Manager validates and manages deployment licenses.
type Manager interface {
	Validate(ctx context.Context, key string) (*License, error)
	Refresh(ctx context.Context, key string) (*License, error)
	Features(ctx context.Context, key string) ([]string, error)
	IsExpired(ctx context.Context, key string) (bool, error)
}

// ErrInvalidLicense indicates the license key is malformed or tampered.
var ErrInvalidLicense = errors.New("license: invalid or tampered license key")

// ErrExpired indicates the license has passed its expiration date.
var ErrExpired = errors.New("license: expired")

// ErrFeatureNotLicensed indicates the requested feature is not included.
var ErrFeatureNotLicensed = errors.New("license: feature not licensed")
