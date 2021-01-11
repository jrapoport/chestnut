package aes

import (
	"crypto/cipher"

	"github.com/jrapoport/chestnut/encryptor/crypto"
)

var (
	_ CipherCall = EncryptCFB // EncryptCFB conforms to CipherCall
	_ CipherCall = DecryptCFB // DecryptCFB conforms to CipherCall
)

// EncryptCFB supports AES128-CFB, AES192-CFB, and AES256-CFB encryption.
func EncryptCFB(length crypto.KeyLen, secret, plaintext []byte) ([]byte, error) {
	// encrypt the data
	return xorStreamEncrypt(length, CFB, secret, plaintext, cipher.NewCFBEncrypter)
}

// DecryptCFB supports AES128-CFB, AES192-CFB, and AES256-CFB decryption.
func DecryptCFB(length crypto.KeyLen, secret, ciphertext []byte) ([]byte, error) {
	// decrypt the data
	return xorStreamDecrypt(length, CFB, secret, ciphertext, cipher.NewCFBDecrypter)
}
