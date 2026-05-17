// Package bcrypt provides a bcrypt implementation of the crypto.Hasher interface.
//
// Intended dependency: golang.org/x/crypto/bcrypt
package bcrypt

// Adapter implements crypto.Hasher using the bcrypt adaptive hashing function.
type Adapter struct {
	// cost int // TODO: uncomment when golang.org/x/crypto dependency is added
}

// New creates a new bcrypt hasher with the default cost.
func New() *Adapter {
	return &Adapter{}
}

// Hash computes a bcrypt hash of the given data.
func (a *Adapter) Hash(data []byte) (string, error) {
	// TODO: implement using golang.org/x/crypto/bcrypt
	return "", nil
}

// Verify reports whether data matches the given bcrypt hash.
func (a *Adapter) Verify(data []byte, hash string) (bool, error) {
	// TODO: implement using golang.org/x/crypto/bcrypt
	return false, nil
}
