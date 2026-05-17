// Package oauth2 provides an OAuth 2.0-based implementation of [auth.Authenticator].
//
// This adapter handles the OAuth 2.0 authorization code and client credentials
// flows, delegating token validation to the authorization server's introspection endpoint.
package oauth2

import (
	"context"

	"github.com/EthanShen10086/voxera-kit/auth"
)

// Adapter implements [auth.Authenticator] via OAuth 2.0 token introspection.
type Adapter struct {
	cfg auth.AuthConfig
}

// New creates an OAuth 2.0 [Adapter] with the supplied configuration.
func New(cfg auth.AuthConfig) *Adapter {
	return &Adapter{cfg: cfg}
}

func (a *Adapter) Authenticate(ctx context.Context, token string) (*auth.Claims, error) {
	// TODO: introspect token at authorization server
	return nil, nil
}

func (a *Adapter) GenerateToken(ctx context.Context, claims *auth.Claims) (*auth.TokenPair, error) {
	// TODO: exchange credentials for token via OAuth 2.0 flow
	return nil, nil
}

func (a *Adapter) RefreshToken(ctx context.Context, refreshToken string) (*auth.TokenPair, error) {
	// TODO: use refresh_token grant to obtain new tokens
	return nil, nil
}

func (a *Adapter) RevokeToken(ctx context.Context, token string) error {
	// TODO: call revocation endpoint
	return nil
}

var _ auth.Authenticator = (*Adapter)(nil)
