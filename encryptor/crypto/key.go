package crypto

import (
	"crypto/sha512"

	"golang.org/x/crypto/pbkdf2"
	"golang.org/x/crypto/scrypt"
)

// KeyLen is used to select 128, 192, or 256 bit keys.
type KeyLen int

// key lengths
const (
	Key128 KeyLen = 16 // 128 bit
	Key192        = 24 // 128 bit
	Key256        = 32 // 128 bit
)

func (k KeyLen) String() string {
	switch k {
	case Key128:
		return "128"
	case Key192:
		return "192"
	case Key256:
		return "256"
	default:
		return ""
	}
}

// NewCipherKey generate a new cipher key of the appropriate key length.
// Note: Currently this is hard-coded to 4096 key iterations. The thinking here is that
// the strength of secret was determined externally and therefore it less important to
// iterate (again) a large number of times. 1<<15 (or 32768) key iterations, seems to
// be the current consensus for passwords in general (2020).
func NewCipherKey(l KeyLen, secret, salt []byte) ([]byte, error) {
	const keyIterations = 4096
	return NewScryptCipherKey(l, keyIterations, secret, salt)
}

// NewPBKDF2CipherKey generate a new cipher key using pbkdf2.
func NewPBKDF2CipherKey(l KeyLen, iterations int, secret, salt []byte) ([]byte, error) {
	// sha512, in addition to being more secure, should be faster on 64-bit systems
	return pbkdf2.Key(secret, salt, iterations, int(l), sha512.New), nil
}

// NewScryptCipherKey generate a new cipher key using scrypt.
func NewScryptCipherKey(l KeyLen, iterations int, secret, salt []byte) ([]byte, error) {
	return scrypt.Key(secret, salt, iterations, 8, 1, int(l))
}
