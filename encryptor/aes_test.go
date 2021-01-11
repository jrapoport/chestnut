package encryptor

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/jrapoport/chestnut/encryptor/aes"
	"github.com/jrapoport/chestnut/encryptor/crypto"
	"github.com/stretchr/testify/assert"
)

const testPlainText = "Lorem ipsum dolor sit amet"

var (
	textSecret    = crypto.TextSecret("i-am-a-good-secret")
	managedSecret = crypto.NewManagedSecret(uuid.New().String(), "i-am-a-managed-secret")
	secureSecret  = crypto.NewSecureSecret(uuid.New().String(), func(s crypto.Secret) []byte {
		return []byte(s.ID())
	})
)

func testAESEncryptor(t *testing.T, secret crypto.Secret, keyLen crypto.KeyLen, mode crypto.Mode) {
	ae := &AESEncryptor{secret, keyLen, mode}
	assert.Equal(t, secret.ID(), ae.ID())
	assert.Equal(t, crypto.CipherName("aes", keyLen, mode), ae.Name())
	e, err := ae.Encrypt([]byte(testPlainText))
	assert.NoError(t, err)
	assert.NotEmpty(t, e)
	d, err := ae.Decrypt(e)
	assert.NoError(t, err)
	assert.NotEmpty(t, d)
	assert.Equal(t, testPlainText, string(d))
}

func TestAESEncryptor(t *testing.T) {
	secrets := []struct {
		name string
		crypto.Secret
	}{
		{"TextSecret", textSecret},
		{"ManagedSecret", managedSecret},
		{"SecureSecret", secureSecret},
	}
	modes := []crypto.Mode{
		aes.CFB,
		aes.CTR,
		aes.GCM,
	}
	keyLens := []crypto.KeyLen{
		crypto.Key128,
		crypto.Key192,
		crypto.Key256,
	}
	testSecrets := func(t *testing.T, keyLen crypto.KeyLen, mode crypto.Mode) {
		for _, secret := range secrets {
			t.Run(secret.name, func(t *testing.T) {
				testAESEncryptor(t, secret, keyLen, mode)
			})
		}
	}
	testKeyLens := func(t *testing.T, mode crypto.Mode) {
		for _, keyLen := range keyLens {
			t.Run(keyLen.String(), func(t *testing.T) {
				testSecrets(t, keyLen, mode)
			})
		}
	}
	for _, mode := range modes {
		t.Run(mode.String(), func(t *testing.T) {
			testKeyLens(t, mode)
		})
	}
	// load a bad cipher
	const invalidMode = "Invalid_Cipher_Mode"
	ae := &AESEncryptor{textSecret, crypto.Key128, invalidMode}
	t.Run(fmt.Sprintf("%s_%s", invalidMode, "Encrypt"), func(t *testing.T) {
		// try to encrypt with a bad cipher
		_, err := ae.Encrypt([]byte(testPlainText))
		assert.Error(t, err)
	})
	t.Run(fmt.Sprintf("%s_%s", invalidMode, "Decrypt"), func(t *testing.T) {
		// try to decrypt with a bad cipher
		_, err := ae.Decrypt([]byte(testPlainText))
		assert.Error(t, err)
	})
}
