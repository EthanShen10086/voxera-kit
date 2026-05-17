// Package crypto defines the port interfaces for encryption, hashing, signing,
// and key generation. It abstracts away specific cryptographic algorithms so that
// implementations can be swapped without changing application code.
package crypto

import "context"

// EncryptionAlgorithm identifies a symmetric encryption scheme.
type EncryptionAlgorithm int

const (
	// AES_GCM uses AES in Galois/Counter Mode (authenticated encryption).
	AES_GCM EncryptionAlgorithm = iota
	// AES_CBC uses AES in Cipher Block Chaining mode.
	AES_CBC
	// ChaCha20Poly1305 uses the ChaCha20-Poly1305 AEAD construction.
	ChaCha20Poly1305
)

// HashAlgorithm identifies a password-hashing or digest algorithm.
type HashAlgorithm int

const (
	// Bcrypt uses the bcrypt adaptive hashing function.
	Bcrypt HashAlgorithm = iota
	// Argon2 uses the Argon2id key derivation function.
	Argon2
	// SHA256 uses the SHA-256 cryptographic hash.
	SHA256
	// SHA512 uses the SHA-512 cryptographic hash.
	SHA512
)

// Encryptor performs symmetric encryption and decryption.
type Encryptor interface {
	// Encrypt encrypts plaintext and returns the ciphertext.
	Encrypt(ctx context.Context, plaintext []byte) ([]byte, error)
	// Decrypt decrypts ciphertext and returns the original plaintext.
	Decrypt(ctx context.Context, ciphertext []byte) ([]byte, error)
}

// Hasher produces and verifies cryptographic hashes.
type Hasher interface {
	// Hash computes a hash of the given data and returns its string representation.
	Hash(data []byte) (string, error)
	// Verify reports whether data matches the given hash.
	Verify(data []byte, hash string) (bool, error)
}

// Signer produces and verifies digital signatures.
type Signer interface {
	// Sign computes a signature over the given payload.
	Sign(payload []byte) ([]byte, error)
	// Verify reports whether the signature is valid for the given payload.
	Verify(payload, signature []byte) (bool, error)
}

// KeyGenerator creates cryptographic keys and key pairs.
type KeyGenerator interface {
	// GenerateKey creates a symmetric key of the specified bit length.
	GenerateKey(bits int) ([]byte, error)
	// GenerateKeyPair creates an asymmetric key pair of the specified bit length.
	// Returns the public key, private key, and any error.
	GenerateKeyPair(bits int) ([]byte, []byte, error)
}
