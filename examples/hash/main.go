package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/jrapoport/chestnut"
	"github.com/jrapoport/chestnut/encryptor/aes"
	"github.com/jrapoport/chestnut/encryptor/crypto"
	"github.com/jrapoport/chestnut/storage/nuts"
)

func main() {
	path := filepath.Join(os.TempDir(), "hash")
	defer os.RemoveAll(path)

	// use nutsdb
	store := nuts.NewStore(path)

	// use a simple text secret
	textSecret := crypto.TextSecret("i-am-a-good-secret")

	// use AES256-CFB encryption
	opt := chestnut.WithAES(crypto.Key256, aes.CFB, textSecret)

	// open the storage chest with nutsdb and the aes encryptor
	cn := chestnut.NewChestnut(store, opt)
	if err := cn.Open(); err != nil {
		log.Panic(err)
	}

	// define an struct with a hash field
	type HashValue struct {
		// ClearString will not be hashed
		ClearString string
		// HashString with the 'hash' tag option
		HashString string `json:",hash"`
	}

	src := &HashValue{
		ClearString: "I am a string",
		HashString:  "I will be hashed",
	}

	// a key for the value
	namespace := "sparse-values"
	key := []byte("sparse-value-id")

	// save the struct with sparse encryption
	if err := cn.Save(namespace, key, src); err != nil {
		log.Panic(err)
	}

	// load the value
	err := cn.Load(namespace, key, src)
	if err != nil {
		log.Panic(err)
	}

	fmt.Println("clear field:", src.ClearString)
	fmt.Println("hashed field:", src.HashString)
}
