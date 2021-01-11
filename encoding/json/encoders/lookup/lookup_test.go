package lookup

import (
	"bytes"
	"testing"

	jsoniter "github.com/json-iterator/go"
	"github.com/stretchr/testify/assert"
)

const testID = "0xtesting"
const testPrefix = "tst"

func newTestStream(t *testing.T) *jsoniter.Stream {
	var buf bytes.Buffer
	conf := jsoniter.ConfigDefault
	stream := jsoniter.NewStream(conf, &buf, 4096)
	stream.Reset(&buf)
	assert.NotNil(t, stream)
	return stream
}
