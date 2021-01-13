package chestnut

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/jrapoport/chestnut/encoding/compress"
	"github.com/jrapoport/chestnut/encoding/compress/zstd"
	"github.com/jrapoport/chestnut/encoding/json"
	"github.com/jrapoport/chestnut/encoding/json/encoders/secure"
	"github.com/jrapoport/chestnut/log"
	"github.com/jrapoport/chestnut/storage"
	"github.com/jrapoport/chestnut/value"
)

// Chestnut is used to manage an encrypted store. It provides additional features such
// as chained encryption, independently secured secrets, sparse encryption, and hashing.
// For more detail, SEE: https://github.com/jrapoport/chestnut/blob/master/README.md
type Chestnut struct {
	opts  ChestOptions
	store storage.Storage
	log   log.Logger
}

// NewChestnut is used to create a new chestnut encrypted store.
func NewChestnut(store storage.Storage, opt ...ChestOption) *Chestnut {
	const logName = "chestnut"
	//logger := storage.LoggerFromStore(store, logName)
	opts := applyOptions(DefaultChestOptions, opt...)
	logger := log.Named(opts.log, logName)
	cn := &Chestnut{opts, store, logger}
	if err := cn.validConfig(); err != nil {
		logger.Panic(err)
		return nil
	}
	return cn
}

func (cn *Chestnut) validConfig() error {
	if cn.store == nil {
		return errors.New("store required")
	}
	if cn.opts.encryptor == nil {
		return errors.New("encryptor is required")
	}
	if cn.opts.compressor != nil || cn.opts.decompressor != nil {
		cn.opts.compression = compress.Custom
	}
	if cn.opts.compression == compress.Custom && cn.opts.compressor == nil {
		return errors.New("compressor is required")
	}
	if cn.opts.compression == compress.Custom && cn.opts.decompressor == nil {
		return errors.New("decompressor is required")
	}
	if !cn.opts.compression.Valid() {
		return errors.New("invalid compression format")
	}
	return nil
}

// Open the storage chest
func (cn *Chestnut) Open() error {
	if err := cn.validConfig(); err != nil {
		return err
	}
	cn.log.Debug("opening storage chest")
	if err := cn.store.Open(); err != nil {
		return cn.logError("open", err)
	}
	cn.log.Info("storage chest open")
	if cn.opts.encryptor != nil {
		cn.log.Infof("using encryption: %s",
			cn.opts.encryptor.Name())
	}
	if cn.opts.compression != compress.None {
		cn.log.Infof("%s compression active",
			cn.opts.compression)
	}
	if !cn.opts.overwrites {
		cn.log.Info("overwrites are disabled")
	}
	return nil
}

// Put encrypts the plaintext and stores it at key.
func (cn *Chestnut) Put(name string, key []byte, plaintext []byte) error {
	cn.log.Debugf("put: %d plaintext bytes to key: %s", len(plaintext), key)
	// the store will make these same checks, but encryption
	// is expensive, so we are going to do them upfront here.
	if err := storage.ValidKey(name, key); err != nil {
		return cn.logError("put", err)
	} else if len(plaintext) <= 0 {
		err = errors.New("plaintext cannot be empty")
		return cn.logError("put", err)
	} else if err = cn.CanPut(name, key); err != nil {
		return cn.logError("put", err)
	}
	if cn.opts.compression != compress.None {
		var err error
		if plaintext, err = cn.compress(plaintext); err != nil {
			return cn.logError("put", err)
		}
	}
	cn.log.Debugf("put: encrypt %d bytes", len(plaintext))
	cipherText, err := cn.encrypt(plaintext)
	if err != nil {
		return cn.logError("put", err)
	}
	cn.log.Debugf("put: encrypted %d bytes", len(cipherText))
	return cn.logError("", cn.store.Put(name, key, cipherText))
}

