package crypto

// Secret is the interface that wraps a cipher keyLen and its id.
type Secret interface {
	// ID return the id of the secret for tracking, or rollover etc.
	ID() string
	// Open returns a byte representation of the secret for encryption and decryption.
	Open() []byte
}

// A TextSecret provides a simple plaintext secret.
type TextSecret string

var _ Secret = (*TextSecret)(nil)

// ID return the id of the secret for tracking, or rollover etc.
func (s TextSecret) ID() string {
	return "text"
}

// Open returns a byte representation of the secret for encryption and decryption.
func (s TextSecret) Open() []byte {
	return []byte(s)
}

// A ManagedSecret provides a simple plaintext secret alongside a unique id.
type ManagedSecret struct {
	id string
	TextSecret
}

var _ Secret = (*ManagedSecret)(nil)

// NewManagedSecret creates a new ManagedSecret with a secret with its corresponding id.
func NewManagedSecret(id, secret string) *ManagedSecret {
	return &ManagedSecret{id, TextSecret(secret)}
}

// ID return the id of the secret for tracking, or rollover etc.
func (s ManagedSecret) ID() string {
	return s.id
}

// SecureSecret provides a unique id for a secret alongside an openSecret callback which
// returns a byte representation of the secret for encryption and decryption on Open.
// When SecureSecret calls openSecret it will pass a copy of itself as a Secret. This allows
// for remote loading of the secret based on its id, or using a secure in-memory storage
// solution for the secret like memguarded (https://github.com/n0rad/memguarded).
type SecureSecret struct {
	id   string
	open func(Secret) []byte
}

var _ Secret = (*SecureSecret)(nil)

// NewSecureSecret creates a new SecureSecret with an id and an callback function which
// returns a byte representation of the secret for encryption and decryption.
func NewSecureSecret(id string, openSecret func(Secret) []byte) *SecureSecret {
	return &SecureSecret{id, openSecret}
}

// ID return the id of the secret for tracking, or rollover etc.
func (s SecureSecret) ID() string {
	return s.id
}

// Open returns a byte representation of the secret for encryption and decryption.
func (s SecureSecret) Open() []byte {
	return s.open(s)
}
