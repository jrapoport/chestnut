package aes

import (
	"crypto/cipher"

	"github.com/jrapoport/chestnut/encryptor/crypto"
)

var (
	_ CipherCall = EncryptGCM // EncryptGCM conforms to CipherCall
	_ CipherCall = DecryptGCM // DecryptGCM conforms to CipherCall
)

// newGMCHeader returns a header containing a nonce suitable for a gcm cipher.
func newGMCHeader(keyLen crypto.KeyLen) (crypto.Header, error) {
	salt, err := crypto.MakeSalt()
	if err != nil {
		return crypto.Header{}, err
	}
	nonce, err := crypto.MakeNonce()
	if err != nil {
		return crypto.Header{}, err
	}
	return crypto.NewHeader("aes", keyLen, GCM, salt, nil, nonce)
}

// EncryptGCM supports AES128-GCM, AES192-GCM, and AES256-GCM encryption.
func EncryptGCM(keyLen crypto.KeyLen, secret, plaintext []byte) ([]byte, error) {
	// create the header
	header, err := newGMCHeader(keyLen)
	if err != nil {
		return nil, err
	}
	// seal the data with gcms
	sealData := func(_ crypto.Header, block cipher.Block, _ []byte) ([]byte, error) {
		// create the AHEAD
		gcm, gcmErr := cipher.NewGCM(block)
		if gcmErr != nil {
			return nil, gcmErr
		}
		// encrypt the data
		return gcm.Seal(nil, header.Nonce, plaintext, nil), nil
	}
	return encrypt(keyLen, secret, plaintext, header, sealData)
}

// DecryptGCM supports AES128-GCM, AES192-GCM, and AES256-GCM decryption.
func DecryptGCM(keyLen crypto.KeyLen, secret, ciphertext []byte) ([]byte, error) {
	// open the data with gcm
	openData := func(header crypto.Header, block cipher.Block, data []byte) ([]byte, error) {
		// create the AHEAD
		gcm, err := cipher.NewGCM(block)
		if err != nil {
			return nil, err
		}
		// decrypt the data
		return gcm.Open(nil, header.Nonce, data, nil)
	}
	return decrypt(keyLen, secret, ciphertext, openData)
}
