package secure

import (
	"reflect"
	"testing"

	"github.com/jrapoport/chestnut/encoding/json/encoders"
	"github.com/jrapoport/chestnut/encoding/json/packager"
	"github.com/stretchr/testify/assert"
)

type encoderTest struct {
	src        interface{}
	dst        []byte
	sealed     []byte
	compressed Option
}

var encoderTests = []encoderTest{
	{noneObj, noneEncoded, noneSealed, noOpt},
	{noneObj, noneEncoded, noneComp, compOpt},
	{jsonObj, jsonEncoded, jsonSealed, noOpt},
	{jsonObj, jsonEncoded, jsonComp, compOpt},
	{hashObj, hashEncoded, hashSealed, noOpt},
	{hashObj, hashEncoded, hashComp, compOpt},
	{secObj, secEncoded, secSealed, noOpt},
	{secObj, secEncoded, secComp, compOpt},
	{bothObj, bothEncoded, bothSealed, noOpt},
	{bothObj, bothEncoded, bothComp, compOpt},
	{allObj, allEncoded, allSealed, noOpt},
	{allObj, allEncoded, allComp, compOpt},
}

func TestSecureEncoderExtension(t *testing.T) {
	for _, test := range encoderTests {
		testName := reflect.TypeOf(test.src).Elem().Name()
		if test.compressed != nil {
			testName += " compressed"
		}
		t.Run(testName, func(t *testing.T) {
			encoder := encoders.NewEncoder()
			// register encoding extension
			encoderExt := NewSecureEncoderExtension(testEncoderID,
				PassthroughEncryption,
				test.compressed)
			encoder.RegisterExtension(encoderExt)
			// open the encoder
			err := encoderExt.Open()
			assert.NoError(t, err)
			// securely encode the value
			encoded, err := encoder.Marshal(test.src)
			assertJSON(t, test.dst, encoded, err)
			// close the encoder
			encoderExt.Close()
			// seal the encoding
			sealed, err := encoderExt.Seal(encoded)
			assert.NoError(t, err)
			assert.Equal(t, test.sealed, sealed)
			// unwrap the sealed package & ake sure it is valid
			pkg, err := packager.DecodePackage(sealed)
			assert.NoError(t, err)
			assert.NotNil(t, pkg)
			assert.NoError(t, pkg.Valid())
		})
	}
}
