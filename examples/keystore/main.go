package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/jrapoport/chestnut"
	"github.com/jrapoport/chestnut/encryptor/aes"
	"github.com/jrapoport/chestnut/encryptor/crypto"
	"github.com/jrapoport/chestnut/keystore"
	"github.com/jrapoport/chestnut/storage/nuts"
)

func main() {

	path := filepath.Join(os.TempDir(), "keystore")
	defer os.RemoveAll(path)

	// use nutsdb
	store := nuts.NewStore(path)

	// use a simple text secret
	textSecret := crypto.TextSecret("i-am-a-good-secret")

	opts := []chestnut.ChestOption{
		// use AES256-CFB encryption
		chestnut.WithAES(crypto.Key256, aes.CFB, textSecret),
	}

	// open the keystore with nutsdb and the aes encryptor
	ks := keystore.NewKeystore(store, opts...)
	if err := ks.Open(); err != nil {
		log.Panic(err)
	}

	// generate a new *btcec.PrivateKey
	pk1, err := btcec.NewPrivateKey()
	if err != nil {
		log.Panic(err)
	}

	// convert pk from *btcec.PrivateKey to ci.PrivKey.
	privKey1 := keystore.BTCECPrivateKeyToPrivKey(pk1)

	// encrypt the private key and put in the keystore
	if err = ks.Put("my private key", privKey1); err != nil {
		log.Panic(err)
	}

	// get the private key from the store and decrypt it
	privKey2, err := ks.Get("my private key")
	if err != nil {
		log.Panic(err)
	}

	// convert the saved private key to *btcec.PrivateKey
	pk2 := keystore.PrivKeyToBTCECPrivateKey(privKey2)

	// compare the keys
	if bytes.Equal(pk1.Serialize(), pk2.Serialize()) {
		fmt.Println("private keys are equal")
	}
}
