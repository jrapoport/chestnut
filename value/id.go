package value

import "github.com/jrapoport/chestnut/storage"

// ID provides a implementation of the Keyed interface.
// It can be embedded in another structure to build custom Keyed values.
type ID struct {
	ID string `json:"id"`
}

var _ Keyed = (*ID)(nil)

// Key returns the key as bytes.
func (k *ID) Key() []byte {
	return []byte(k.ID)
}

// Namespace is the namespace to use when storing the key.
func (k *ID) Namespace() string {
	name := ""
	if k.ID != "" {
		name = k.ID[:1]
	}
	return name
}

// ValidKey checks the key is valid.
func (k *ID) ValidKey() error {
	return storage.ValidKey(k.Namespace(), k.Key())
}

// String returns the key as a string.
func (k *ID) String() string {
	return k.ID
}
