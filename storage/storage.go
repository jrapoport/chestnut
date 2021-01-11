package storage

import (
	"errors"
	"fmt"
)

// Storage provides a management interface for a datastore.
type Storage interface {
	// Open opens the store.
	Open() error

	// Put a value in the store.
	Put(namespace string, key []byte, value []byte) error

	// Get a value from the store.
	Get(namespace string, key []byte) (value []byte, err error)

	// Has checks for a key in the store.
	Has(namespace string, key []byte) (bool, error)

	// Save the value in v and stores the result at key.
	Save(namespace string, key []byte, v interface{}) error

	// Load the value at key and stores the result in v.
	Load(namespace string, key []byte, v interface{}) error

	// List returns a list of all keys in the namespace.
	List(namespace string) ([][]byte, error)

	// Delete removes a key from the store.
	Delete(name string, key []byte) error

	// Close closes the store.
	Close() error

	// Export saves the store to path.
	Export(path string) error
}

// ErrInvalidKey the storage key is invalid.
var ErrInvalidKey = errors.New("invalid storage key")

// ValidKey returns nil if the key is valid, otherwise ErrInvalidKey.
func ValidKey(name string, key []byte) error {
	if name == "" {
		return fmt.Errorf("%w namespace: %s", ErrInvalidKey, name)
	}
	if len(key) <= 0 {
		return fmt.Errorf("%w: %s", ErrInvalidKey, key)
	}
	return nil
}
