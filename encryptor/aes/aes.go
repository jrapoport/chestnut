package aes

import (
	"crypto/aes"
	"crypto/cipher"
	"errors"

	"github.com/jrapoport/chestnut/encryptor/crypto"
)

// currently supported modes
const (
	CFB crypto.Mode = "cfb"
	CTR             = "ctr"
	GCM             = "gcm"
)

// CipherCall is function the prototype for the encryption and decryption.
type CipherCall func(length crypto.KeyLen, secret, data []byte) ([]byte, error)

// cipherTransform preforms the encryption or decryption and returns the result.
type cipherTransform func(header crypto.Header, block cipher.Block, data []byte) ([]byte, error)

// encrypt is a generalized AES decryption function that takes plaintext and return a serialized Entry.
func encrypt(keyLen crypto.KeyLen, secret, plaintext []byte, header crypto.Header, encryptT cipherTransform) ([]byte, error) {
	if plaintext == nil || len(plaintext) <= 0 {
		return nil, errors.New("invalid plain data")
	}
	// create the cipher key
	key, err := crypto.NewCipherKey(keyLen, secret, header.Salt)
	if err != nil {
		return nil, err
	}
	// create the cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	ciphertext, err := encryptT(header, block, plaintext)
	if err != nil {
		return nil, err
	}
	// check the result
	data := crypto.NewData(header, ciphertext)
	if err = isDataValid(data); err != nil {
		return nil, err
	}
	// encode the encrypted data and return the result
	return crypto.EncodeData(data)
}

// decrypt is a generalized AES decryption function that takes a serialized Entry and returns plaintext.
func decrypt(keyLen crypto.KeyLen, secret, ciphertext []byte, decryptT cipherTransform) ([]byte, error) {
	if ciphertext == nil || len(ciphertext) <= 0 {
		return nil, errors.New("invalid cipher data")
	}
	// decode the encrypted data
	data, err := crypto.DecodeData(ciphertext)
	if err != nil {
		return nil, err
	}
	// check the encoding
	if err = isDataValid(data); err != nil {
		return nil, err
	}
	// get the cipher key
	key, err := crypto.NewCipherKey(keyLen, secret, data.Salt)
	if err != nil {
		return nil, err
	}
	// get the cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	// decrypt the data
	return decryptT(data.Header, block, data.Bytes)
}

func isDataValid(data crypto.Data) error {
	if err := data.Valid(); err != nil {
		return err
	}
	// check the iv
	if data.IV != nil && len(data.IV) < aes.BlockSize {
		return errors.New("invalid iv")
	}
	return nil
}
