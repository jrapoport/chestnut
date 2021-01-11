package hash

import (
	"strings"
	"unsafe"

	"github.com/jrapoport/chestnut/log"
	jsoniter "github.com/json-iterator/go"
)

// Encoder is a ValEncoder strings that hashes string data
// with HashingFunction before encoding it to stream.
type Encoder struct {
	hashName string
	hashFunc HashingFunction
	encoder  jsoniter.ValEncoder
	log      log.Logger
}

// NewHashEncoder returns a string encoder that with encode string value using the supplied hashFn.
// The hash encoder will run before the other encoders, ensuring that struct fields are hashed first.
func NewHashEncoder(name string, hashFn HashingFunction, encoder jsoniter.ValEncoder) jsoniter.ValEncoder {
	if name == "" {
		name = "hash"
	}
	return &Encoder{hashName: name, hashFunc: hashFn, encoder: encoder, log: log.Log}
}

// SetLogger changes the logger for the encoder.
func (e *Encoder) SetLogger(l log.Logger) {
	e.log = l
}

// Encode writes the value of ptr to stream.
func (e *Encoder) Encode(ptr unsafe.Pointer, stream *jsoniter.Stream) {
	e.log.Debug("encoding hash")
	if e.IsEmpty(ptr) || e.hashFunc == nil {
		e.log.Warn("cannot encode empty ptr or nil hash function")
		e.encoder.Encode(ptr, stream)
		return
	}
	prefix := e.hashName + ":"
	if strings.HasPrefix(*((*string)(ptr)), prefix) {
		e.log.Warn("do not re-hash field")
		e.encoder.Encode(ptr, stream)
		return
	}
	data := *((*[]byte)(ptr))
	e.log.Debugf("hash string: %s", string(data))
	hash, err := e.hashFunc(data)
	if err == nil {
		hash = string(prefix) + hash
		e.log.Debugf("encoding hash: %s", hash)
		ptr = unsafe.Pointer(&hash)
	} else {
		e.log.Error(err)
	}
	e.encoder.Encode(ptr, stream)
}

// IsEmpty returns true is ptr is empty, otherwise false.
func (e *Encoder) IsEmpty(ptr unsafe.Pointer) bool {
	return e.encoder.IsEmpty(ptr)
}
