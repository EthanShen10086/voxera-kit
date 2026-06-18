package oidc_test

import (
	"context"
	"testing"

	"github.com/EthanShen10086/voxera-kit/auth"
	"github.com/EthanShen10086/voxera-kit/auth/oidc"
)

func TestOIDCStubMethods(t *testing.T) {
	a := oidc.New(auth.Config{Issuer: "https://issuer"})
	ctx := context.Background()
	if _, err := a.Authenticate(ctx, "token"); err != nil {
		t.Fatal(err)
	}
	if _, err := a.GenerateToken(ctx, &auth.Claims{UserID: "u"}); err != nil {
		t.Fatal(err)
	}
	if _, err := a.RefreshToken(ctx, "refresh"); err != nil {
		t.Fatal(err)
	}
	if err := a.RevokeToken(ctx, "token"); err != nil {
		t.Fatal(err)
	}
}
