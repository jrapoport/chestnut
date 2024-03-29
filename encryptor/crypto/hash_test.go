package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashSHA256(t *testing.T) {
	var tests = []struct {
		in  string
		out []byte
	}{
		{
			"",
			[]byte{0xe3, 0xb0, 0xc4, 0x42, 0x98, 0xfc, 0x1c, 0x14, 0x9a, 0xfb, 0xf4, 0xc8,
				0x99, 0x6f, 0xb9, 0x24, 0x27, 0xae, 0x41, 0xe4, 0x64, 0x9b, 0x93, 0x4c, 0xa4,
				0x95, 0x99, 0x1b, 0x78, 0x52, 0xb8, 0x55},
		},
		{
			"abcdefghijklmnopqrstuvwxyz",
			[]byte{0x71, 0xc4, 0x80, 0xdf, 0x93, 0xd6, 0xae, 0x2f, 0x1e, 0xfa, 0xd1, 0x44,
				0x7c, 0x66, 0xc9, 0x52, 0x5e, 0x31, 0x62, 0x18, 0xcf, 0x51, 0xfc, 0x8d, 0x9e, 0xd8,
				0x32, 0xf2, 0xda, 0xf1, 0x8b, 0x73},
		},
		{
			"abcdefghijklmnopqrstuvwxyz1234567890",
			[]byte{0x77, 0xd7, 0x21, 0xc8, 0x17, 0xf9, 0xd2, 0x16, 0xc1, 0xfb, 0x78, 0x3b, 0xca,
				0xd9, 0xcd, 0xc2, 0xa, 0xaa, 0x24, 0x27, 0x40, 0x26, 0x83, 0xf1, 0xf7, 0x5d, 0xd6,
				0xdf, 0xbe, 0x65, 0x74, 0x70},
		},
	}
	for _, test := range tests {
		h, err := HashSHA256([]byte(test.in))
		assert.NoError(t, err)
		assert.Equal(t, test.out, h, test.in)
	}
}
