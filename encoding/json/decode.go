package json

import (
	"errors"

	"github.com/jrapoport/chestnut/encoding/json/encoders"
	"github.com/jrapoport/chestnut/encoding/json/encoders/secure"
)

// SecureUnmarshal decrypts & parses the JSON-encoded data returned by SecureUnmarshal and stores
// the result in the value pointed to by v. If v is nil or not a pointer, Unmarshal returns an
// error. SecureUnmarshal adds support for sparse decryption and via JSON struct tag options. If
// SecureMarshal is called at least one 'secure' option set on a struct field JSON tag, only those
// fields will be encrypted. The remaining encoded data stored as sparse plaintext. If SecureUnmarshal
// is called on a sparse encoding with the sparse option set, SecureUnmarshal will skip the decryption
// step and return only the plaintext decoding of v with encrypted fields replaced by empty values.
// For more detail, SEE: https://github.com/jrapoport/chestnut/blob/master/README.md
func SecureUnmarshal(data []byte, v interface{}, decryptFunc secure.DecryptionFunction, opt ...secure.Option) error {
	if v == nil {
		return errors.New("nil value")
	}
	enc := encoders.NewEncoder()
	ext := secure.NewSecureDecoderExtension(encoders.DefaultID, decryptFunc, opt...)
	enc.RegisterExtension(ext)
	defer ext.Close()
	unsealed, err := ext.Unseal(data)
	if err != nil {
		return err
	}
	if err = ext.Open(); err != nil {
		return err
	}
	return enc.Unmarshal(unsealed, v)
}
