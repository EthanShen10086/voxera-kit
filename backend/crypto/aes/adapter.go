// Package aes provides an AES-GCM implementation of the crypto.Encryptor interface.
package aes

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
)

// Adapter implements crypto.Encryptor using AES-GCM authenticated encryption.
type Adapter struct {
	aead cipher.AEAD
}

// New creates a new AES-GCM encryptor. The key must be 16, 24, or 32 bytes
// for AES-128, AES-192, or AES-256 respectively.
func New(key []byte) (*Adapter, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return &Adapter{aead: aead}, nil
}

// Encrypt encrypts plaintext using AES-GCM with a random nonce.
// The returned ciphertext is nonce || encrypted_data.
func (a *Adapter) Encrypt(_ context.Context, plaintext []byte) ([]byte, error) {
	nonce := make([]byte, a.aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	return a.aead.Seal(nonce, nonce, plaintext, nil), nil
}

// Decrypt decrypts ciphertext that was produced by Encrypt.
func (a *Adapter) Decrypt(_ context.Context, ciphertext []byte) ([]byte, error) {
	nonceSize := a.aead.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}
	nonce, ct := ciphertext[:nonceSize], ciphertext[nonceSize:]
	return a.aead.Open(nil, nonce, ct, nil)
}
