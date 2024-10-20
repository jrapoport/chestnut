package keystore

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"testing"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/ed25519"
)

func testPrivKeyToPrivateKey(t *testing.T, pk1 interface{}, conv func() interface{}) {
	assert.NotNil(t, pk1)
	stdKey := conv()
	assert.NotNil(t, stdKey)
	pk2, _, err := crypto.KeyPairFromStdKey(stdKey)
	assert.NoError(t, err)
	assert.Equal(t, pk1, pk2)
}

func testPrivateKeyToPrivKey(t *testing.T, pk1 interface{}, conv func() crypto.PrivKey) {
	assert.NotNil(t, pk1)
	privKey := conv()
	assert.NotNil(t, privKey)
	pk2, err := crypto.PrivKeyToStdKey(privKey)
	assert.NoError(t, err)
	assert.Equal(t, pk1, pk2)
}

func TestPrivKeyToRSAPrivateKey(t *testing.T) {
	privKey, _, err := crypto.GenerateRSAKeyPair(2048, rand.Reader)
	assert.NoError(t, err)
	testPrivKeyToPrivateKey(t, privKey, func() interface{} {
		return PrivKeyToRSAPrivateKey(privKey)
	})
	assert.Panics(t, func() {
		_ = PrivKeyToRSAPrivateKey(nil)
	})
}

func TestPrivKeyToECDSAPrivateKey(t *testing.T) {
	privKey, _, err := crypto.GenerateECDSAKeyPair(rand.Reader)
	assert.NoError(t, err)
	testPrivKeyToPrivateKey(t, privKey, func() interface{} {
		return PrivKeyToECDSAPrivateKey(privKey)
	})
	assert.Panics(t, func() {
		_ = PrivKeyToECDSAPrivateKey(nil)
	})
}

func TestPrivKeyToEd25519PrivateKey(t *testing.T) {
	privKey, _, err := crypto.GenerateEd25519Key(rand.Reader)
	assert.NoError(t, err)
	testPrivKeyToPrivateKey(t, privKey, func() interface{} {
		return PrivKeyToEd25519PrivateKey(privKey)
	})
	assert.Panics(t, func() {
		_ = PrivKeyToEd25519PrivateKey(nil)
	})
}

func TestPrivKeyToBTCECPrivateKey(t *testing.T) {
	privKey, _, err := crypto.GenerateSecp256k1Key(rand.Reader)
	assert.NoError(t, err)
	testPrivKeyToPrivateKey(t, privKey, func() interface{} {
		return PrivKeyToBTCECPrivateKey(privKey)
	})
	assert.Panics(t, func() {
		_ = PrivKeyToBTCECPrivateKey(nil)
	})
}

func TestRSAPrivateKeyToPrivKey(t *testing.T) {
	rsaKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(t, err)
	testPrivateKeyToPrivKey(t, rsaKey, func() crypto.PrivKey {
		return RSAPrivateKeyToPrivKey(rsaKey)
	})
	assert.Panics(t, func() {
		_ = RSAPrivateKeyToPrivKey(nil)
	})
}

func TestECDSAPrivateKeyToPrivKey(t *testing.T) {
	ecdsaKey, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	assert.NoError(t, err)
	testPrivateKeyToPrivKey(t, ecdsaKey, func() crypto.PrivKey {
		return ECDSAPrivateKeyToPrivKey(ecdsaKey)
	})
	assert.Panics(t, func() {
		_ = ECDSAPrivateKeyToPrivKey(nil)
	})
}

func TestEd25519PrivateKeyToPrivKey(t *testing.T) {
	_, edKey, err := ed25519.GenerateKey(rand.Reader)
	assert.NoError(t, err)
	testPrivateKeyToPrivKey(t, &edKey, func() crypto.PrivKey {
		return Ed25519PrivateKeyToPrivKey(&edKey)
	})
	assert.Panics(t, func() {
		_ = Ed25519PrivateKeyToPrivKey(nil)
	})
}

func TestBTCECPrivateKeyToPrivKey(t *testing.T) {
	btcecKey, err := btcec.NewPrivateKey()
	key := (*crypto.Secp256k1PrivateKey)(btcecKey)
	assert.NoError(t, err)
	testPrivateKeyToPrivKey(t, key, func() crypto.PrivKey {
		return BTCECPrivateKeyToPrivKey(btcecKey)
	})
	assert.Panics(t, func() {
		_ = BTCECPrivateKeyToPrivKey(nil)
	})
}
