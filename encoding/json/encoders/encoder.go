package encoders

import (
	"encoding/hex"
	"log"

	"github.com/jrapoport/chestnut/encryptor/crypto"
	"github.com/json-iterator/go"
)

// NewEncoder returns a new encoder with a _clean_ configuration and _no_ registered
// extensions. Extensions registered to this encoder will not impact the global encoder.
// Config options match jsoniter ConfigCompatibleWithStandardLibrary.
func NewEncoder() jsoniter.API {
	return jsoniter.Config{
		EscapeHTML:             true,
		SortMapKeys:            true,
		ValidateJsonRawMessage: true,
	}.Froze()
}

// DefaultID can be used with a SecureEncoderExtension instead of a set id. When used,
// it will be replaced with a randomly generated 8 character hex id for the encoder.
// #954535 is hex color code for Chestnut. https://en.wikipedia.org/wiki/Chestnut_(color)
var DefaultID = "0x954535"

// InvalidID is an invalid encoder id.
const InvalidID = ""

// NewEncoderID returns a new random encoder id as a hex string. This id
// is not guaranteed to be unique and does not have to be. It is only used
// internally in the encoder so there is no risk of collision.
func NewEncoderID() string {
	id, err := crypto.MakeRand(4)
	if err != nil {
		log.Fatal(err)
	}
	return hex.EncodeToString(id)
}
