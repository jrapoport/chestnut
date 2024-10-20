package nuts

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/jrapoport/chestnut/log"
	"github.com/jrapoport/chestnut/storage"
	jsoniter "github.com/json-iterator/go"
	"github.com/nutsdb/nutsdb"
)

const logName = "nutsdb"

// nutsDBStore is an implementation the Storage interface for nutsdb
// https://github.com/nutsdb/nutsdb.
type nutsDBStore struct {
	opts storage.StoreOptions
	path string
	db   *nutsdb.DB
	log  log.Logger
}

var _ storage.Storage = (*nutsDBStore)(nil)

// NewStore is used to instantiate a datastore backed by nutsdb.
func NewStore(path string, opt ...storage.StoreOption) storage.Storage {
	opts := storage.ApplyOptions(storage.DefaultStoreOptions, opt...)
	logger := log.Named(opts.Logger(), logName)
	if path == "" {
		logger.Panic("store path required")
	}
	return &nutsDBStore{path: path, opts: opts, log: logger}
}

// Options returns the configuration options for the store.
func (s *nutsDBStore) Options() storage.StoreOptions {
	return s.opts
}

// Open opens the store.
func (s *nutsDBStore) Open() (err error) {
	s.log.Debugf("opening store at path: %s", s.path)
	opt := nutsdb.DefaultOptions
	opt.Dir = s.path
	if s.db, err = nutsdb.Open(opt); err != nil {
		err = s.logError("open", err)
		return
	}
	if s.db == nil {
		err = errors.New("unable to open backing store")
		err = s.logError("open", err)
		return
	}
	s.log.Infof("opened store at path: %s", s.path)
	return
}

// Put an entry in the store.
func (s *nutsDBStore) Put(name string, key []byte, value []byte) error {
	s.log.Debugf("put: %d value bytes to key: %s", len(value), key)
	if err := storage.ValidKey(name, key); err != nil {
		return s.logError("put", err)
	} else if len(value) <= 0 {
		err = errors.New("value cannot be empty")
		return s.logError("put", err)
	}
	newBucket := func(tx *nutsdb.Tx) error {
		e := tx.NewBucket(nutsdb.DataStructureBTree, name)
		if e != nil && !errors.Is(e, nutsdb.ErrBucketAlreadyExist) {
			return e
		}
		return nil
	}
	if err := s.db.Update(newBucket); err != nil {
		return s.logError("put", err)
	}
	putValue := func(tx *nutsdb.Tx) error {
		s.log.Debugf("put: tx %d bytes to key: %s.%s",
			len(value), name, string(key))
		return tx.Put(name, key, value, 0)
	}
	return s.logError("put", s.db.Update(putValue))
}

// Get a value from the store.
func (s *nutsDBStore) Get(name string, key []byte) ([]byte, error) {
	s.log.Debugf("get: value at key: %s", key)
	if err := storage.ValidKey(name, key); err != nil {
		return nil, s.logError("get", err)
	}
	var value []byte
	var err error
	getValue := func(tx *nutsdb.Tx) error {
		s.log.Debugf("get: tx key: %s.%s", name, key)
		value, err = tx.Get(name, key)
		if err != nil {
			return err
		}
		s.log.Debugf("get: tx key: %s.%s value (%d bytes)",
			name, string(key), len(value))
		return nil
	}
	if err := s.db.View(getValue); err != nil {
		return nil, s.logError("get", err)
	}
	return value, nil
}

// Save the value in v and store the result at key.
func (s *nutsDBStore) Save(name string, key []byte, v interface{}) error {
	b, err := jsoniter.Marshal(v)
	if err != nil {
		return s.logError("save", err)
	}
	return s.Put(name, key, b)
}

// Load the value at key and stores the result in v.
func (s *nutsDBStore) Load(name string, key []byte, v interface{}) error {
	b, err := s.Get(name, key)
	if err != nil {
		return s.logError("load", err)
	}
	return s.logError("load", jsoniter.Unmarshal(b, v))
}

