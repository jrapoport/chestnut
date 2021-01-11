package json

import (
	"errors"

	"github.com/jrapoport/chestnut/encoding/json/encoders"
	"github.com/jrapoport/chestnut/encoding/json/encoders/secure"
)

// SecureMarshal returns an encrypted JSON encoding of v. It adds support for sparse encryption and
// hashing via JSON struct tag options. If SecureMarshal is called at least one 'secure' option set
// on a struct field JSON tag, only those fields will be encrypted. The remaining encoded data stored
// as sparse plaintext. If no secure tag option is found, all the encoded data will be encrypted.
// For more detail, SEE: https://github.com/jrapoport/chestnut/blob/master/README.md
func SecureMarshal(v interface{}, encryptFunc secure.EncryptionFunction, opt ...secure.Option) ([]byte, error) {
	if v == nil {
		return nil, errors.New("nil value")
	}
	enc := encoders.NewEncoder()
	ext := secure.NewSecureEncoderExtension(encoders.DefaultID, encryptFunc, opt...)
	enc.RegisterExtension(ext)
	if err := ext.Open(); err != nil {
		return nil, err
	}
	buf, err := enc.Marshal(v)
	if err != nil {
		return nil, err
	}
	ext.Close()
	return ext.Seal(buf)
}
