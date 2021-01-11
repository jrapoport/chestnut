package hash

import (
	"bytes"
	"reflect"
	"testing"
	"unsafe"

	"github.com/jrapoport/chestnut/encoding/tags"
	jsoniter "github.com/json-iterator/go"
	"github.com/modern-go/reflect2"
	"github.com/stretchr/testify/assert"
)

func TestHashEncoder(t *testing.T) {
	var tests = []struct {
		in          []byte
		out         string
		assertEmpty assert.BoolAssertionFunc
	}{
		{
			nil,
			`""`,
			assert.True,
		},
		{
			[]byte(""),
			`""`,
			assert.True,
		},
		{
			[]byte("abcdefghijklmnopqrstuvwxyz"),
			`"sha256:71c480df93d6ae2f1efad1447c66c9525e316218cf51fc8d9ed832f2daf18b73"`,
			assert.False,
		},
		{
			[]byte("abcdefghijklmnopqrstuvwxyz1234567890"),
			`"sha256:77d721c817f9d216c1fb783bcad9cdc20aaa2427402683f1f75dd6dfbe657470"`,
			assert.False,
		},
	}
	for _, test := range tests {
		var buf bytes.Buffer
		conf := jsoniter.ConfigDefault
		valEncoder := conf.EncoderOf(reflect2.DefaultTypeOfKind(reflect.String))
		stream := jsoniter.NewStream(conf, &buf, 100)
		stream.Reset(&buf)
		he := NewHashEncoder(tags.HashSHA256, EncodeToSHA256, valEncoder)
		he.Encode(unsafe.Pointer(&test.in), stream)
		assert.Equal(t, test.out, string(stream.Buffer()))
		test.assertEmpty(t, he.IsEmpty(unsafe.Pointer(&test.in)))
	}
}

func TestHashEncoder_NoRehash(t *testing.T) {
	var testIn = "sha256:71c480df93d6ae2f1efad1447c66c9525e316218cf51fc8d9ed832f2daf18b73"
	const testOut = `"sha256:71c480df93d6ae2f1efad1447c66c9525e316218cf51fc8d9ed832f2daf18b73"`
	var buf bytes.Buffer
	conf := jsoniter.ConfigDefault
	valEncoder := conf.EncoderOf(reflect2.DefaultTypeOfKind(reflect.String))
	stream := jsoniter.NewStream(conf, &buf, 100)
	stream.Reset(&buf)
	he := NewHashEncoder(tags.HashSHA256, EncodeToSHA256, valEncoder)
	he.Encode(unsafe.Pointer(&testIn), stream)
	assert.Equal(t, testOut, string(stream.Buffer()))
}
