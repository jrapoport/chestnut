package aes

import "testing"

func TestCipherGCM(t *testing.T) {
	testCipher(t, EncryptGCM, DecryptGCM)
}
