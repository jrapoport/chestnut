package json

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSecureUnmarshal(t *testing.T) {
	// uncompressed secure
	secureObj := &Family{}
	err := SecureUnmarshal(familyEnc, secureObj, decrypt)
	assertDecoding(t, familyDec, secureObj, err)
	// compressed secure
	secureObj = &Family{}
	err = SecureUnmarshal(familyComp, secureObj, decrypt, compOpt)
	assertDecoding(t, familyDec, secureObj, err)
	// uncompressed sparse
	sparseObj := &Family{}
	err = SecureUnmarshal(familyEnc, sparseObj, decrypt, sparseOpt)
	assertDecoding(t, familySpr, sparseObj, err)
	// compressed sparse
	sparseObj = &Family{}
	err = SecureUnmarshal(familyComp, sparseObj, decrypt, compOpt, sparseOpt)
	assertDecoding(t, familySpr, sparseObj, err)
}

func TestSecureUnmarshal_Error(t *testing.T) {
	secureObj := &Family{}
	assert.Panics(t, func() {
		_ = SecureUnmarshal(familyEnc, secureObj, nil)
	})
	err := SecureUnmarshal(familyEnc, nil, decrypt)
	assert.Error(t, err)
	err = SecureUnmarshal(nil, secureObj, decrypt)
	assert.Error(t, err)
	err = SecureUnmarshal([]byte("bad encoding"), secureObj, decrypt)
	assert.Error(t, err)
	var p chan bool
	err = SecureUnmarshal(familyEnc, &p, decrypt)
	assert.Error(t, err)
}
