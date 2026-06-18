package vault

import (
	"errors"
	"testing"

	vaultapi "github.com/hashicorp/vault/api"
)

func TestIsNotFound(t *testing.T) {
	if isNotFound(nil) {
		t.Fatal("nil should not be not found")
	}
	if !isNotFound(vaultapi.ErrSecretNotFound) {
		t.Fatal("expected secret not found")
	}
	if !isNotFound(errors.New("404 page not found")) {
		t.Fatal("expected message match")
	}
}

func TestNewManagerDefaultMount(t *testing.T) {
	mgr, err := NewManager(Config{Address: "http://127.0.0.1:8200", Token: "t"})
	if err != nil {
		t.Fatal(err)
	}
	if mgr.mount != "secret" {
		t.Fatalf("mount = %q", mgr.mount)
	}
}
