package keystore

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"log"

	"github.com/btcsuite/btcd/btcec"
	"github.com/libp2p/go-libp2p-core/crypto"
	"golang.org/x/crypto/ed25519"
)

// PrivKeyToRSAPrivateKey converts libp2p/go-libp2p-core/crypto
// private keys to standard library rsa private keys.
func PrivKeyToRSAPrivateKey(privKey crypto.PrivKey) *rsa.PrivateKey {
	key, err := crypto.PrivKeyToStdKey(privKey)
	if err != nil {
		log.Panic(err)
		return nil
	}
	if pk, ok := key.(*rsa.PrivateKey); ok {
		return pk
	}
	return nil
}

// RSAPrivateKeyToPrivKey converts standard library rsa
// private keys to libp2p/go-libp2p-core/crypto private keys.
func RSAPrivateKeyToPrivKey(privateKey *rsa.PrivateKey) crypto.PrivKey {
	// because we are strongly typing the interface it will never fail
	pk, _, _ := crypto.KeyPairFromStdKey(privateKey)
	return pk
}

// PrivKeyToECDSAPrivateKey converts libp2p/go-libp2p-core/crypto
// private keys to new standard library ecdsa private keys.
func PrivKeyToECDSAPrivateKey(privKey crypto.PrivKey) *ecdsa.PrivateKey {
	key, err := crypto.PrivKeyToStdKey(privKey)
	if err != nil {
		log.Panic(err)
		return nil
	}
	if pk, ok := key.(*ecdsa.PrivateKey); ok {
		return pk
	}
	return nil
}

// ECDSAPrivateKeyToPrivKey converts standard library ecdsa
// private keys to libp2p/go-libp2p-core/crypto private keys.
func ECDSAPrivateKeyToPrivKey(privateKey *ecdsa.PrivateKey) crypto.PrivKey {
	// because we are strongly typing the interface it will never fail
	pk, _, _ := crypto.KeyPairFromStdKey(privateKey)
	return pk
}

// PrivKeyToEd25519PrivateKey converts libp2p/go-libp2p-core/crypto
// private keys to ed25519 private keys.
func PrivKeyToEd25519PrivateKey(privKey crypto.PrivKey) *ed25519.PrivateKey {
	key, err := crypto.PrivKeyToStdKey(privKey)
	if err != nil {
		log.Panic(err)
		return nil
	}
	if pk, ok := key.(*ed25519.PrivateKey); ok {
		return pk
	}
	return nil
}

// Ed25519PrivateKeyToPrivKey converts ed25519 private keys
// to libp2p/go-libp2p-core/crypto private keys.
func Ed25519PrivateKeyToPrivKey(privateKey *ed25519.PrivateKey) crypto.PrivKey {
	// because we are strongly typing the interface it will never fail
	pk, _, _ := crypto.KeyPairFromStdKey(privateKey)
	return pk
}

// PrivKeyToBTCECPrivateKey converts libp2p/go-libp2p-core/crypto
// private keys to standard library btcec (and secp256k1) private keys.
// Internally equivalent to (*btcec.PrivateKey)(privKey.(*crypto.Secp256k1PrivateKey)).
func PrivKeyToBTCECPrivateKey(privKey crypto.PrivKey) *btcec.PrivateKey {
	key, err := crypto.PrivKeyToStdKey(privKey)
	if err != nil {
		log.Panic(err)
		return nil
	}
	if pk, ok := key.(*crypto.Secp256k1PrivateKey); ok {
		return (*btcec.PrivateKey)(pk)
	}
	return nil
}

// BTCECPrivateKeyToPrivKey converts standard library btcec (and secp256k1)
// private keys to libp2p/go-libp2p-core/crypto private keys. Internally
// equivalent to (*crypto.Secp256k1PrivateKey)(privateKey).
func BTCECPrivateKeyToPrivKey(privateKey *btcec.PrivateKey) crypto.PrivKey {
	// because we are strongly typing the interface it will never fail
	pk, _, _ := crypto.KeyPairFromStdKey(privateKey)
	return pk
}
