package crypto

// Encryptor is the interface use to supply cipher implementations to the datastore.
type Encryptor interface {
	// ID returns the id of the secret used to encrypt the data.
	ID() string

	// Name returns the name of encryption cipher, keyLen length
	// and mode used to encrypt the data ("aes192-ctr").
	Name() string

	// Encrypt returns data encrypted with the secret.
	Encrypt(plaintext []byte) (ciphertext []byte, err error)

	// Decrypt returns data decrypted with the secret.
	Decrypt(ciphertext []byte) (plaintext []byte, err error)
}
