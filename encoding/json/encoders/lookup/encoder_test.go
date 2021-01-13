package lookup

import (
	"fmt"
	"github.com/jrapoport/chestnut/log"
	jsoniter "github.com/json-iterator/go"
	"testing"

	"github.com/jrapoport/chestnut/encoding/json/encoders"
	"github.com/modern-go/reflect2"
	"github.com/stretchr/testify/assert"
)

func TestLookupEncoder_Encode(t *testing.T) {
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
	encoded := ""
	lookup := ""
	stream := newTestStream(t)
	ctx := &Context{
		NewLookupToken(testPrefix, testID),
		newTestStream(t),
	}
	enc := encoders.NewEncoder()
	for i, test := range tests {
		typ := reflect2.TypeOf(test.value)
		encoder := enc.EncoderOf(typ)
		le := NewLookupEncoder(ctx, typ, encoder)
		le.Encode(reflect2.PtrOf(test.value), stream)
		key := fmt.Sprintf(test.key, i)
		encoded += key
		assert.Equal(t, encoded, string(stream.Buffer()))
		if i > 0 {
			lookup += ","
		}
		entry := fmt.Sprintf("%s:%s", key, test.encoding)
		lookup += entry
		assert.Equal(t, lookup, string(ctx.Stream.Buffer()))
	}
}

func TestLookupEncoder_IsEmpty(t *testing.T) {
	tests := []struct {
		value       interface{}
		assertEmpty assert.BoolAssertionFunc
	}{
		{"", assert.True},
		{"not-empty", assert.False},
		{[]string{}, assert.True},
		{[]string{"not-empty"}, assert.False},
	}
	encoder := encoders.NewEncoder()
	for _, test := range tests {
		enc := encoder.EncoderOf(reflect2.TypeOf(test.value))
		le := &Encoder{encoder: enc}
		empty := le.IsEmpty(reflect2.PtrOf(test.value))
		test.assertEmpty(t, empty, "value: %v", test.value)
	}
}

func TestLookupEncoder_NewLookupEncoder(t *testing.T) {
	encoder := encoders.NewEncoder()
	typ := reflect2.TypeOf("a-string")
	enc := encoder.EncoderOf(typ)
	bad1 := &Context{}
	bad2 := &Context{InvalidToken, newTestStream(t)}
	bad3 := &Context{"a-string-value",nil}
	good :=  &Context{"a-string-value", newTestStream(t)}
	for _, ctx := range []*Context {nil, bad1, bad2, bad3, good} {
		for _, tp := range []reflect2.Type{nil, typ} {
			for _, ve := range []jsoniter.ValEncoder{nil, enc} {
				if ctx == good && tp == typ && ve == enc {
					continue
				}
				assert.Panics(t, func() {
					_ = NewLookupEncoder(ctx, tp, ve)
				}, ctx, tp, enc)
			}
		}
	}
}

func TestLookupEncoder_Fallback(t *testing.T) {
	strVal := "not-empty"
	stream := newTestStream(t)
	encoder := encoders.NewEncoder()
	kty := reflect2.TypeOf("a-string")
	enc := encoder.EncoderOf(kty)
	le := &Encoder{stream: stream, valType: kty, encoder: enc, log: log.Log}
	le.Encode(reflect2.PtrOf(strVal), stream)
}