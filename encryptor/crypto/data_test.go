package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestData(t *testing.T) {
	const name = "aes256-gcm"
	s, err := MakeRand(SaltLength)
	assert.NoError(t, err)
	iv, err := MakeRand(8)
	assert.NoError(t, err)
	nonce, err := MakeRand(NonceLength)
	assert.NoError(t, err)
	bytes, err := MakeRand(512)
	assert.NoError(t, err)
	type testCase struct {
		cipher string
		key    KeyLen
		mode   Mode
		salt   []byte
		iv     []byte
		nonce  []byte
		bytes  []byte
		err    assert.ErrorAssertionFunc
	}
	tests := []testCase{
		{"aes", Key256, "gcm", nil, nil, nil, nil, assert.Error},
		{"aes", Key256, "gcm", s, nil, nonce, nil, assert.Error},
		{"aes", Key256, "gcm", s, iv, nil, nil, assert.Error},
		{"aes", Key256, "gcm", s, iv, nonce, nil, assert.Error},
		{"aes", Key256, "gcm", s, iv, nil, bytes, assert.NoError},
		{"aes", Key256, "gcm", s, nil, nonce, bytes, assert.NoError},
		{"aes", Key256, "gcm", s, iv, nonce, bytes, assert.NoError},
	}
	for _, test := range tests {
		data := NewData(Header{test.cipher, test.key, test.mode,
			test.salt, test.iv, test.nonce}, test.bytes)
		test.err(t, data.Valid())
	}
}

func makeHeader(t *testing.T) Header {
	s, err := MakeRand(SaltLength)
	assert.NoError(t, err)
	iv, err := MakeRand(NonceLength)
	assert.NoError(t, err)
	nonce, err := MakeRand(NonceLength)
	assert.NoError(t, err)
	h, err := NewHeader("aes", Key256, "gcm", s, iv, nonce)
	assert.NoError(t, err)
	return h
}

func TestEncodeData(t *testing.T) {
	bytes, err := MakeRand(512)
	assert.NoError(t, err)
	data := NewData(makeHeader(t), bytes)
	assert.NoError(t, data.Valid())
	enc, err := EncodeData(data)
	assert.NoError(t, err)
	assert.NotEmpty(t, enc)
	dec, err := DecodeData(enc)
	assert.NoError(t, err)
	assert.Equal(t, data, dec)
}

func TestGobEncodeData(t *testing.T) {
	bytes, err := MakeRand(512)
	assert.NoError(t, err)
	data := NewData(makeHeader(t), bytes)
	assert.NoError(t, data.Valid())
	enc, err := GobEncodeData(data)
	assert.NoError(t, err)
	assert.NotEmpty(t, enc)
	dec, err := GobDecodeData(enc)
	assert.NoError(t, err)
	assert.Equal(t, data, dec)
}