// Get decrypts the ciphertext at key and returns the plaintext.
func (cn *Chestnut) Get(name string, key []byte) ([]byte, error) {
	cn.log.Debugf("get: ciphertext at key: %s", key)
	ciphertext, err := cn.store.Get(name, key)
	if err != nil {
		return nil, cn.logError("", err)
	}
	cn.log.Debugf("get: decrypt %d bytes", len(ciphertext))
	plaintext, err := cn.decrypt(ciphertext)
	if err != nil {
		return nil, cn.logError("get", err)
	}
	cn.log.Debugf("put: decrypted %d bytes", len(plaintext))
	// decompress will check to see if the data is compressed.
	// if sit is not compressed, it returns the buffer.
	if plaintext, err = cn.decompress(plaintext); err != nil {
		return nil, cn.logError("get", err)
	}
	return plaintext, nil
}

// Save encrypts the struct in v and stores the encoded result at key.
func (cn *Chestnut) Save(name string, key []byte, v interface{}) error {
	cn.log.Debugf("save: %v value to key: %s", reflect.TypeOf(v), key)
	// the store will make these same checks, but encryption
	// is expensive, so we are going to do them upfront here.
	if err := storage.ValidKey(name, key); err != nil {
		return cn.logError("save", err)
	} else if v == nil {
		err = errors.New("value cannot be nil")
		return cn.logError("save", err)
	} else if err = cn.CanPut(name, key); err != nil {
		return cn.logError("save", err)
	}
	cn.log.Debugf("save: encrypt %v value", reflect.TypeOf(v))
	ciphertext, err := cn.marshal(v)
	if err != nil {
		return cn.logError("save", err)
	}
	cn.log.Debugf("save: put %d encrypted bytes", len(ciphertext))
	if err = cn.store.Put(name, key, ciphertext); err != nil {
		return cn.logError("save", err)
	}
	cn.log.Debugf("save: encrypted %v value", reflect.TypeOf(v))
	return nil
}

// Load decrypts the struct at key and returns the decoded result in v.
func (cn *Chestnut) Load(name string, key []byte, v interface{}) error {
	cn.log.Debugf("load: %v value at key: %s", reflect.TypeOf(v), key)
	if err := cn.load(name, key, v, false); err != nil {
		return cn.logError("load", err)
	}
	cn.log.Debugf("load: decrypted %v value", reflect.TypeOf(v))
	return nil
}

// Sparse loads the struct at key and returns the sparsely decoded result in v.
// Unlike Load, it does not decrypt the encoded struct and secure fields are
// replaced with empty values. If the struct was not saved as a sparsely encoded
// struct this has no effect and is equivalent to calling Load. Structs must
// have been saved with secure fields to be loaded as sparse structs by Sparse.
func (cn *Chestnut) Sparse(name string, key []byte, v interface{}) error {
	cn.log.Debugf("sparse: %v value at key: %s", reflect.TypeOf(v), key)
	if err := cn.load(name, key, v, true); err != nil {
		return cn.logError("sparse", err)
	}
	cn.log.Debugf("sparse: decrypted sparse %v value", reflect.TypeOf(v))
	return nil
}

// SaveKeyed encrypts the keyed value and stores the result.
func (cn *Chestnut) SaveKeyed(v value.Keyed) error {
	if v == nil || reflect.ValueOf(v).IsNil() {
		err := errors.New("value cannot be nil")
		return cn.logError("save value", err)
	} else if err := v.ValidKey(); err != nil {
		return cn.logError("save value", err)
	}
	err := cn.Save(v.Namespace(), v.Key(), v)
	return cn.logError("save value", err)
}

// LoadKeyed decrypts the keyed value and returns the result.
func (cn *Chestnut) LoadKeyed(v value.Keyed) error {
	if v == nil || reflect.ValueOf(v).IsNil() {
		err := errors.New("value cannot be nil")
		return cn.logError("load value", err)
	} else if err := v.ValidKey(); err != nil {
		return cn.logError("load value", err)
	}
	err := cn.Load(v.Namespace(), v.Key(), v)
	return cn.logError("load value", err)
}

// SparseKeyed sparsely loads the keyed value and returns the result.
// Unlike LoadKeyed, it does not decrypt the keyed value. SEE: Sparse.
func (cn *Chestnut) SparseKeyed(v value.Keyed) error {
	if v == nil || reflect.ValueOf(v).IsNil() {
		err := errors.New("value cannot be nil")
		return cn.logError("sparse value", err)
	} else if err := v.ValidKey(); err != nil {
		return cn.logError("sparse value", err)
	}
	err := cn.Sparse(v.Namespace(), v.Key(), v)
	return cn.logError("sparse value", err)
}

