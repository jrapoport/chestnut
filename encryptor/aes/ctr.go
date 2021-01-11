package aes

import (
	"crypto/cipher"

	"github.com/jrapoport/chestnut/encryptor/crypto"
)

var (
	_ CipherCall = EncryptCTR // EncryptCTR conforms to CipherCall
	_ CipherCall = DecryptCTR // DecryptCTR conforms to CipherCall
)

// EncryptCTR supports AES128-CTR, AES192-CTR, and AES256-CTR encryption.
func EncryptCTR(length crypto.KeyLen, secret, plaintext []byte) ([]byte, error) {
	// encrypt the data
	return xorStreamEncrypt(length, CTR, secret, plaintext, cipher.NewCTR)
}

// DecryptCTR supports AES128-CTR, AES192-CTR, and AES256-CTR decryption.
func DecryptCTR(length crypto.KeyLen, secret, ciphertext []byte) ([]byte, error) {
	// decrypt the data
	return xorStreamDecrypt(length, CTR, secret, ciphertext, cipher.NewCTR)
}
