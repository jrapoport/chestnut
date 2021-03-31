package bolt

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jrapoport/chestnut/log"
	"github.com/jrapoport/chestnut/storage"
	jsoniter "github.com/json-iterator/go"
	bolt "go.etcd.io/bbolt"
)

const (
	logName   = "bolt"
	storeName = "chest.db"
	storeExt  = ".db"
)

// boltStore is an implementation the Storage interface for bbolt
// https://github.com/etcd-io/bbolt.
type boltStore struct {
	opts storage.StoreOptions
	path string
	db   *bolt.DB
	log  log.Logger
}

var _ storage.Storage = (*boltStore)(nil)

// NewStore is used to instantiate a datastore backed by bbolt.
func NewStore(path string, opt ...storage.StoreOption) storage.Storage {
	opts := storage.ApplyOptions(storage.DefaultStoreOptions, opt...)
	logger := log.Named(opts.Logger(), logName)
	if path == "" {
		logger.Panic("store path required")
	}
	return &boltStore{path: path, opts: opts, log: logger}
}

// Options returns the configuration options for the store.
func (s *boltStore) Options() storage.StoreOptions {
	return s.opts
}

// Open opens the store.
func (s *boltStore) Open() (err error) {
	s.log.Debugf("opening store at path: %s", s.path)
	var path string
	path, err = ensureDBPath(s.path)
	if err != nil {
		err = s.logError("open", err)
		return
	}
	s.db, err = bolt.Open(path, 0600, nil)
	if err != nil {
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
func (s *boltStore) Put(name string, key []byte, value []byte) error {
	s.log.Debugf("put: %d value bytes to key: %s", len(value), key)
	if err := storage.ValidKey(name, key); err != nil {
		return s.logError("put", err)
	} else if len(value) <= 0 {
		err = errors.New("value cannot be empty")
		return s.logError("put", err)
	}
	putValue := func(tx *bolt.Tx) error {
		s.log.Debugf("put: tx %d bytes to key: %s.%s",
			len(value), name, string(key))
		b, err := tx.CreateBucketIfNotExists([]byte(name))
		if err != nil {
			return err
		}
		return b.Put(key, value)
	}
	return s.logError("put", s.db.Update(putValue))
}

// Get a value from the store.
func (s *boltStore) Get(name string, key []byte) ([]byte, error) {
	s.log.Debugf("get: value at key: %s", key)
	if err := storage.ValidKey(name, key); err != nil {
		return nil, s.logError("get", err)
	}
	var value []byte
	getValue := func(tx *bolt.Tx) error {
		s.log.Debugf("get: tx key: %s.%s", name, key)
		b := tx.Bucket([]byte(name))
		if b == nil {
			return fmt.Errorf("bucket not found: %s", name)
		}
		v := b.Get(key)
		if len(v) <= 0 {
			return errors.New("nil value")
		}
		value = v
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
func (s *boltStore) Save(name string, key []byte, v interface{}) error {
	b, err := jsoniter.Marshal(v)
	if err != nil {
		return s.logError("save", err)
	}
	return s.Put(name, key, b)
}

// Load the value at key and stores the result in v.
func (s *boltStore) Load(name string, key []byte, v interface{}) error {
	b, err := s.Get(name, key)
	if err != nil {
		return s.logError("load", err)
	}
	return s.logError("load", jsoniter.Unmarshal(b, v))
}

// Has checks for a key in the store.
func (s *boltStore) Has(name string, key []byte) (bool, error) {
	s.log.Debugf("has: key: %s", key)
	if err := storage.ValidKey(name, key); err != nil {
		return false, s.logError("has", err)
	}
	var has bool
	hasKey := func(tx *bolt.Tx) error {
		s.log.Debugf("has: tx get namespace: %s", name)
		b := tx.Bucket([]byte(name))
		if b == nil {
			err := fmt.Errorf("bucket not found: %s", name)
			return err
		}
		v := b.Get(key)
		has = len(v) > 0
		if has {
			s.log.Debugf("has: tx key found: %s.%s", name, string(key))
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
func (s *boltStore) Delete(name string, key []byte) error {
	s.log.Debugf("delete: key: %s", key)
	if err := storage.ValidKey(name, key); err != nil {
		return s.logError("delete", err)
	}
	del := func(tx *bolt.Tx) error {
		s.log.Debugf("delete: tx key: %s.%s", name, string(key))
		b := tx.Bucket([]byte(name))
		if b == nil {
			err := fmt.Errorf("bucket not found: %s", name)
			// an error just means we couldn't find the bucket
			s.log.Warn(err)
			return nil
		}
		return b.Delete(key)
	}
	return s.logError("delete", s.db.Update(del))
}

// List returns a list of all keys in the namespace.
func (s *boltStore) List(name string) (keys [][]byte, err error) {
	s.log.Debugf("list: keys in namespace: %s", name)
	listKeys := func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(name))
		if b == nil {
			err = fmt.Errorf("bucket not found: %s", name)
			return err
		}
		keys, err = s.listKeys(name, b)
		return err
	}
	if err = s.db.View(listKeys); err != nil {
		return nil, s.logError("list", err)
	}
	s.log.Debugf("list: found %d keys: %s", len(keys), keys)
	return
}

func (s *boltStore) listKeys(name string, b *bolt.Bucket) ([][]byte, error) {
	if b == nil {
		err := fmt.Errorf("invalid bucket: %s", name)
		return nil, err
	}
	var keys [][]byte
	s.log.Debugf("list: tx scan namespace: %s", name)
	count := b.Stats().KeyN
	keys = make([][]byte, count)
	s.log.Debugf("list: tx found %d keys in: %s", count, name)
	var i int
	_ = b.ForEach(func(k, _ []byte) error {
		s.log.Debugf("list: tx found key: %s.%s", name, string(k))
		keys[i] = k
		i++
		return nil
	})
	return keys, nil
}

// ListAll returns a mapped list of all keys in the store.
func (s *boltStore) ListAll() (map[string][][]byte, error) {
	s.log.Debugf("list: all keys")
	var total int
	allKeys := map[string][][]byte{}
	listKeys := func(tx *bolt.Tx) error {
		err := tx.ForEach(func(name []byte, b *bolt.Bucket) error {
			keys, err := s.listKeys(string(name), b)
			if err != nil {
				return err
			}
			if len(keys) <= 0 {
				return nil
			}
			allKeys[string(name)] = keys
			total += len(keys)
			return nil
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
func (s *boltStore) Export(path string) error {
	s.log.Debugf("export: to path: %s", path)
	if path == "" {
		err := fmt.Errorf("invalid path: %s", path)
		return s.logError("export", err)
	} else if s.path == path {
		err := fmt.Errorf("path cannot be store path: %s", path)
		return s.logError("export", err)
	}
	var err error
	path, err = ensureDBPath(path)
	if err != nil {
		return s.logError("export", err)
	}
	err = s.db.View(func(tx *bolt.Tx) error {
		return tx.CopyFile(path, 0600)
	})
	if err != nil {
		return s.logError("export", err)
	}
	s.log.Debugf("export: to path complete: %s", path)
	return nil
}

// Close closes the datastore and releases all db resources.
func (s *boltStore) Close() error {
	s.log.Debugf("closing store at path: %s", s.path)
	err := s.db.Close()
	s.db = nil
	s.log.Info("store closed")
	return s.logError("close", err)
}

func (s *boltStore) logError(name string, err error) error {
	if err == nil {
		return nil
	}
	if name != "" {
		err = fmt.Errorf("%s: %w", name, err)
	}
	s.log.Error(err)
	return err
}

func ensureDBPath(path string) (string, error) {
	if path == "" {
		return "", errors.New("path not found")
	}
	// does the path exist?
	info, err := os.Stat(path)
	exists := !os.IsNotExist(err)
	// this is some kind of actual error
	if err != nil && exists {
		return "", err
	}
	if exists && info.Mode().IsDir() {
		// if we have a directory, then append our default name
		path = filepath.Join(path, storeName)
	}
	ext := filepath.Ext(path)
	if ext == "" {
		path += storeExt
	}
	dir, _ := filepath.Split(path)
	// make sure the directory path exists
	if err = os.MkdirAll(dir, 0700); err != nil {
		return "", err
	}
	_, err = os.Stat(path)
	exists = !os.IsNotExist(err)
	// this is some kind of actual error
	if err != nil && exists {
		return "", err
	}
	if exists {
		return path, nil
	}
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return "", err
	}
	defer f.Close()
	return path, nil
}
