package secure

import (
	"reflect"
	"testing"

	"github.com/jrapoport/chestnut/encoding/compress"
	"github.com/jrapoport/chestnut/encoding/json/encoders"
	jsoniter "github.com/json-iterator/go"
	"github.com/stretchr/testify/assert"
)

const testEncoderID = "86fb3fa0"

type testCase struct {
	src    interface{}
	dst    interface{}
	res    interface{}
	mp     map[string]interface{}
	sparse Option
}

var (
	noOpt = EmptyOption{}
	// ignored on non-sparse packages
	sparseOpt = SparseDecode()
	compOpt   = WithCompression(compress.Zstd)
)

var tests = []testCase{
	{noneObj, &None{}, noneDecoded, noneMap, noOpt},
	{noneObj, &None{}, noneDecoded, noneMap, sparseOpt},
	{jsonObj, &JSON{}, jsonDecoded, jsonMap, noOpt},
	{jsonObj, &JSON{}, jsonDecoded, jsonMap, sparseOpt},
	{hashObj, &Hash{}, hashDecoded, hashMap, noOpt},
	{hashObj, &Hash{}, hashDecoded, hashMap, sparseOpt},
	{secObj, &Secure{}, secDecoded, secMap, noOpt},
	{secObj, &Secure{}, secSparse, secMapSparse, sparseOpt},
	{bothObj, &Both{}, bothDecoded, bothMap, noOpt},
	{bothObj, &Both{}, bothSparse, bothMapSparse, sparseOpt},
	{allObj, &All{SI: ifc{}}, allDecoded, allMap, noOpt},
	{allObj, &All{SI: ifc{}}, allSparse, allMapSparse, sparseOpt},
}

func TestSecureExtension(t *testing.T) {
	comps := []Option{noOpt, compOpt}
	for _, compressed := range comps {
		for _, test := range tests {
			testName := reflect.TypeOf(test.dst).Elem().Name()
			if test.sparse != nil {
				testName += " sparse"
			}
			if compressed != nil {
				testName += " compressed"
			}
			t.Run(testName, func(t *testing.T) {
				encoder := encoders.NewEncoder()
				// register encoding extension
				encoderExt := NewSecureEncoderExtension(testEncoderID,
					PassthroughEncryption, compressed)
				encoder.RegisterExtension(encoderExt)
				// register decoding extension
				decoderExt := NewSecureDecoderExtension(testEncoderID,
					PassthroughDecryption, compressed, test.sparse)
				encoder.RegisterExtension(decoderExt)
				// open the encoder
				err := encoderExt.Open()
				assert.NoError(t, err)
				// securely encode the value
				encoded, err := encoder.Marshal(test.src)
				assert.NoError(t, err)
				// close the encoder
				encoderExt.Close()
				// seal the encoding
				sealed, err := encoderExt.Seal(encoded)
				assert.NoError(t, err)
				// unseal the encoding
				unsealed, err := decoderExt.Unseal(sealed)
				assert.NoError(t, err)
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
	}
}

func assertJSON(t *testing.T, expected, actual []byte, err error) {
	e := assert.NoError(t, err)
	if !e {
		t.Fatal(err)
	}
	valid := jsoniter.Valid(actual)
	assert.True(t, valid, "invalid JSON")
	assert.Equal(t, string(expected), string(actual))
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
