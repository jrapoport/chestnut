package value

// Keyed provides a management interface for keyed values.
type Keyed interface {
	// Key is the byte representation of the key.
	Key() []byte

	// Namespace is the namespace to use when storing the key.
	Namespace() string

	// ValidKey returns nil if the key is valid, otherwise ErrInvalidKey.
	ValidKey() error
}
