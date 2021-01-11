package aes

import "testing"

func TestCipherCFB(t *testing.T) {
	testCipher(t, EncryptCFB, DecryptCFB)
}
