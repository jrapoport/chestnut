package lookup

import (
	"fmt"
	"reflect"
	"testing"
	"unsafe"

	"github.com/jrapoport/chestnut/encoding/json/encoders"
	jsoniter "github.com/json-iterator/go"
	"github.com/modern-go/reflect2"
	"github.com/stretchr/testify/assert"
)

func TestLookupDecoder_Decode(t *testing.T) {
	type testObject struct {
		Value string
	}
	tests := []struct {
		value    interface{}
		key      string
		encoding string
	}{
		{
			"a-string-value",
			`"tst0xtesting%d_24"`,
			`"a-string-value"`,
		},
		{
			[]string{"a-string-slice"},
			`"tst0xtesting%d_23"`,
			`["a-string-slice"]`,
		},
		{
			99.9,
			`"tst0xtesting%d_14"`,
			`99.9`,
		},
		{
			testObject{"a-struct-value"},
			`"tst0xtesting%d_25"`,
			`{"Value":"a-struct-value"}`,
		},
		{
			&testObject{"a-struct-ptr-value"},
			`"tst0xtesting%d_22"`,
			`{"Value":"a-struct-ptr-value"}`,
		},
	}
	lookUpTable := "{"
	for i, test := range tests {
		key := fmt.Sprintf(test.key, i)
		if i > 0 {
			lookUpTable += ","
		}
		entry := fmt.Sprintf("%s:%s", key, test.encoding)
		lookUpTable += entry
	}
	lookUpTable += "}"
	ctx := &Context{
		NewLookupToken(testPrefix, testID),
		newTestStream(t),
	}
	enc := encoders.NewEncoder()
	ctx.Stream.Attachment = enc.Get([]byte(lookUpTable))
	for i, test := range tests {
		typ := reflect2.TypeOf(&test.value)
		decoder := enc.DecoderOf(typ)
		le := NewLookupDecoder(ctx, typ, decoder)
		key := fmt.Sprintf(test.key, i)
		iter := enc.BorrowIterator([]byte(key))
		ptr := reflect.New(reflect.TypeOf(test.value)).Interface()
		le.Decode(unsafe.Pointer(&ptr), iter)
		enc.ReturnIterator(iter)
		assert.Equal(t, test.encoding, string(ctx.Stream.Buffer()))
		any := jsoniter.Get(ctx.Stream.Buffer())
		assert.NotEqual(t, jsoniter.InvalidValue, any.ValueType())
	}
}

func TestLookupEncoder_NewLookupDecoder(t *testing.T) {
	encoder := encoders.NewEncoder()
	str := "a-string"
	typ := reflect2.TypeOf(&str)
	enc := encoder.DecoderOf(typ)
	bad1 := &Context{}
	bad2 := &Context{InvalidToken, newTestStream(t)}
	bad3 := &Context{"a-string-value",nil}
	good :=  &Context{"a-string-value", newTestStream(t)}
	for _, ctx := range []*Context {nil, bad1, bad2, bad3, good} {
		for _, tp := range []reflect2.Type{nil, typ} {
			for _, ve := range []jsoniter.ValDecoder{nil, enc} {
				if ctx == good && tp == typ && ve == enc {
					continue
				}
				assert.Panics(t, func() {
					_ = NewLookupDecoder(ctx, tp, ve)
				}, ctx, tp, enc)
			}
		}
	}
}
