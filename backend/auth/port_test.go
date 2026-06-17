package auth_test

import (
	"testing"

	"github.com/EthanShen10086/voxera-kit/auth"
	"github.com/EthanShen10086/voxera-kit/auth/jwt"
)

// Contract: JWT adapter must implement Authenticator (stub until TODOs land).
func TestJWTAdapterImplementsAuthenticator(t *testing.T) {
	var _ auth.Authenticator = jwt.New(auth.Config{Secret: "test-secret-key-32-chars-minimum"})
}

func TestPermissionStruct(t *testing.T) {
	p := auth.Permission{Resource: "models", Action: "read"}
	if p.Resource != "models" {
		t.Fatal("resource")
	}
}
