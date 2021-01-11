package encryptor

import (
	"strings"

	"github.com/jrapoport/chestnut/encryptor/crypto"
)

// ChainEncryptor is an encryptor that supports an chain of other Encryptors.
// Bytes will be encrypted by chaining the Encryptors in a FIFO order.
type ChainEncryptor struct {
	id         string
	name       string
	ids        []string
	names      []string
	encryption []crypto.Encryptor
	decryption []crypto.Encryptor
}

var _ crypto.Encryptor = (*ChainEncryptor)(nil)

const chainSep = " "

// NewChainEncryptor creates a new ChainEncryptor consisting of a chain
// of the supplied Encryptors.
func NewChainEncryptor(encryptors ...crypto.Encryptor) *ChainEncryptor {
	if len(encryptors) == 0 {
		return nil
	}
	// reverse the encryptors from FIFO to LIFO
	decryptors := make([]crypto.Encryptor, len(encryptors))
	for i := range encryptors {
		decryptors[len(encryptors)-1-i] = encryptors[i]
	}
	chain := new(ChainEncryptor)
	chain.encryption = encryptors
	chain.decryption = decryptors
	chain.ids = make([]string, len(encryptors))
	chain.names = make([]string, len(encryptors))
	for i, e := range chain.encryption {
		chain.ids[i] = e.ID()
		chain.names[i] = e.Name()
	}
	chain.id = strings.Join(chain.ids, chainSep)
	chain.name = strings.Join(chain.names, chainSep)
	return chain
}

// ID returns a concatenated list of the ids of chained encryptor(s) / secrets
// that were used to encrypt the data (for tracking) separated by spaces.
func (e *ChainEncryptor) ID() string {
	return e.id
}

// Name returns a concatenated list of the cipher names of the chained encryptor(s)
// that were used to encrypt the data separated by spaces.
func (e *ChainEncryptor) Name() string {
	return e.name
}

// Encrypt returns data encrypted with the chain of Encryptors.
func (e *ChainEncryptor) Encrypt(plaintext []byte) ([]byte, error) {
	var err error
	ciphertext := plaintext
	for _, en := range e.encryption {
		ciphertext, err = en.Encrypt(ciphertext)
		if err != nil {
			break
		}
	}
	return ciphertext, err
}

// Decrypt returns data decrypted with the chain of Encryptors.
func (e *ChainEncryptor) Decrypt(ciphertext []byte) ([]byte, error) {
	var err error
	plaintext := ciphertext
	for _, de := range e.decryption {
		plaintext, err = de.Decrypt(plaintext)
		if err != nil {
			break
		}
	}
	return plaintext, err
}
