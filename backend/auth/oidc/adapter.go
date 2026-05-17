// Package oidc provides an OpenID Connect implementation of [auth.Authenticator].
//
// This adapter extends OAuth 2.0 with ID token verification and
// UserInfo endpoint integration for identity-aware authentication.
package oidc

import (
	"context"

	"github.com/EthanShen10086/voxera-kit/auth"
)

// Adapter implements [auth.Authenticator] using OpenID Connect discovery and
// ID token validation.
type Adapter struct {
	cfg auth.AuthConfig
}

// New creates an OIDC [Adapter] with the supplied configuration.
func New(cfg auth.AuthConfig) *Adapter {
	return &Adapter{cfg: cfg}
}

func (a *Adapter) Authenticate(ctx context.Context, token string) (*auth.Claims, error) {
	// TODO: validate ID token using OIDC discovery and JWKS
	return nil, nil
}

func (a *Adapter) GenerateToken(ctx context.Context, claims *auth.Claims) (*auth.TokenPair, error) {
	// TODO: initiate OIDC authorization code flow
	return nil, nil
}

func (a *Adapter) RefreshToken(ctx context.Context, refreshToken string) (*auth.TokenPair, error) {
	// TODO: refresh tokens via OIDC token endpoint
	return nil, nil
}

func (a *Adapter) RevokeToken(ctx context.Context, token string) error {
	// TODO: revoke token at OIDC provider
	return nil
}

var _ auth.Authenticator = (*Adapter)(nil)
