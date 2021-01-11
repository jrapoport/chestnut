package lookup

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	jsoniter "github.com/json-iterator/go"
	"github.com/modern-go/reflect2"
)

// InvalidToken is an invalid lookup token.
const InvalidToken = ""

const tokenSeparator = "_"

// NewLookupToken returns the field name for sparse encoded data
// as the encoder id with the format "[prefix]-[encoder id]".
func NewLookupToken(prefix, encoderID string) string {
	return fmt.Sprintf("%s%s", prefix, encoderID)
}

// Key is an encoded lookup data key.
type Key string

// NewLookupKey creates a new lookup table key with the encoding field index and type. The field
// index is *not* the index relative to a StructField, but relative to the JSON encoding itself.
func NewLookupKey(token string, index int, typ reflect2.Type) Key {
	return Key(fmt.Sprintf("%s%d%s%d", token, index, tokenSeparator, typ.Kind()))
}

// IsTokenKey returns true if the key was derived from the lookup token.
func (k Key) IsTokenKey(token string) bool {
	return strings.HasPrefix(k.String(), token)
}

// Kind returns the encoded reflect.Kind for the key.
func (k Key) Kind() reflect.Kind {
	parts := strings.Split(k.String(), tokenSeparator)
	if len(parts) < 2 {
		return reflect.Invalid
	}
	// the last part should be the type
	kind, err := strconv.Atoi(parts[len(parts)-1])
	if err != nil {
		return reflect.Invalid
	}
	return reflect.Kind(kind)
}

func (k Key) String() string {
	return string(k)
}

// Context holds the context for the lookup coders.
type Context struct {
	Token  string
	Stream *jsoniter.Stream
}
