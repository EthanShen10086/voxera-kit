package bcrypt_test

import (
	"testing"

	"github.com/EthanShen10086/voxera-kit/crypto/bcrypt"
)

func TestBcryptStub(t *testing.T) {
	a := bcrypt.New()
	hash, err := a.Hash([]byte("secret"))
	if err != nil || hash != "" {
		t.Fatalf("Hash: %q err=%v", hash, err)
	}
	ok, err := a.Verify([]byte("secret"), "")
	if err != nil || ok {
		t.Fatalf("Verify: ok=%v err=%v", ok, err)
	}
}