// Has checks for a key in the store.
func (s *nutsDBStore) Has(name string, key []byte) (bool, error) {
	s.log.Debugf("has: key: %s", key)
	if err := storage.ValidKey(name, key); err != nil {
		return false, s.logError("has", err)
	}
	var has bool
	hasKey := func(tx *nutsdb.Tx) error {
		s.log.Debugf("has: tx get namespace: %s", name)
		keys, err := tx.GetKeys(name)
		if err != nil {
			return err
		}
		s.log.Debugf("has: tx found %d keys in: %s", len(keys), name)
		for _, k := range keys {
			has = bytes.Equal(key, k)
			if has {
				s.log.Debugf("has: tx key found: %s.%s", name, string(key))
				break
			}
		}
		return nil
	}
	if err := s.db.View(hasKey); err != nil {
		return false, s.logError("has", err)
	}
	s.log.Debugf("has: found key %s: %t", key, has)
	return has, nil
}

// Delete removes a key from the store.
func (s *nutsDBStore) Delete(name string, key []byte) error {
	s.log.Debugf("delete: key: %s", key)
	if err := storage.ValidKey(name, key); err != nil {
		return s.logError("delete", err)
	}
	del := func(tx *nutsdb.Tx) error {
		s.log.Debugf("delete: tx key: %s.%s", name, string(key))
		err := tx.Delete(name, key)
		if errors.Is(err, nutsdb.ErrKeyNotFound) {
			return nil
		}
		return err
	}
	return s.logError("delete", s.db.Update(del))
}

// List returns a list of all keys in the namespace.
func (s *nutsDBStore) List(name string) (keys [][]byte, err error) {
	s.log.Debugf("list: keys in namespace: %s", name)
	listKeys := func(tx *nutsdb.Tx) error {
		keys, err = s.listKeys(name, tx)
		return err
	}
	if err = s.db.View(listKeys); err != nil {
		return nil, s.logError("list", err)
	}
	s.log.Debugf("list: found %d keys: %s", len(keys), keys)
	return
}

func (s *nutsDBStore) listKeys(name string, tx *nutsdb.Tx) ([][]byte, error) {
	var keys [][]byte
	s.log.Debugf("list: tx scan namespace: %s", name)
	keys, err := tx.GetKeys(name)
	if err != nil {
		return nil, err
	}
	s.log.Debugf("list: tx found %d keys in: %s", len(keys), name)
	// for _, key := range keys {
	//	s.log.Debugf("list: tx found key: %s.%s", name, key)
	// }
	return keys, nil
}

// ListAll returns a mapped list of all keys in the store.
func (s *nutsDBStore) ListAll() (map[string][][]byte, error) {
	s.log.Debugf("list: all keys")
	var total int
	allKeys := map[string][][]byte{}
	listKeys := func(tx *nutsdb.Tx) error {
		err := tx.IterateBuckets(nutsdb.DataStructureBTree, "*", func(bucket string) bool {
			keys, err := s.listKeys(bucket, tx)
			if err != nil {
				return false
			}
			if len(keys) <= 0 {
				return true
			}
			allKeys[bucket] = keys
			total += len(keys)
			return true
		})
		return err
	}
	if err := s.db.View(listKeys); err != nil {
		return nil, s.logError("list", err)
	}
	s.log.Debugf("list: found %d keys: %s", total, allKeys)
	return allKeys, nil
}

// Export copies the datastore to directory at path.
func (s *nutsDBStore) Export(path string) error {
	s.log.Debugf("export: to path: %s", path)
	if path == "" {
		err := fmt.Errorf("invalid path: %s", path)
		return s.logError("export", err)
	} else if s.path == path {
		err := fmt.Errorf("path cannot be store path: %s", path)
		return s.logError("export", err)
	}
	if err := s.db.Backup(path); err != nil {
		return s.logError("export", err)
	}
	s.log.Debugf("export: to path complete: %s", path)
	return nil
}

// Close closes the datastore and releases all db resources.
func (s *nutsDBStore) Close() error {
	s.log.Debugf("closing store at path: %s", s.path)
	err := s.db.Close()
	s.db = nil
	s.log.Info("store closed")
	return s.logError("close", err)
}

func (s *nutsDBStore) logError(name string, err error) error {
	if err == nil {
		return nil
	}
	if name != "" {
		err = fmt.Errorf("%s: %w", name, err)
	}
	s.log.Error(err)
	return err
}
