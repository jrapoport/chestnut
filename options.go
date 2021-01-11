package chestnut

import (
	"github.com/jrapoport/chestnut/encoding/compress"
	"github.com/jrapoport/chestnut/encryptor"
	"github.com/jrapoport/chestnut/encryptor/crypto"
	"github.com/jrapoport/chestnut/log"
)

// ChestOptions provides a default implementation for common options for a secure store.
type ChestOptions struct {
	encryptor       crypto.Encryptor
	chainEncryptors []crypto.Encryptor
	compression     compress.Format
	compressor      compress.CompressorFunc
	decompressor    compress.DecompressorFunc
	// overwrites allows a storage chest to save data over existing data with the same storage key.
	// if Overwrite is true, overwrite are enabled and successive calls to save data
	// 	with the same key will succeed. The existing data will be overwritten by the new data.
	// if Overwrite is false, overwrite are disabled and successive calls to save data
	// 	with the same key will fail with an error. The existing data will not be overwritten.
	overwrites bool
	log        log.Logger
}

// DefaultChestOptions represents the recommended default ChestOptions for a store.
var DefaultChestOptions = ChestOptions{
	overwrites: true,
	log:        log.Log,
}

// A ChestOption sets options such as encryptors, key rolling, and other parameters, etc.
type ChestOption interface {
	apply(*ChestOptions)
}

// EmptyChestOption does not alter the encrypted store's configuration.
// It can be embedded in another structure to build custom options.
type EmptyChestOption struct{}

func (EmptyChestOption) apply(*ChestOptions) {}

// funcOption wraps a function that modifies ChestOptions
// into an implementation of the ChestOption interface.
type funcOption struct {
	f func(*ChestOptions)
}

// apply applies an Option to ChestOptions.
func (fdo *funcOption) apply(do *ChestOptions) {
	fdo.f(do)
}

func newFuncOption(f func(*ChestOptions)) *funcOption {
	return &funcOption{
		f: f,
	}
}

// applyOptions accepts a ChestOptions struct and applies the ChestOption(s) to it.
func applyOptions(opts ChestOptions, opt ...ChestOption) ChestOptions {
	for _, o := range opt {
		o.apply(&opts)
	}
	chainEncryptors(&opts)
	return opts
}

// chainEncryptors chains all encryptors into one.
func chainEncryptors(opts *ChestOptions) {
	// Prepend opts.encryptor to the chaining encryptors if it exists, so that single
	// encryptor will be executed before any other chained encryptor.
	encryptors := opts.chainEncryptors
	if opts.encryptor != nil {
		encryptors = append([]crypto.Encryptor{opts.encryptor}, opts.chainEncryptors...)
	}
	var chained crypto.Encryptor
	if len(encryptors) == 0 {
		chained = nil
	} else if len(encryptors) == 1 {
		chained = encryptors[0]
	} else {
		chained = encryptor.NewChainEncryptor(encryptors...)
	}
	opts.encryptor = chained
}

// WithEncryptor returns a ChestOption that specifies the encryptor to use.
func WithEncryptor(e crypto.Encryptor) ChestOption {
	return newFuncOption(func(o *ChestOptions) {
		if o.encryptor != nil {
			panic("The encryptor was already set and may not be reset.")
		}
		o.encryptor = e
	})
}

// WithEncryptorChain returns a ChestOption that specifies an encryptor chain.
// for encrypted stores. The first encryptor will be the outer most,
// while the last encryptor will be the inner most wrapper around the real call.
// All encryptors added by this method will be chained. If a single encryptor
// has also been set, it will be *prepended* to the encryptor chain,
// making it the outer most encryptor in the encryptor chain.
func WithEncryptorChain(encryptors ...crypto.Encryptor) ChestOption {
	return newFuncOption(func(o *ChestOptions) {
		o.chainEncryptors = append(o.chainEncryptors, encryptors...)
	})
}

// WithAES is a convenience that returns a ChestOption which sets the encryptor
// to be an AESEncryptor initialized with a key length, cipher mode, and Secret.
func WithAES(keyLen crypto.KeyLen, mode crypto.Mode, secret crypto.Secret) ChestOption {
	return WithEncryptor(encryptor.NewAESEncryptor(keyLen, mode, secret))
}

// WithCompressors instructs the storage chest to compress/decompress data with these compressor
// functions before committing it. If this option is set, WithCompression is ignored.
func WithCompressors(c compress.CompressorFunc, d compress.DecompressorFunc) ChestOption {
	return newFuncOption(func(o *ChestOptions) {
		o.compression = compress.Custom
		o.compressor = c
		o.decompressor = d
	})
}

// WithCompression instructs the storage chest to compress data using the this compression format
// before committing it. Compression this way is self-contained, meaning changes only effect data
// going forward. Previously saved data, compressed or uncompressed, will be transparently retrieved
// regardless of a change to this setting.
func WithCompression(format compress.Format) ChestOption {
	return newFuncOption(func(o *ChestOptions) {
		o.compression = format
	})
}

// OverwritesForbidden prevents the store from overwriting existing data.
func OverwritesForbidden() ChestOption {
	return newFuncOption(func(o *ChestOptions) {
		o.overwrites = false
	})
}

// WithLogger returns a StoreOption which sets the logger to use for the encrypted store.
func WithLogger(l log.Logger) ChestOption {
	return newFuncOption(func(o *ChestOptions) {
		o.log = l
	})
}

// WithStdLogger is a convenience that returns a StoreOption for a standard err logger.
func WithStdLogger(lvl log.Level) ChestOption {
	return WithLogger(log.NewStdLoggerWithLevel(lvl))
}

// WithLogrusLogger is a convenience that returns a StoreOption for a default logrus logger.
func WithLogrusLogger(lvl log.Level) ChestOption {
	return WithLogger(log.NewLogrusLoggerWithLevel(lvl))
}

// WithZapLogger is a convenience that returns a StoreOption for a production zap logger.
func WithZapLogger(lvl log.Level) ChestOption {
	return WithLogger(log.NewZapLoggerWithLevel(lvl))
}
