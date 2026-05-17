// Package auth defines the ports (interfaces) and domain types for
// authentication, token management, and role-based authorization.
package auth

import (
	"context"
	"time"
)

// TokenPair holds an access/refresh token pair returned after successful authentication.
type TokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
	TokenType    string
}

// Claims represents the decoded payload of an authentication token.
type Claims struct {
	UserID      string
	Roles       []string
	Permissions []string
	IssuedAt    time.Time
	ExpiresAt   time.Time
	Metadata    map[string]any
}

// AuthConfig carries the settings required to issue and validate tokens.
type AuthConfig struct {
	// Secret is the signing key or secret used for token generation.
	Secret string

	// Issuer identifies the token issuer (iss claim).
	Issuer string

	// AccessTokenTTL controls the lifetime of access tokens.
	AccessTokenTTL time.Duration

	// RefreshTokenTTL controls the lifetime of refresh tokens.
	RefreshTokenTTL time.Duration

	// Algorithm specifies the signing algorithm (e.g. "HS256", "RS256").
	Algorithm string
}

// Authenticator is the primary port for token-based authentication.
type Authenticator interface {
	// Authenticate validates a raw token string and returns the embedded claims.
	Authenticate(ctx context.Context, token string) (*Claims, error)

	// GenerateToken issues a new token pair for the given claims.
	GenerateToken(ctx context.Context, claims *Claims) (*TokenPair, error)

	// RefreshToken exchanges a valid refresh token for a fresh token pair.
	RefreshToken(ctx context.Context, refreshToken string) (*TokenPair, error)

	// RevokeToken invalidates a previously issued token.
	RevokeToken(ctx context.Context, token string) error
}

// Permission pairs a resource with an action for fine-grained access control.
type Permission struct {
	Resource string
	Action   string
}

// Authorizer is the port for role- and permission-based authorization decisions.
type Authorizer interface {
	// Authorize checks whether the claims holder may perform action on resource.
	Authorize(ctx context.Context, claims *Claims, resource, action string) (bool, error)

	// HasPermission checks a structured Permission value.
	HasPermission(ctx context.Context, claims *Claims, perm Permission) (bool, error)

	// HasRole returns true when the claims holder possesses the given role.
	HasRole(ctx context.Context, claims *Claims, role string) bool
}
