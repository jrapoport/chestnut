package encryptor

import (
	"fmt"

	"github.com/jrapoport/chestnut/encryptor/aes"
	"github.com/jrapoport/chestnut/encryptor/crypto"
)

// AESEncryptor is an encryptor that supports the
// following AES keyLen lengths & cipher modes:
// 	- AES128-CFB, AES192-CFB, AES256-CFB
// 	- AES128-CTR, AES192-CTR, AES256-CTR
// 	- AES128-GCM, AES192-GCM, AES256-GCM
type AESEncryptor struct {
	secret crypto.Secret
	keyLen crypto.KeyLen
	mode   crypto.Mode
}

var _ crypto.Encryptor = (*AESEncryptor)(nil)

// NewAESEncryptor returns a new AESEncryptor configured
// with an AES keyLen length and mode for a secret.
func NewAESEncryptor(keyLen crypto.KeyLen, mode crypto.Mode, secret crypto.Secret) *AESEncryptor {
	ae := new(AESEncryptor)
	ae.secret = secret
	ae.keyLen = keyLen
	ae.mode = mode
	return ae
}

// ID returns the id of the encryptor (secret) that
// was used to encrypt the data (for tracking).
func (e *AESEncryptor) ID() string {
	return e.secret.ID()
}

// Name returns the name of the configured AES encryption cipher
// in following format "[cipher][keyLen length]-[mode]" e.g. "aes192-ctr".
func (e *AESEncryptor) Name() string {
	return crypto.CipherName("aes", e.keyLen, e.mode)
}

// Encrypt returns the plain data encrypted with the configured cipher mode and secret.
func (e *AESEncryptor) Encrypt(plaintext []byte) ([]byte, error) {
	var encryptCall aes.CipherCall
	switch e.mode {
	case aes.CFB:
		encryptCall = aes.EncryptCFB
	case aes.CTR:
		encryptCall = aes.EncryptCTR
	case aes.GCM:
		encryptCall = aes.EncryptGCM
	default:
		return nil, fmt.Errorf("unsupported encryption cipher mode: %s", e.mode)
	}
	return encryptCall(e.keyLen, e.secret.Open(), plaintext)
}

// Decrypt returns the cipher data decrypted with the configured cipher mode and secret.
func (e *AESEncryptor) Decrypt(ciphertext []byte) ([]byte, error) {
	var decryptCall aes.CipherCall
	switch e.mode {
	case aes.CFB:
		decryptCall = aes.DecryptCFB
	case aes.CTR:
		decryptCall = aes.DecryptCTR
	case aes.GCM:
		decryptCall = aes.DecryptGCM
	default:
		return nil, fmt.Errorf("unsupported decryption cipher mode: %s", e.mode)
	}
	return decryptCall(e.keyLen, e.secret.Open(), ciphertext)
}
