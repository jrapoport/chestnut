package crypto

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
)

// Data is a serializable wrapper for encrypted
// bytes with additional metadata in the Header.
type Data struct {
	Header
	Bytes []byte
}

// NewData returns an Data initialized
// with a Header and encrypted data.
func NewData(h Header, data []byte) Data {
	return Data{h, data}
}

// Valid returns an error if the Data is not valid.
func (e Data) Valid() error {
	if err := e.Header.Valid(); err != nil {
		return fmt.Errorf("invalid header %w", err)
	}
	// check the data
	if len(e.Bytes) <= 0 {
		return errors.New("invalid data")
	}
	return nil
}

// EncodeData encodes Data to a byte representation. This provides a small abstraction
// in case we want to swap out the gob encoder for something else.
func EncodeData(data Data) ([]byte, error) {
	if err := data.Valid(); err != nil {
		return nil, err
	}
	return GobEncodeData(data)
}

// DecodeData decodes a byte representation to Data. This provides a small abstraction
// in case we want to swap out the gob decoder for something else.
func DecodeData(b []byte) (Data, error) {
	return GobDecodeData(b)
}

// GobEncodeData serializes Data to a gob binary representation.
func GobEncodeData(data Data) ([]byte, error) {
	b := bytes.Buffer{}
	e := gob.NewEncoder(&b)
	if err := e.Encode(data); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

// GobDecodeData deserializes a gob binary representation to Data.
func GobDecodeData(b []byte) (Data, error) {
	data := Data{}
	buf := bytes.Buffer{}
	buf.Write(b)
	d := gob.NewDecoder(&buf)
	return data, d.Decode(&data)
}
