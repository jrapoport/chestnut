package json

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSecureMarshal(t *testing.T) {
	// uncompressed
	bytes, err := SecureMarshal(family, encrypt)
	assert.NoError(t, err)
	assert.Equal(t, familyEnc, bytes)
	// compressed
	bytes, err = SecureMarshal(family, encrypt, compOpt)
	assert.NoError(t, err)
	assert.Equal(t, familyComp, bytes)
}

func TestSecureMarshal_Error(t *testing.T) {
	assert.Panics(t, func() {
		_, _ = SecureMarshal(family, nil)
	})
	bytes, err := SecureMarshal(nil, nil)
	assert.Error(t, err)
	assert.Nil(t, bytes)
	bytes, err = SecureMarshal(nil, encrypt)
	assert.Error(t, err)
	assert.Nil(t, bytes)
	var p chan bool
	bytes, err = SecureMarshal(p, encrypt)
	assert.Error(t, err)
	assert.Nil(t, bytes)
}
