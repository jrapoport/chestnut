package secure

import (
	"github.com/jrapoport/chestnut/encoding/compress"
	"github.com/jrapoport/chestnut/encoding/compress/zstd"
	"github.com/jrapoport/chestnut/log"
)

// Options provides a default implementation for common options for a secure encoding.
type Options struct {
	// compressor is only valid for encoders
	compressor compress.CompressorFunc

	// decompressor is only valid for decoders
	decompressor compress.DecompressorFunc

	// sparse is only valid for decoding sparse packages
	sparse bool

	// log is the logger to use
	log log.Logger
}

// DefaultOptions represents the recommended default Options for secure encoding.
var DefaultOptions = Options{
	log: log.Log,
}

// A Option sets options such as compression or sparse decoding.
type Option interface {
	apply(*Options)
}

// EmptyOption does not alter the encoder configuration. It can be embedded
// in another structure to build custom encoder options.
type EmptyOption struct{}

func (EmptyOption) apply(*Options) {}

// funcOption wraps a function that modifies Options
// into an implementation of the Option interface.
type funcOption struct {
	f func(*Options)
}

// apply applies an Option to Options.
func (fdo *funcOption) apply(do *Options) {
	fdo.f(do)
}

func newFuncOption(f func(*Options)) *funcOption {
	return &funcOption{
		f: f,
	}
}

// applyOptions accepts a Options struct and applies the Option(s) to it.
func applyOptions(opts Options, opt ...Option) Options {
	if opt != nil {
		for _, o := range opt {
			o.apply(&opts)
		}
	}
	return opts
}

// SparseDecode returns a Option that set the decoder to return sparsely
// decoded data. If the JSON data was not sparely encoded, this does nothing.
func SparseDecode() Option {
	return newFuncOption(func(o *Options) {
		o.sparse = true
	})
}

// WithCompressor returns a Option that compresses data.
func WithCompressor(compressor compress.CompressorFunc) Option {
	return newFuncOption(func(o *Options) {
		o.compressor = compressor
	})
}

// WithDecompressor returns a Option that decompresses data.
func WithDecompressor(decompressor compress.DecompressorFunc) Option {
	return newFuncOption(func(o *Options) {
		o.decompressor = decompressor
	})
}

// WithCompression returns a Option that compresses & decompresses data with Zstd.
func WithCompression(format compress.Format) Option {
	return newFuncOption(func(o *Options) {
		switch format {
		case compress.Zstd:
			o.compressor = zstd.Compress
			o.decompressor = zstd.Decompress
		default:
			break
		}
	})
}

// WithLogger returns a Option which sets the logger for the extension.
func WithLogger(l log.Logger) Option {
	return newFuncOption(func(o *Options) {
		o.log = l
	})
}
