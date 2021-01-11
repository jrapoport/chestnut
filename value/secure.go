package value

// Secure provides a simple value for storing sparsely
// encrypted blobs and plaintext metadata.
type Secure struct {
	ID
	Data     []byte                 `json:"data,secure"`
	Metadata map[string]interface{} `json:"metadata"`
}

// NewSecureValue returns a new Secure value.
func NewSecureValue(id string, data []byte) *Secure {
	return &Secure{
		ID:       ID{ID: id},
		Data:     data,
		Metadata: map[string]interface{}{},
	}
}

// SetMetadata sets the metadata entry for k to v.
func (e *Secure) SetMetadata(k string, v interface{}) {
	e.Metadata[k] = v
}

// GetMetadata gets the metadata entry for k and returns it as v.
func (e *Secure) GetMetadata(k string) (v interface{}) {
	v = e.Metadata[k]
	return
}
