// Package jwt provides a JSON Web Token implementation of [auth.Authenticator].
//
// It relies on a third-party JWT library for token signing and verification.
package jwt

import (
	"context"

	"github.com/EthanShen10086/voxera-kit/auth"
)

// Adapter implements [auth.Authenticator] using signed JWTs.
type Adapter struct {
	cfg auth.Config
}

// New creates a JWT [Adapter] with the supplied configuration.
func New(cfg auth.Config) *Adapter {
	return &Adapter{cfg: cfg}
}

// Authenticate parses and validates a JWT, returning the embedded claims.
func (a *Adapter) Authenticate(ctx context.Context, token string) (*auth.Claims, error) {
	// TODO: parse and validate JWT, return claims
	return nil, nil
}

// GenerateToken signs a new JWT pair from the provided claims.
func (a *Adapter) GenerateToken(ctx context.Context, claims *auth.Claims) (*auth.TokenPair, error) {
	// TODO: sign a new JWT pair from the provided claims
	return nil, nil
}

// RefreshToken validates a refresh token and issues a fresh pair.
func (a *Adapter) RefreshToken(ctx context.Context, refreshToken string) (*auth.TokenPair, error) {
	// TODO: validate refresh token and issue a fresh pair
	return nil, nil
}

// RevokeToken adds a token to the revocation list.
func (a *Adapter) RevokeToken(ctx context.Context, token string) error {
	// TODO: add token to revocation list / blacklist
	return nil
}

var _ auth.Authenticator = (*Adapter)(nil)
