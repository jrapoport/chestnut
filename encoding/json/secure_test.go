package json

import (
	"reflect"
	"testing"

	"github.com/jrapoport/chestnut/encoding/compress"
	"github.com/jrapoport/chestnut/encoding/json/encoders/secure"
	"github.com/stretchr/testify/assert"
)

var (
	encrypt   = secure.PassthroughEncryption
	decrypt   = secure.PassthroughDecryption
	compOpt   = secure.WithCompression(compress.Zstd)
	sparseOpt = secure.SparseDecode()
)

func TestSecureEncoding(t *testing.T) {
	secureObj := &Family{}
	bytes, err := SecureMarshal(family, encrypt)
	assert.NoError(t, err)
	assert.Equal(t, familyEnc, bytes)
	err = SecureUnmarshal(bytes, secureObj, decrypt)
	assertDecoding(t, familyDec, secureObj, err)
}

func TestCompressedEncoding(t *testing.T) {
	secureObj := &Family{}
	bytes, err := SecureMarshal(family, encrypt, compOpt)
	assert.NoError(t, err)
	assert.Equal(t, familyComp, bytes)
	err = SecureUnmarshal(bytes, secureObj, decrypt, compOpt)
	assertDecoding(t, familyDec, secureObj, err)
}

func TestSparseDecoding(t *testing.T) {
	sparseObj := &Family{}
	bytes, err := SecureMarshal(family, encrypt)
	assert.NoError(t, err)
	assert.Equal(t, familyEnc, bytes)
	err = SecureUnmarshal(bytes, sparseObj, decrypt, sparseOpt)
	assertDecoding(t, familySpr, sparseObj, err)
}

func TestCompressedSparseDecoding(t *testing.T) {
	sparseObj := &Family{}
	bytes, err := SecureMarshal(family, encrypt, compOpt)
	assert.NoError(t, err)
	assert.Equal(t, familyComp, bytes)
	err = SecureUnmarshal(bytes, sparseObj, decrypt, compOpt, sparseOpt)
	assertDecoding(t, familySpr, sparseObj, err)
}

func assertDecoding(t *testing.T, expected, actual interface{}, err error) {
	e := assert.NoError(t, err)
	if !e {
		t.Fatal(err)
	}
	assert.Equal(t, expected, actual)
	deep := reflect.DeepEqual(expected, actual)
	assert.True(t, deep, "values are not deep equal")
}
