package encryptor

import (
	"strings"
	"testing"

	"github.com/jrapoport/chestnut/encryptor/aes"
	"github.com/jrapoport/chestnut/encryptor/crypto"
	"github.com/stretchr/testify/assert"
)

func TestChainEncryptor_Nil(t *testing.T) {
	assert.Nil(t, NewChainEncryptor())
}

func TestChainEncryptor_Single(t *testing.T) {
	ae := NewAESEncryptor(crypto.Key128, aes.CFB, textSecret)
	assert.NotNil(t, ae)
	chain := NewChainEncryptor(ae)
	assert.Equal(t, ae.Name(), chain.Name())
	assert.Equal(t, ae.ID(), chain.ID())
	testChainEncryptor(t, chain)
}

func TestChainEncryptor_Chained(t *testing.T) {
	encryptors := []crypto.Encryptor{
		&AESEncryptor{textSecret, crypto.Key128, aes.CFB},
		&AESEncryptor{managedSecret, crypto.Key192, aes.CTR},
		&AESEncryptor{secureSecret, crypto.Key256, aes.GCM},
	}
	chain := NewChainEncryptor(encryptors...)
	testChainName(t, chain, encryptors)
	testChainID(t, chain, encryptors)
	testChainEncryptor(t, chain)
}

func testChainName(t *testing.T, chain *ChainEncryptor, encryptors []crypto.Encryptor) {
	var names []string
	for _, e := range encryptors {
		names = append(names, e.Name())
	}
	name := strings.Join(names, chainSep)
	assert.Equal(t, name, chain.Name())
}

func testChainID(t *testing.T, chain *ChainEncryptor, encryptors []crypto.Encryptor) {
	var ids []string
	for _, e := range encryptors {
		ids = append(ids, e.ID())
	}
	id := strings.Join(ids, chainSep)
	assert.Equal(t, id, chain.ID())
}

func testChainEncryptor(t *testing.T, chain *ChainEncryptor) {
	assert.NotNil(t, chain)
	e, err := chain.Encrypt([]byte(testPlainText))
	assert.NoError(t, err)
	assert.NotEmpty(t, e)
	d, err := chain.Decrypt(e)
	assert.NoError(t, err)
	assert.NotEmpty(t, d)
	assert.Equal(t, testPlainText, string(d))
}
