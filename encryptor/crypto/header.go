package crypto

import (
	"errors"
	"fmt"
	"strings"
)

// MinSaltLength is the minimum length of the salt buffer.
const MinSaltLength = 8

// A Header describes an encryption block. It contains the cipher name,
// key length, mode used as well as the cipher key salt, iv or nonce.
type Header struct {
	Cipher string // e.g. "aes"
	KeyLen KeyLen // e.g. 128
	Mode   Mode   // e.g. "gcm"
	Salt   []byte
	IV     []byte
	Nonce  []byte
}

// NewHeader create a new Header checking the length of the
// salt buffer against MinSaltLength. If the length of the
// salt buffer is less than MinSaltLength it returns an error.
func NewHeader(cipher string, keyLen KeyLen, mode Mode, salt []byte, iv []byte, nonce []byte) (Header, error) {
	cipher = strings.ToLower(cipher)
	mode = Mode(strings.ToLower(mode.String()))
	h := Header{cipher, keyLen, mode, salt, iv, nonce}
	if err := h.Valid(); err != nil {
		return Header{}, err
	}
	return h, nil
}

// Valid returns an error if the Header is not valid.
func (h Header) Valid() error {
	if h.Cipher == "" {
		return errors.New("cipher required")
	}
	if h.KeyLen <= 0 {
		return errors.New("key length required")
	}
	if h.Mode == "" {
		return errors.New("mode required")
	}
	if len(h.Salt) < MinSaltLength {
		return fmt.Errorf("salt length %d < %d minimum", len(h.Salt), MinSaltLength)
	}
	if h.Nonce != nil && len(h.Nonce) < NonceLength {
		return fmt.Errorf("nonce length %d < %d minimum", len(h.Nonce), NonceLength)
	}
	return nil
}

// Name returns the name of the cipher in following
// format "[cipher][key length]-[mode]" e.g. "aes192-ctr".
func (h *Header) Name() string {
	return CipherName(h.Cipher, h.KeyLen, h.Mode)
}

// CipherName is a convenience function that returns the name,
// key length, and mode of a cipher in the following format
// "[cipher][key length]-[mode]" e.g. "aes192-ctr".
func CipherName(cipher string, keyLen KeyLen, mode Mode) string {
	return fmt.Sprintf("%s%s-%s", strings.ToLower(cipher), keyLen, strings.ToLower(mode.String()))
}
