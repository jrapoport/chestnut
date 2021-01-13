package bolt

import (
	"testing"

	"github.com/jrapoport/chestnut/storage/store_test"
)

func TestStore(t *testing.T) {
	store_test.TestStore(t, NewStore)
}
