package aes

import (
	"crypto/aes"
	"crypto/cipher"

	"github.com/jrapoport/chestnut/encryptor/crypto"
)

type streamCipher func(block cipher.Block, iv []byte) cipher.Stream

// newStreamHeader returns a generic header suitable for aes stream ciphers that require an iv.
func newStreamHeader(keyLen crypto.KeyLen, mode crypto.Mode) (crypto.Header, error) {
	salt, err := crypto.MakeRand(crypto.SaltLength)
	if err != nil {
		return crypto.Header{}, err
	}
	iv, err := crypto.MakeRand(aes.BlockSize)
	if err != nil {
		return crypto.Header{}, err
	}
	return crypto.NewHeader("aes", keyLen, mode, salt, iv, nil)
}

// xorStreamEncrypt is a generic function for AES XOR stream encryption ciphers.
func xorStreamEncrypt(keyLen crypto.KeyLen, mode crypto.Mode, secret,
	plaintext []byte, newEncryptor streamCipher) ([]byte, error) {
	// create the header
	header, err := newStreamHeader(keyLen, mode)
	if err != nil {
		return nil, err
	}
	// encrypt the data
	encryptStream := func(_ crypto.Header, block cipher.Block, _ []byte) ([]byte, error) {
		ciphertext := make([]byte, len(plaintext))
		stream := newEncryptor(block, header.IV)
		stream.XORKeyStream(ciphertext, plaintext)
		return ciphertext, nil
	}
	return encrypt(keyLen, secret, plaintext, header, encryptStream)
}

// xorStreamDecrypt is a generic function for AES XOR stream decryption ciphers.
func xorStreamDecrypt(keyLen crypto.KeyLen, _ crypto.Mode, secret,
	ciphertext []byte, newDecrypter streamCipher) ([]byte, error) {
	// decrypt the data
	var decryptStream = func(header crypto.Header, block cipher.Block, data []byte) ([]byte, error) {
		plaintext := make([]byte, len(data))
		stream := newDecrypter(block, header.IV)
		stream.XORKeyStream(plaintext, data)
		// return the plain data
		return plaintext, nil
	}
	return decrypt(keyLen, secret, ciphertext, decryptStream)
}
