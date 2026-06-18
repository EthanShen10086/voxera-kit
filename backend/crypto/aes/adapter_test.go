package aes_test

import (
	"context"
	"testing"

	"github.com/EthanShen10086/voxera-kit/crypto/aes"
)

func TestEncryptDecryptRoundtrip(t *testing.T) {
	a, err := aes.New(make([]byte, 32))
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	plain := []byte("secret payload")
	enc, err := a.Encrypt(ctx, plain)
	if err != nil {
		t.Fatal(err)
	}
	out, err := a.Decrypt(ctx, enc)
	if err != nil || string(out) != string(plain) {
		t.Fatalf("Decrypt: %q err=%v", out, err)
	}
}

func TestNewInvalidKey(t *testing.T) {
	_, err := aes.New([]byte("short"))
	if err == nil {
		t.Fatal("expected key size error")
	}
}

func TestDecryptTooShort(t *testing.T) {
	a, _ := aes.New(make([]byte, 16))
	_, err := a.Decrypt(context.Background(), []byte{1, 2, 3})
	if err == nil {
		t.Fatal("expected ciphertext too short")
	}
}
