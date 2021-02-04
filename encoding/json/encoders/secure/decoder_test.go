package secure

import (
	"errors"
	"reflect"
	"testing"

	"github.com/jrapoport/chestnut/encoding/compress/zstd"
	"github.com/jrapoport/chestnut/encoding/json/encoders"
	"github.com/stretchr/testify/assert"
)

type decoderTest struct {
	src      []byte
	unsealed []byte
	dst      interface{}
	res      interface{}
	mp       map[string]interface{}
	sparse   Option
}

var decoderTests = []decoderTest{
	{noneSealed, noneUnsealed, &None{}, noneDecoded, noneMap, noOpt},
	{noneSealed, noneUnsealed, &None{}, noneDecoded, noneMap, sparseOpt},
	{noneComp, noneUnsealed, &None{}, noneDecoded, noneMap, noOpt},
	{noneComp, noneUnsealed, &None{}, noneDecoded, noneMap, sparseOpt},
	{jsonSealed, jsonUnsealed, &JSON{}, jsonDecoded, jsonMap, noOpt},
	{jsonSealed, jsonUnsealed, &JSON{}, jsonDecoded, jsonMap, sparseOpt},
	{jsonComp, jsonUnsealed, &JSON{}, jsonDecoded, jsonMap, noOpt},
	{jsonComp, jsonUnsealed, &JSON{}, jsonDecoded, jsonMap, sparseOpt},
	{hashSealed, hashUnsealed, &Hash{}, hashDecoded, hashMap, noOpt},
	{hashSealed, hashUnsealed, &Hash{}, hashDecoded, hashMap, sparseOpt},
	{hashComp, hashUnsealed, &Hash{}, hashDecoded, hashMap, noOpt},
	{hashComp, hashUnsealed, &Hash{}, hashDecoded, hashMap, sparseOpt},
	{secSealed, secUnsealed, &Secure{}, secDecoded, secMap, noOpt},
	{secSealed, secUnsealed, &Secure{}, secSparse, secMapSparse, sparseOpt},
	{secComp, secUnsealed, &Secure{}, secDecoded, secMap, noOpt},
	{secComp, secUnsealed, &Secure{}, secSparse, secMapSparse, sparseOpt},
	{bothSealed, bothUnsealed, &Both{}, bothDecoded, bothMap, noOpt},
	{bothSealed, bothUnsealed, &Both{}, bothSparse, bothMapSparse, sparseOpt},
	{bothComp, bothUnsealed, &Both{}, bothDecoded, bothMap, noOpt},
	{bothComp, bothUnsealed, &Both{}, bothSparse, bothMapSparse, sparseOpt},
	{allSealed, allUnsealed, &All{SI: ifc{}}, allDecoded, allMap, noOpt},
	{allSealed, allUnsealed, &All{SI: ifc{}}, allSparse, allMapSparse, sparseOpt},
	{allComp, allUnsealed, &All{SI: ifc{}}, allDecoded, allMap, noOpt},
	{allComp, allUnsealed, &All{SI: ifc{}}, allSparse, allMapSparse, sparseOpt},
}

func TestSecureDecoderExtension(t *testing.T) {
	for _, test := range decoderTests {
		testName := reflect.TypeOf(test.dst).Elem().Name()
		if test.sparse != nil {
			testName += " sparse"
		}
		t.Run(testName, func(t *testing.T) {
			encoder := encoders.NewEncoder()
			// register decoding extension
			decoderExt := NewSecureDecoderExtension(testEncoderID,
				PassthroughDecryption,
				WithDecompressor(zstd.Decompress),
				test.sparse)
			encoder.RegisterExtension(decoderExt)
			// unseal the encoding
			unsealed, err := decoderExt.Unseal(test.src)
			assert.NoError(t, err)
			assert.Equal(t, test.unsealed, unsealed)
			// open the decoder
			err = decoderExt.Open()
			assert.NoError(t, err)
			// securely decode the value
			err = encoder.Unmarshal(unsealed, test.dst)
			assert.NoError(t, err)
			assertDecoding(t, test.res, test.dst, err)
			// securely decode the reflected interface
			typ := reflect.ValueOf(test.dst).Elem().Type()
			ptr := reflect.New(typ).Interface()
			err = encoder.Unmarshal(unsealed, ptr)
			assertDecoding(t, test.res, ptr, err)
			// securely decode the mapped struct
			var mapped interface{}
			err = encoder.Unmarshal(unsealed, &mapped)
			assertDecoding(t, test.mp, mapped, err)
			// close the decoder
			decoderExt.Close()
		})
	}
	d := NewSecureDecoderExtension(encoders.InvalidID, PassthroughDecryption)
	assert.NotNil(t, d)
	assert.Empty(t, d.encoderID)
	assert.Panics(t, func() {
		_ = NewSecureDecoderExtension(encoders.InvalidID, nil)
	})
}

func TestSecureDecoderExtension_BadUnseal(t *testing.T) {
	var i int
	badCompressor := func(data []byte) (compressed []byte, err error) {
		if i%2 != 0 && i < 10 {
			i++
			return nil, errors.New("compression error")
		}
		i++
		return nil, err
	}
	bade := true
	ext := NewSecureDecoderExtension(testEncoderID, func(plaintext []byte) (ciphertext []byte, err error) {
		if bade {
			return nil, errors.New("encryption error")
		}
		return nil, err
	},
		WithCompressor(badCompressor))
	err := ext.Open()
	assert.NoError(t, err)
	err = ext.Open()
	assert.Error(t, err)
	_, err = ext.Unseal(bothEncoded)
	assert.Error(t, err)
	ext.Close()
	_, err = ext.Unseal(bothEncoded)
	assert.Error(t, err)
	_, err = ext.Unseal(bothSealed)
	assert.Error(t, err)
	bade = false
	_, err = ext.Unseal(bothComp)
	assert.Error(t, err)
	i = 1
	_, err = ext.Unseal(bothComp)
	i = 0
	ext.Close()
	encoder := encoders.NewEncoder()
	encoder.RegisterExtension(ext)
	err = encoder.Unmarshal(allComp, &None{})
	assert.Error(t, err)
	err = ext.Open()
	assert.NoError(t, err)
	assert.Panics(t, func() {
		ext.decryptFunc = nil
		_, err = ext.Unseal(bothComp)
		assert.Error(t, err)
	})
}

func TestSecureDecoderExtension_BadOpen(t *testing.T) {
	ext := NewSecureDecoderExtension(testEncoderID, PassthroughDecryption)
	err := ext.Open()
	assert.NoError(t, err)
	err = ext.Open()
	assert.Error(t, err)
	ext.Close()
	ext.lookupCtx = nil
	err = ext.Open()
	assert.Error(t, err)
	ext.Close()
}
