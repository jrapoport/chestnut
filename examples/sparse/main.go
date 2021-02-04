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
	path := filepath.Join(os.TempDir(), "sparse")
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

	// define a sparse struct with a secure field
	type Sparse struct {
		// SecretString with the 'secure' tag option
		SecretString string `json:",secure"`
		// PublicString will not be encrypted
		PublicString string
	}

	src := &Sparse{
		SecretString: "I am secret",
		PublicString: "I am visible",
	}

	// a key for the value
	namespace := "sparse-values"
	key := []byte("sparse-value-id")

	// save the value with sparse encryption
	if err := cn.Save(namespace, key, src); err != nil {
		log.Panic(err)
	}

	sparse := &Sparse{}

	// load a sparse copy
	err := cn.Sparse(namespace, key, sparse)
	if err != nil {
		log.Panic(err)
	}

	fmt.Println("secure field:", sparse.SecretString)
	fmt.Println("public field:", sparse.PublicString)
}
