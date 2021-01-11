package zstd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	testNil   = []byte(nil)
	testEmpty = []byte("")
	testSpace = []byte(" ")
	testValue = []byte("i-am-an-uncompressed-string")
	compSpace = []byte{0x28, 0xb5, 0x2f, 0xfd, 0x4, 0x0, 0x9, 0x0, 0x0, 0x20, 0x8d, 0x63, 0x68, 0xb6}
	compValue = []byte{0x28, 0xb5, 0x2f, 0xfd, 0x4, 0x0, 0xd9, 0x0, 0x0, 0x69, 0x2d, 0x61, 0x6d, 0x2d,
		0x61, 0x6e, 0x2d, 0x75, 0x6e, 0x63, 0x6f, 0x6d, 0x70, 0x72, 0x65, 0x73, 0x73, 0x65, 0x64, 0x2d,
		0x73, 0x74, 0x72, 0x69, 0x6e, 0x67, 0xab, 0x52, 0xd3, 0x9d}
)

func TestCompressZStd(t *testing.T) {
	tests := []struct {
		src []byte
		out []byte
		err assert.ErrorAssertionFunc
	}{
		{testNil, testEmpty, assert.NoError},
		{testEmpty, testEmpty, assert.NoError},
		{testSpace, compSpace, assert.NoError},
		{testValue, compValue, assert.NoError},
	}
	for _, test := range tests {
		bytes, err := Compress(test.src)
		test.err(t, err)
		assert.Equal(t, test.out, bytes)
	}
}

func TestDecompressZStd(t *testing.T) {
	tests := []struct {
		src []byte
		out []byte
		err assert.ErrorAssertionFunc
	}{
		{testNil, testNil, assert.NoError},
		{testEmpty, testNil, assert.NoError},
		{testValue, testNil, assert.Error},
		{compSpace, testSpace, assert.NoError},
		{compValue, testValue, assert.NoError},
	}
	for _, test := range tests {
		bytes, err := Decompress(test.src)
		test.err(t, err)
		assert.Equal(t, test.out, bytes)
	}
}

func TestZStd(t *testing.T) {
	// compress the src
	buf, err := Compress(testValue)
	assert.NoError(t, err)
	assert.NotNil(t, buf)
	// decompress the result
	src, err := Decompress(buf)
	assert.NoError(t, err)
	assert.Equal(t, testValue, src)
}
