package hash

import (
	"testing"

	"github.com/jrapoport/chestnut/encoding/tags"
	"github.com/stretchr/testify/assert"
)

func TestHashFunctionForName(t *testing.T) {
	fn := FunctionForName(tags.HashNone)
	assert.Nil(t, fn)
	fn = FunctionForName(tags.HashSHA256)
	assert.NotNil(t, fn)
	h1, err := EncodeToSHA256([]byte("test"))
	assert.NoError(t, err)
	h2, err := fn([]byte("test"))
	assert.NoError(t, err)
	assert.Equal(t, h1, h2)
}

func TestEncodeToSHA256(t *testing.T) {
	var tests = []struct {
		in  []byte
		out string
	}{
		{
			nil,
			"e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		},
		{
			[]byte(""),
			"e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		},
		{
			[]byte("abcdefghijklmnopqrstuvwxyz"),
			"71c480df93d6ae2f1efad1447c66c9525e316218cf51fc8d9ed832f2daf18b73",
		},
		{[]byte("abcdefghijklmnopqrstuvwxyz1234567890"),
			"77d721c817f9d216c1fb783bcad9cdc20aaa2427402683f1f75dd6dfbe657470",
		},
	}
	for _, test := range tests {
		h, err := EncodeToSHA256(test.in)
		assert.NoError(t, err)
		assert.Equal(t, test.out, h, test.in)
	}
}
