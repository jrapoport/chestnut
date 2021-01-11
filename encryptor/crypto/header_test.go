package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHeader(t *testing.T) {
	const name = "aes256-gcm"
	s, err := MakeRand(SaltLength)
	assert.NoError(t, err)
	iv, err := MakeRand(8)
	assert.NoError(t, err)
	nonce, err := MakeRand(NonceLength)
	assert.NoError(t, err)
	type testCase struct {
		cipher string
		key    KeyLen
		mode   Mode
		salt   []byte
		iv     []byte
		nonce  []byte
		name   string
		err    assert.ErrorAssertionFunc
	}
	tests := []testCase{
		{"", 0, "", nil, nil, nil, "", assert.Error},
		{"aes", 0, "", nil, nil, nil, "", assert.Error},
		{"aes", Key256, "", nil, nil, nil, "", assert.Error},
		{"aes", Key256, "gcm", nil, nil, nil, "", assert.Error},
		{"aes", Key256, "gcm", []byte(""), nil, nil, "", assert.Error},
		{"aes", Key256, "gcm", s, nil, []byte(""), "", assert.Error},
		{"aes", Key256, "gcm", s, nil, nil, name, assert.NoError},
		{"aes", Key256, "gcm", s, iv, nil, name, assert.NoError},
		{"aes", Key256, "gcm", s, nil, nonce, name, assert.NoError},
		{"aes", Key256, "gcm", s, iv, nonce, name, assert.NoError},
		{"AES", Key256, "GCM", s, iv, nonce, name, assert.NoError},
	}
	for _, test := range tests {
		h, err := NewHeader(test.cipher, test.key, test.mode, test.salt, test.iv, test.nonce)
		test.err(t, err)
		if err == nil {
			assert.NotNil(t, h)
			assert.Equal(t, test.name, h.Name())
		}
	}
}
