package oauth2_test

import (
	"context"
	"testing"

	"github.com/EthanShen10086/voxera-kit/auth"
	"github.com/EthanShen10086/voxera-kit/auth/oauth2"
)

func TestOAuth2StubMethods(t *testing.T) {
	a := oauth2.New(auth.Config{Issuer: "https://auth.example"})
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
