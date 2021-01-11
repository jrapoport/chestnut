package hash

import (
	"encoding/hex"

	"github.com/jrapoport/chestnut/encoding/tags"
	"github.com/jrapoport/chestnut/encryptor/crypto"
)

// HashingFunction defines the prototype for the hash callback. Defaults to EncodeToSHA256.
type HashingFunction func(buf []byte) (hash string, err error)

// FunctionForName returns the hash function for a given otherwise nil (passthrough).
func FunctionForName(name tags.Hash) HashingFunction {
	switch name {
	case tags.HashSHA256:
		return EncodeToSHA256
	default:
		return nil
	}
}

// EncodeToSHA256 returns a sha256 hash of data as string.
var EncodeToSHA256 = func(buf []byte) (string, error) {
	hash, err := crypto.HashSHA256(buf)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(hash), nil
}