// Has checks for a key in the storage chest. Has returns true
// if the key is found, otherwise false.
func (cn *Chestnut) Has(name string, key []byte) (bool, error) {
	cn.log.Debugf("has: key: %s", key)
	has, err := cn.store.Has(name, key)
	cn.log.Debugf("has: key %s: %t", key, has)
	return has, cn.logError("", err)
}

// CanPut returns nil if writing to key is ok. If overwrites
// are disabled and the key exists, ErrForbidden is returned.
func (cn *Chestnut) CanPut(name string, key []byte) error {
	cn.log.Debugf("can put: key: %s", key)
	if err := storage.ValidKey(name, key); err != nil {
		return cn.logError("can put", err)
	}
	if cn.opts.overwrites {
		cn.log.Debug("can put: overwrites enabled")
		return nil
	}
	// if overwrites are disabled check to see if we have the key. if
	// the key does not exist, has will return an error. technically,
	// this could get confused because an error would also be returned
	// if the key was invalid, but since we already check that above
	// we can safely assume that's not the error. since we expect an error
	// when the key is not found has Has will log it, we can ignore it.
	if has, _ := cn.Has(name, key); has {
		return cn.logError("can put", ErrForbidden)
	}
	// we didn't find the key and there is no error, this is not an overwrite.
	cn.log.Debugf("can put: can write to key: %s", key)
	return nil
}

// Delete removes a key from the storage chest.
func (cn *Chestnut) Delete(name string, key []byte) error {
	cn.log.Debugf("delete: key: %s", key)
	return cn.logError("", cn.store.Delete(name, key))
}

// List returns a list of keys in the namespace.
func (cn *Chestnut) List(namespace string) ([][]byte, error) {
	cn.log.Infof("list: all keys")
	keys, err := cn.store.List(namespace)
	cn.log.Debugf("list: found %d keys: %s", len(keys), keys)
	return keys, cn.logError("", err)
}

// Export saves a copy of the storage chest to directory at path.
func (cn *Chestnut) Export(path string) error {
	cn.log.Debugf("export: to path: %s", path)
	return cn.logError("", cn.store.Export(path))
}

// Close the storage chest
func (cn *Chestnut) Close() error {
	cn.log.Info("closing storage chest")
	if err := cn.store.Close(); err != nil {
		return cn.logError("close", err)
	}
	cn.log.Info("storage chest closed")
	return nil
}

// Logger gets a copy of the logger from the storage chest's options
func (cn *Chestnut) Logger() log.Logger {
	return cn.opts.log
}

// SetLogger sets the logger for the storage chest
func (cn *Chestnut) SetLogger(l log.Logger) {
	if l == nil {
		l = log.Log
	}
	cn.log = l
}

// load decrypts the secure or sparse value at key and stores the result in v.
func (cn *Chestnut) load(name string, key []byte, v interface{}, sparse bool) error {
	if v == nil {
		return errors.New("value cannot be nil")
	}
	ciphertext, err := cn.store.Get(name, key)
	if err != nil {
		return err
	}
	return cn.unmarshal(ciphertext, v, sparse)
}

// encrypt returns the plaintext data as ciphertext.
func (cn *Chestnut) encrypt(plaintext []byte) (ciphertext []byte, err error) {
	cn.log.Debugf("encrypt: encrypting %d bytes", len(plaintext))
	ciphertext, err = cn.opts.encryptor.Encrypt(plaintext)
	if err != nil {
		err = cn.logError("encrypt", err)
		return
	}
	cn.log.Debugf("encrypt: encrypted %d bytes", len(ciphertext))
	return
}

