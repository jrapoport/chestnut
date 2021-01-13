package secure

import (
	"errors"
	"github.com/jrapoport/chestnut/log"
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
				WithLogger(log.Log),
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
	e := NewSecureEncoderExtension(encoders.InvalidID, PassthroughEncryption)
	assert.NotNil(t, e)
	assert.NotEmpty(t, e.encoderID )
	assert.Panics(t, func() {
		_ = NewSecureEncoderExtension(encoders.InvalidID, nil)
	})
}

func TestSecureEncoderExtension_BadSeal(t *testing.T) {
	var i int
	badCompressor := func(data []byte) (compressed []byte, err error) {
		if i % 2 != 0 && i < 10 {
			i++
			return nil, errors.New("compression error")
		}
		i++
		return nil, err
	}
	bade := true
	ext := NewSecureEncoderExtension(testEncoderID, func(plaintext []byte) (ciphertext []byte, err error) {
		if bade {
			return nil,  errors.New("encryption error")
		}
		return nil, err
	},
		WithCompressor(badCompressor))
	err := ext.Open()
	assert.NoError(t, err)
	i = 0
	ext.Close()
	ext.lookupBuffer = []byte("121343546432343546576453423142534653423142536435243142536463524")
	_, err = ext.Seal(bothEncoded)
	i = 1
	ext.Close()
	ext.lookupBuffer = []byte("121343546432343546576453423142534653423142536435243142536463524")
	_, err = ext.Seal(bothEncoded)
	i = 10
	ext.Close()
	assert.Error(t, err)
	ext.lookupBuffer = []byte("121343546432343546576453423142534653423142536435243142536463524")
	_, err = ext.Seal(bothEncoded)
	assert.Error(t, err)
	i = 10
	bade = false
	ext.Close()
	assert.Error(t, err)
	ext.lookupBuffer = []byte("121343546432343546576453423142534653423142536435243142536463524")
	ext.encoderID = encoders.InvalidID
	_, err = ext.Seal(bothEncoded)
	assert.Error(t, err)
	i = 10
	bade = false
	ext.Close()
	assert.Error(t, err)
	ext.lookupBuffer = []byte("121343546432343546576453423142534653423142536435243142536463524")
	ext.encoderID = testEncoderID
	ext.lookupCtx.Stream = nil
	_, err = ext.Seal(bothEncoded)
	assert.Error(t, err)
}

func TestSecureEncoderExtension_BadOpen(t *testing.T) {
	ext := NewSecureEncoderExtension(testEncoderID, PassthroughEncryption)
	err := ext.Open()
	assert.NoError(t, err)
	err = ext.Open()
	assert.Error(t, err)
	ext.Close()
    ctx := ext.lookupCtx
	ext.lookupCtx = nil
	err = ext.Open()
	assert.Error(t, err)
	ext.lookupCtx = ctx
	ext.lookupCtx.Token = encoders.InvalidID
	err = ext.Open()
	assert.Error(t, err)
	ext.lookupCtx = ctx
	ext.lookupCtx.Stream = nil
	err = ext.Open()
	assert.Error(t, err)
}

