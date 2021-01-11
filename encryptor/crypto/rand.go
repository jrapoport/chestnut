package crypto

import (
	"crypto/rand"
	"fmt"
	"io"
)

const (
	// SaltLength is the default salt length.
	SaltLength = 32

	// NonceLength is the default nonce length.
	NonceLength = 12
)

// MakeRand returns a buffer of size length filled with random bytes.
func MakeRand(length uint) ([]byte, error) {
	// generate random bytes
	r := make([]byte, length)
	n, err := io.ReadFull(rand.Reader, r)
	if err != nil {
		return nil, err
	} else if uint(n) != length {
		return nil, fmt.Errorf("invalid buffer length %d != %d", n, length)
	}
	return r, nil
}

// MakeSalt returns random salt of size length.
func MakeSalt() ([]byte, error) {
	return MakeRand(SaltLength)
}

// MakeNonce returns random nonce of size length.
func MakeNonce() ([]byte, error) {
	return MakeRand(NonceLength)
}
