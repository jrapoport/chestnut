package crypto

import "crypto/sha256"

// HashSHA256 returns a sha256 hash of data.
func HashSHA256(data []byte) ([]byte, error) {
	h := sha256.New()
	if _, err := h.Write(data); err != nil {
		return nil, err
	}
	return h.Sum(nil), nil
}
