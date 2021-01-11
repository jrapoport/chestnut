package lookup

import (
	"errors"
	"fmt"
	"unsafe"

	"github.com/jrapoport/chestnut/encoding/json/encoders"
	"github.com/jrapoport/chestnut/log"
	jsoniter "github.com/json-iterator/go"
	"github.com/modern-go/reflect2"
)

var cleanEncoder = encoders.NewEncoder()

// Encoder is a ValEncoder that encodes the data to lookup table and encodes a
// entry key for the data into the stream that can be read later by the decoder.
type Encoder struct {
	token   string
	stream  *jsoniter.Stream
	valType reflect2.Type
	encoder jsoniter.ValEncoder
	log     log.Logger
}

// NewLookupEncoder returns an encoder that builds a lookup table. It will strip out tagged
// struct fields and collect the encoded values in the provided stream as a map. As it strips
// out values, it replaces them with a token key for the lookup table. Later we can use this
// key as a lookup to reconstruct the encoded struct as it is decoded. The hash encoder must
// be run before this encoder, so the struct fields are hashed before they are stripped.
func NewLookupEncoder(ctx *Context, typ reflect2.Type, encoder jsoniter.ValEncoder) jsoniter.ValEncoder {
	logger := log.Log
	if encoder == nil {
		logger.Fatal(errors.New("value encoder required"))
		return nil
	}
	if ctx == nil {
		logger.Fatal(errors.New("lookup context required"))
		return nil
	}
	if ctx.Token == InvalidToken {
		logger.Fatal(errors.New("lookup token required"))
		return nil
	}
	if ctx.Stream == nil {
		logger.Fatal(errors.New("lookup stream required"))
		return nil
	}
	return &Encoder{
		token:   ctx.Token,
		stream:  ctx.Stream,
		valType: typ,
		encoder: encoder,
		log:     logger,
	}
}

// SetLogger changes the logger for the encoder.
func (e *Encoder) SetLogger(l log.Logger) {
	e.log = l
}

// Encode writes the value of ptr to stream.
func (e *Encoder) Encode(ptr unsafe.Pointer, stream *jsoniter.Stream) {
	e.log.Debugf("encoding type %s", e.valType)
	// FIXME: I've looked around for a way to avoid this, or unwrap the encoder, but it's
	//  not clear what the best way to do that is or if it's possible with jsoniter as-is.
	// NOTE: This is *SUPER important*. This is so when UpdateStructDescriptor is called
	// recursively for nested structs the ValEncoder we use is a ORIGINAL ValEncoder, and
	// NOT a copy of our modified Encode (that strips out values). If we don't do this,
	// tagged fields will also be stripped out of our steam and not just the encoded stream.
	// We know this is happening because when it does: encoding stream == lookup stream.
	if stream == e.stream {
		// we are being called recursively so try and get a clean encoder.
		if subEncoder := cleanEncoder.EncoderOf(e.valType); subEncoder != nil {
			e.log.Debugf("use sub-encoder type %s", e.valType)
			// use the clean encoder to encode to our own stream.
			subEncoder.Encode(ptr, stream)
		} else {
			e.log.Error(fmt.Errorf("sub-encoder for type %s not found", e.valType))
		}
		return
	}
	// encode the ptr to the lookup table
	key := e.encodeLookup(ptr, e.nextIndex())
	// encode our lookup key to the main stream
	e.log.Debugf("encoded lookup key: %s", key)
	stream.WriteString(key.String())
}

// IsEmpty returns true is ptr is empty, otherwise false.
func (e *Encoder) IsEmpty(ptr unsafe.Pointer) bool {
	return e.encoder.IsEmpty(ptr)
}

func (e *Encoder) encodeLookup(ptr unsafe.Pointer, tableIndex int) Key {
	key := NewLookupKey(e.token, tableIndex, e.valType)
	// encode the actual data into our lookup table
	if tableIndex > 0 {
		e.stream.WriteMore()
	}
	e.stream.WriteObjectField(key.String())
	e.encoder.Encode(ptr, e.stream)
	e.log.Debugf("encoded lookup for key %s: %s", string(e.stream.Buffer()), key)
	return key
}

// we shouldn't need locking here since it should not to be called concurrently.
func (e *Encoder) nextIndex() int {
	idx, _ := e.stream.Attachment.(int)
	e.stream.Attachment = idx + 1
	return idx
}
