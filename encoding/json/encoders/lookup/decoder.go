package lookup

import (
	"errors"
	"fmt"
	"reflect"
	"unsafe"

	"github.com/jrapoport/chestnut/log"
	jsoniter "github.com/json-iterator/go"
	"github.com/modern-go/reflect2"
)

// Decoder is a ValDecoder that reads lookup table key strings from the iterator. When
// a key is found in the lookup table it decodes the lookup table data in place of the key.
type Decoder struct {
	token   string
	stream  *jsoniter.Stream
	valType reflect2.Type
	decoder jsoniter.ValDecoder
	log     log.Logger
}

// NewLookupDecoder returns an decoder that reads a lookup table. It will check the
// iterated string values to see if they match our lookup token. If there is a match,
// it will replace it with a decoded value from the lookup table or an empty value.
func NewLookupDecoder(ctx *Context, typ reflect2.Type, decoder jsoniter.ValDecoder) jsoniter.ValDecoder {
	logger := log.Log
	if decoder == nil {
		logger.Panic(errors.New("value encoder required"))
		return nil
	}
	if typ == nil {
		logger.Panic(errors.New("decoder typ required"))
		return nil
	}
	if ctx == nil {
		logger.Panic(errors.New("lookup context required"))
		return nil
	}
	if ctx.Token == "" {
		logger.Panic(errors.New("lookup token required"))
		return nil
	}
	if ctx.Stream == nil {
		logger.Panic(errors.New("lookup stream required"))
		return nil
	}
	return &Decoder{
		token:   ctx.Token,
		stream:  ctx.Stream,
		valType: typ,
		decoder: decoder,
		log:     logger,
	}
}

// SetLogger changes the logger for the decoder.
func (d *Decoder) SetLogger(l log.Logger) {
	d.log = l
}

// Decode sets ptr to the next value of iterator.
func (d *Decoder) Decode(ptr unsafe.Pointer, iter *jsoniter.Iterator) {
	d.log.Debugf("decoding type %s", d.valType)
	// if we are dealing with an empty interface, skip it.
	if d.isEmptyInterface(ptr) {
		d.log.Warn("cannot encode to empty interface")
		iter.Skip()
		return
	}
	// we really shouldn't be here with an invalid token, if for
	// some reason we are, call the default decoder and bail.
	if d.token == InvalidToken {
		d.log.Warn("invalid token")
		d.decoder.Decode(ptr, iter)
		return
	}
	// get the from type
	fromType := iter.WhatIsNext()
	// secure tokens will be type string. if this is not
	// a string, call the default decoder and bail.
	if fromType != jsoniter.StringValue {
		d.log.Debug("skipping non-string value")
		d.decoder.Decode(ptr, iter)
		return
	}
	// read the string & for mat a key
	key := Key(iter.ReadString())
	// check to see if it is one of ours
	if !key.IsTokenKey(d.token) {
		// we use an Iterator avoid setting the ptr directly since it might be a string
		// or an interface or who knows what. this was the codecs handle it for us.
		subIter := iter.Pool().BorrowIterator([]byte(fmt.Sprintf(`"%s"`, key)))
		defer iter.Pool().ReturnIterator(subIter)
		d.log.Debugf("decode string: %s", key)
		// decode the string
		d.decoder.Decode(ptr, subIter)
		return
	}
	// we have a valid lookup key. look it up in our table
	val, err := d.lookupKey(key)
	// did we find something in the lookup table?
	if err != nil || val == nil {
		d.log.Debugf("lookup entry not found: %s", key)
		// this is expected when sparse decoding a struct.
		if d.valType.Kind() == reflect.Interface {
			d.log.Debugf("decode empty %s for interface", key.Kind())
			// if we have a map then set an explicitly typed empty value
			*(*interface{})(ptr) = emptyValueOfKind(key.Kind())
		}
		return
	}
	// clear the buffer
	d.stream.Reset(nil)
	val.WriteTo(d.stream)
	subIter := iter.Pool().BorrowIterator(d.stream.Buffer())
	defer iter.Pool().ReturnIterator(subIter)
	// decode the string
	d.decoder.Decode(ptr, subIter)
	d.log.Debugf("decoded lookup entry for %s: %s", key, string(d.stream.Buffer()))
}

func (d *Decoder) lookupKey(key Key) (jsoniter.Any, error) {
	d.log.Debugf("lookup key: %s", key)
	logErr := func(err error) error {
		d.log.Error(err)
		return err
	}
	if d.stream == nil {
		return nil, logErr(errors.New("lookup stream not found"))
	}
	table, ok := d.stream.Attachment.(jsoniter.Any)
	if !ok || table == nil {
		return nil, logErr(errors.New("lookup table not found"))
	}
	val := table.Get(key.String())
	if val.ValueType() == jsoniter.InvalidValue {
		err := fmt.Errorf("lookup key not found: %s", key)
		d.log.Debug(err) // this is an expected error
		return nil, err
	}
	d.log.Debugf("lookup found %s for key %s: %s", val.ValueType(), key, val.ToString())
	return val, nil
}

func (d *Decoder) isEmptyInterface(ptr unsafe.Pointer) bool {
	if d.valType.Kind() != reflect.Interface {
		return false
	}
	i, ok := d.valType.(*reflect2.UnsafeIFaceType)
	if !ok {
		return false
	}
	return reflect2.IsNil(i.UnsafeIndirect(ptr))
}

func emptyValueOfKind(kind reflect.Kind) interface{} {
	var v interface{}
	switch kind {
	case reflect.String:
		v = ""
	case reflect.Bool:
		v = false
	case reflect.Uint8, reflect.Int8,
		reflect.Uint16, reflect.Int16,
		reflect.Uint32, reflect.Int32,
		reflect.Uint64, reflect.Int64,
		reflect.Uint, reflect.Int,
		reflect.Float32, reflect.Float64,
		reflect.Uintptr:
		v = 0.0
	default:
	}
	return v
}
