package value

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ID(t *testing.T) {
	keyTest := []struct {
		name string
		key  string
		err  assert.ErrorAssertionFunc
	}{
		{"", "", assert.Error},
		{"a", "a", assert.NoError},
		{"t", "test", assert.NoError},
	}
	for _, test := range keyTest {
		key := &ID{test.key}
		assert.Equal(t, test.name, key.Namespace())
		assert.Equal(t, []byte(test.key), key.Key())
		assert.Equal(t, test.key, key.String())
		test.err(t, key.ValidKey())
	}
}
