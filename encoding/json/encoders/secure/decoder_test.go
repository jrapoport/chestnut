package secure

import (
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
	assert.Empty(t, d.encoderID )
	assert.Panics(t, func() {
		_ = NewSecureDecoderExtension(encoders.InvalidID, nil)
	})
}
