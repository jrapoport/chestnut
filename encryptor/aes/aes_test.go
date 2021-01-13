package aes

import (
	"math"
	"testing"

	"github.com/jrapoport/chestnut/encryptor/crypto"
	"github.com/stretchr/testify/assert"
)

func TestAllCiphers(t *testing.T) {
	ciphers := []struct {
		name        crypto.Mode
		encryptCall CipherCall
		decryptCall CipherCall
	}{
		{CFB, EncryptCFB, DecryptCFB},
		{CTR, EncryptCTR, DecryptCTR},
		{GCM, EncryptGCM, DecryptGCM},
	}
	for _, cipher := range ciphers {
		t.Run(cipher.name.String(), func(t *testing.T) {
			testCipher(t, cipher.encryptCall, cipher.decryptCall)
		})
	}
}

func testCipher(t *testing.T, encryptCall, decryptCall CipherCall) {
	const (
		secret    = "i-am-a-good-secret"
		plaintext = "Lorem ipsum dolor sit amet"
	)
	lens := []crypto.KeyLen{
		crypto.Key128,
		crypto.Key192,
		crypto.Key256,
	}
	for _, l := range lens {
		t.Run(l.String(), func(t *testing.T) {
			encrypted, err := encryptCall(l, []byte(secret), []byte(plaintext))
			assert.NoError(t, err)
			assert.NotEmpty(t, encrypted)
			data, err := crypto.DecodeData(encrypted)
			assert.NoError(t, err)
			assert.NotNil(t, data)
			assert.NoError(t, isDataValid(data))
			assert.Equal(t, l, data.KeyLen)
			decrypted, err := decryptCall(l, []byte(secret), encrypted)
			assert.NoError(t, err)
			assert.NotEmpty(t, decrypted)
			assert.Equal(t, plaintext, string(decrypted))
		})
	}
	// bad plain data
	_, err := encryptCall(crypto.Key256, []byte(secret), nil)
	assert.Error(t, err)
	// mismatch
	e, _ := encryptCall(crypto.Key256, []byte(secret), []byte(plaintext))
	d, _ := decryptCall(crypto.Key128, []byte(secret), e)
	assert.NotEqual(t, plaintext, string(d))
	// bad cipher data
	badData := [][]byte{
		nil,
		[]byte(""),
		[]byte("bad"),
	}
	for _, bd := range badData {
		_, err = decryptCall(crypto.Key256, []byte(secret), bd)
		assert.Error(t, err)
	}
	for _, bd := range badData {
		_, err = decryptCall(0, nil, bd)
		assert.Error(t, err)
	}
	for _, bd := range badData {
		_, err = decryptCall(math.MaxInt64, nil, bd)
		assert.Error(t, err)
	}
}
