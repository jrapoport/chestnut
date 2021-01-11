package aes

import "testing"

func TestCipherCTR(t *testing.T) {
	testCipher(t, EncryptCTR, DecryptCTR)
}
