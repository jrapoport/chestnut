package keystore

import (
	"errors"

	"github.com/ipfs/go-ipfs/keystore"
	"github.com/jrapoport/chestnut"
	"github.com/jrapoport/chestnut/log"
	"github.com/jrapoport/chestnut/storage"
	ci "github.com/libp2p/go-libp2p-core/crypto"
)

const (
	namespace = "keys"
	logName   = "keystore"
)

// Keystore is used to manage an encrypted IPFS-compliant keystore.
type Keystore struct {
	cn    *chestnut.Chestnut
	store storage.Storage
	log   log.Logger
}

var _ keystore.Keystore = (*Keystore)(nil)

// NewKeystore is used to create a new chestnut ipfs-compliant keystore.
// Suggest using using this with AES256-CTR encryption based in part
// on this helpful analysis from Shawn Wang, PostgreSQL Database Core:
// https://www.highgo.ca/2019/08/08/the-difference-in-five-modes-in-the-aes-encryption-algorithm/
func NewKeystore(store storage.Storage, opt ...chestnut.ChestOption) *Keystore {
	// keystore requires that overwrites are forbidden
	opt = append(opt, chestnut.OverwritesForbidden())
	cn := chestnut.NewChestnut(store, opt...)
	logger := log.Named(cn.Logger(), logName)
	ks := &Keystore{cn, store, logger}
	if err := ks.validConfig(); err != nil {
		logger.Panic(err)
		return nil
	}
	return ks
}

func (ks *Keystore) validConfig() error {
	if ks.store == nil {
		return errors.New("store required")
	}
	return nil
}

// Open the Keystore
func (ks *Keystore) Open() error {
	if err := ks.validConfig(); err != nil {
		return err
	}
	return ks.cn.Open()
}

// Has returns whether or not a key exists in the Keystore
func (ks *Keystore) Has(s string) (bool, error) {
	return ks.cn.Has(namespace, []byte(s))
}

// Put stores a key in the Keystore, if a key with
// the same name already exists, returns ErrKeyExists
func (ks *Keystore) Put(s string, key ci.PrivKey) error {
	if key == nil {
		return errors.New("invalid key")
	}
	data, err := ci.MarshalPrivateKey(key)
	if err != nil {
		return err
	}
	err = ks.cn.Put(namespace, []byte(s), data)
	if errors.Is(err, chestnut.ErrForbidden) {
		return keystore.ErrKeyExists
	}
	return err
}

// Get retrieves a key from the Keystore if it
// exists, and returns ErrNoSuchKey otherwise.
func (ks *Keystore) Get(s string) (ci.PrivKey, error) {
	data, err := ks.cn.Get(namespace, []byte(s))
	if err != nil {
		return nil, keystore.ErrNoSuchKey
	}
	return ci.UnmarshalPrivateKey(data)
}

// Delete removes a key from the Keystore
func (ks *Keystore) Delete(s string) error {
	return ks.cn.Delete(namespace, []byte(s))
}

// List returns a list of key identifier
func (ks *Keystore) List() ([]string, error) {
	list, err := ks.cn.List(namespace)
	if err != nil {
		return nil, err
	}
	keys := make([]string, len(list))
	for i, key := range list {
		keys[i] = string(key)
	}
	return keys, nil
}

// Export the Keystore
func (ks *Keystore) Export(path string) error {
	return ks.cn.Export(path)
}

// Close the Keystore
func (ks *Keystore) Close() error {
	return ks.cn.Open()
}
