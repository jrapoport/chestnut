package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMakeRand(t *testing.T) {
	const testLength = 20
	buf, err := MakeRand(0)
	assert.NoError(t, err)
	assert.Len(t, buf, 0)
	buf, err = MakeRand(testLength)
	assert.NoError(t, err)
	assert.Len(t, buf, testLength)
}
