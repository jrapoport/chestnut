package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	secret     = "i-am-a-secret"
	saltLen    = 8
	keyLen     = 32
	iterations = 1024
)

func TestNewCipherKeys(t *testing.T) {
	salt, err := MakeRand(saltLen)
	assert.NoError(t, err)
	assert.Len(t, salt, saltLen)
	sec := []byte(secret)
	cipher := func() ([]byte, error) { return NewCipherKey(keyLen, sec, salt) }
	pbkdf2 := func() ([]byte, error) { return NewPBKDF2CipherKey(keyLen, iterations, sec, salt) }
	scrypt := func() ([]byte, error) { return NewScryptCipherKey(keyLen, iterations, sec, salt) }
	test := func(newKey func() ([]byte, error)) {
		key1, err1 := newKey()
		key2, err2 := newKey()
		assert.NoError(t, err1)
		assert.NoError(t, err2)
		assert.Equal(t, key1, key2)
	}
	t.Run("NewCipherKey", func(t *testing.T) { test(cipher) })
	t.Run("NewPBKDF2CipherKey", func(t *testing.T) { test(pbkdf2) })
	t.Run("NewScryptCipherKey", func(t *testing.T) { test(scrypt) })
}