// decrypt returns the ciphertext data as plaintext.
func (cn *Chestnut) decrypt(ciphertext []byte) (plaintext []byte, err error) {
	cn.log.Debugf("decrypt: decrypting %d bytes", len(ciphertext))
	plaintext, err = cn.opts.encryptor.Decrypt(ciphertext)
	if err != nil {
		err = cn.logError("decrypt", err)
		return
	}
	cn.log.Debugf("decrypt: decrypted %d bytes", len(plaintext))
	return
}

// marshal returns the JSON encoding of v as ciphertext.
func (cn *Chestnut) marshal(v interface{}) (ciphertext []byte, err error) {
	if v == nil {
		err = errors.New("value cannot be nil")
		return nil, cn.logError("marshal", err)
	}
	cn.log.Debugf("marshal: %v value", reflect.TypeOf(v))
	ciphertext, err = json.SecureMarshal(v, cn.encrypt, secure.WithLogger(cn.log))
	if err != nil {
		err = cn.logError("marshal", err)
		return
	}
	cn.log.Debugf("marshal: encrypted %d bytes", len(ciphertext))
	return
}

// unmarshal returns the plaintext decoded JSON value at v.
func (cn *Chestnut) unmarshal(ciphertext []byte, v interface{}, sparse bool) error {
	if v == nil {
		err := errors.New("value cannot be nil")
		return cn.logError("unmarshal", err)
	}
	cn.log.Debugf("unmarshal: decrypt %d bytes to %v value",
		len(ciphertext), reflect.TypeOf(v))
	opts := []secure.Option{secure.WithLogger(cn.log)}
	if sparse {
		cn.log.Debug("use sparse decoding")
		opts = append(opts, secure.SparseDecode())
	}
	err := json.SecureUnmarshal(ciphertext, v, cn.decrypt, opts...)
	if err != nil {
		return cn.logError("unmarshal", err)
	}
	cn.log.Debugf("unmarshal: decrypted %v value", reflect.TypeOf(v))
	return nil
}

func (cn *Chestnut) compress(data []byte) ([]byte, error) {
	format := cn.opts.compression
	if format == compress.None {
		return data, nil
	}
	var compressor compress.CompressorFunc
	switch format {
	case compress.Zstd:
		compressor = zstd.Compress
	case compress.Custom:
		compressor = cn.opts.compressor
	default:
		break
	}
	if compressor == nil {
		err := fmt.Errorf("%s unsupported", format)
		return nil, cn.logError("compress", err)
	}
	size := len(data)
	cn.log.Debugf("compressing %d bytes with %s", size, format)
	compressed, err := compressor(data)
	if err != nil {
		return nil, cn.logError("compress", err)
	}
	compressed = compress.EncodeFormat(compressed, format)
	cn.log.Debugf("%s compressed encrypted bytes to %d (%0.2f%% reduction)",
		format, len(compressed), calcReduction(size, len(compressed)))
	return compressed, nil
}

func calcReduction(oldSize, newSize int) float64 {
	return ((float64(oldSize) - float64(newSize)) / float64(oldSize)) * 100
}

func (cn *Chestnut) decompress(data []byte) ([]byte, error) {
	// check for compression
	compressed, format := compress.DecodeFormat(data)
	// this does not appear to be data we compressed
	if format == compress.None {
		return compressed, nil
	}
	var decompressor compress.DecompressorFunc
	switch format {
	case compress.Zstd:
		decompressor = zstd.Decompress
	case compress.Custom:
		decompressor = cn.opts.decompressor
	default:
		break
	}
	if decompressor == nil {
		err := fmt.Errorf("%s unsupported", format)
		return nil, cn.logError("decompress", err)
	}
	cn.log.Debugf("decompressing %d bytes with %s", len(compressed), format)
	decompressed, err := decompressor(compressed)
	if err != nil {
		return nil, cn.logError("decompress", err)
	}
	cn.log.Debugf("decompressed %d bytes with %s",
		len(decompressed), format)
	return decompressed, nil
}

func (cn *Chestnut) logError(name string, err error) error {
	if err == nil {
		return nil
	}
	if name != "" {
		err = fmt.Errorf("%s: %w", name, err)
	}
	cn.log.Error(err)
	return err
}

// ErrForbidden the storage operation is forbidden
var ErrForbidden = errors.New("forbidden")
