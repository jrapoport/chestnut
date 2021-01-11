package secure

import (
	"encoding/hex"
	"errors"
	"fmt"
	"sync"

	"github.com/jrapoport/chestnut/encoding/json/encoders"
	"github.com/jrapoport/chestnut/encoding/json/encoders/lookup"
	"github.com/jrapoport/chestnut/encoding/json/packager"
	"github.com/jrapoport/chestnut/log"
	jsoniter "github.com/json-iterator/go"
	"github.com/modern-go/reflect2"
)

// DecryptionFunction defines the prototype for the decryption callback.
// See WARNING regarding use of PassthroughDecryption.
type DecryptionFunction func(ciphertext []byte) (plaintext []byte, err error)

// PassthroughDecryption is a dummy function for development and testing *ONLY*.
/*
*   WARNING: DO NOT USE IN PRODUCTION.
*	PassthroughDecryption is *NOT* decryption and *DOES NOT* decrypt data.
 */
var PassthroughDecryption DecryptionFunction = func(ciphertext []byte) ([]byte, error) {
	return hex.DecodeString(string(ciphertext))
}

// DecoderExtension is a JSON encoder extension for the encryption and decryption of JSON
// encoded data. It supports full encryption / decryption of the encoded block in in
// addition to sparse encryption and hashing of structs on a per field basis via supplementary
// JSON struct field tag options. For addition information sparse encryption & hashing, please
// SEE: https://github.com/jrapoport/chestnut/blob/master/README.md
//
// For additional information on json-iterator extensions, please
// SEE: https://github.com/json-iterator/go/wiki/Extension
type DecoderExtension struct {
	jsoniter.DecoderExtension
	opts         Options
	encoderID    string
	encoder      jsoniter.API
	lookupCtx    *lookup.Context
	lookupBuffer []byte
	open         bool
	decryptFunc  DecryptionFunction
	log          log.Logger
	mu           sync.RWMutex
}

// NewSecureDecoderExtension returns a new DecoderExtension using the supplied DecryptionFunction. If
// an encoder id is supplied, this decoder will restrict itself to packages with a matching id.
func NewSecureDecoderExtension(encoderID string, dfn DecryptionFunction, opt ...Option) *DecoderExtension {
	const decoderName = "decoder"
	opts := DefaultOptions
	opts = applyOptions(opts, opt...)
	encoder := encoders.NewEncoder()
	logName := decoderName
	if encoderID != encoders.InvalidID {
		logName += fmt.Sprintf(" [%s]", encoderID)
	}
	ext := new(DecoderExtension)
	ext.opts = opts
	ext.log = log.Named(opts.log, logName)
	ext.encoderID = encoderID
	ext.decryptFunc = dfn
	ext.encoder = encoder
	ext.lookupCtx = &lookup.Context{}
	if ext.encoder == nil {
		ext.log.Fatal(errors.New("encoder not found"))
	}
	if ext.decryptFunc == nil {
		ext.log.Panic(errors.New("decryption function required"))
	}
	return ext
}

// Unseal decrypts and returns the encoded value as an unsealed package. If sparse
// is true AND the data format is sparse, the data will not be decrypted the struct
// will be decoded with empty values in place of secure fields.
// TODO: We could hash the encoded data and add that to our plaintext block before we
//  encrypt it as a tamper check. Not sure that is necessary or useful right now though.
func (ext *DecoderExtension) Unseal(encoded []byte) ([]byte, error) {
	ext.mu.Lock()
	defer ext.mu.Unlock()
	ext.log.Debugf("unsealing encoded %d bytes", len(encoded))
	/// must do this first
	if ext.open {
		ext.log.Debug("decoder is open, closing it")
		ext.close()
	}
	// unwrap the package
	pkg, err := packager.DecodePackage(encoded)
	if err != nil {
		return nil, ext.logError(err)
	}
	if err = pkg.Valid(); err != nil {
		err = fmt.Errorf("invalid encoding %w", err)
		return nil, ext.logError(err)
	}
	compressed := pkg.Compressed
	ext.log.Debugf("package data is compressed: %t", compressed)
	// IF we have an encoder ID, check that it matches the package
	ext.log.Debugf("checking encoding id %s", pkg.EncoderID)
	if ext.encoderID != encoders.DefaultID &&
		ext.encoderID != pkg.EncoderID {
		err = fmt.Errorf(" encoder %s package %s id mismatch", ext.encoderID, pkg.EncoderID)
		return nil, ext.logError(err)
	}
	ext.log.Debugf("sparse option set: %t", ext.opts.sparse)
	isSparse := pkg.Format == packager.Sparse && ext.opts.sparse
	ext.log.Debugf("sparse decoding: %t", isSparse)
	if !isSparse {
		// decrypt the data unless we are sparse decoding
		ext.log.Debugf("decrypting %d ciphertext bytes", len(pkg.Cipher))
		if pkg.Cipher, err = ext.decrypt(pkg.Cipher); err != nil {
			return nil, ext.logError(err)
		}
		ext.log.Debugf("decrypted %d bytes", len(pkg.Cipher))
		if compressed {
			ext.log.Debug("ciphertext is compressed")
			if !ext.hasDecompressor() {
				err = errors.New("compressed package requires decompressor")
				return nil, ext.logError(err)
			}
			ext.log.Debugf("decompress %d ciphertext bytes", len(pkg.Cipher))
			pkg.Cipher, err = ext.decompress(pkg.Cipher)
			if err != nil {
				return nil, ext.logError(err)
			}
			ext.log.Debugf("decompressed %d ciphertext bytes", len(pkg.Cipher))
		}
	}
	switch pkg.Format {
	// the format is secure, we are done
	case packager.Secure:
		ext.log.Debugf("unsealed %d secure data bytes: %s", len(pkg.Cipher), string(pkg.Cipher))
		return pkg.Cipher, nil
	case packager.Sparse:
		// set the lookup context
		ext.log.Debugf("unsealed sparse token: %s", pkg.Token)
		ext.lookupCtx.Token = pkg.Token
		if !isSparse {
			ext.log.Debugf("unsealed %d lookup data bytes: %s", len(pkg.Cipher), string(pkg.Cipher))
			ext.lookupBuffer = pkg.Cipher
		}
		break
	default:
		return nil, ext.logError(errors.New("unknown package format"))
	}
	if compressed {
		ext.log.Debug("encoded data is compressed")
		if !ext.hasDecompressor() {
			err = errors.New("compressed package requires decompressor")
			return nil, ext.logError(err)
		}
		ext.log.Debugf("decompress %d encoded bytes", len(pkg.Encoded))
		pkg.Encoded, err = ext.decompress(pkg.Encoded)
		if err != nil {
			return nil, ext.logError(err)
		}
		ext.log.Debugf("decompressed %d encoded bytes", len(pkg.Encoded))
	}
	if len(pkg.Encoded) > 0 {
		ext.log.Debugf("unsealed %d sparse data bytes: %s", len(pkg.Encoded), string(pkg.Encoded))
	}
	return pkg.Encoded, nil
}

func (ext *DecoderExtension) hasDecompressor() bool {
	return ext.opts.decompressor != nil
}

func (ext *DecoderExtension) decompress(data []byte) ([]byte, error) {
	if len(data) <= 0 {
		return nil, nil
	}
	if !ext.hasDecompressor() {
		return data, nil
	}
	return ext.opts.decompressor(data)
}

// decrypt calls the DecryptionFunction if set, otherwise panic.
// See WARNING regarding the use of PassthroughDecryption.
func (ext *DecoderExtension) decrypt(ciphertext []byte) ([]byte, error) {
	if ext.decryptFunc == nil {
		ext.log.Panic(errors.New("decryption function required"))
	}
	return ext.decryptFunc(ciphertext)
}

// DecorateDecoder customizes the decoding by specifying alternate lookup table decoder that
// recognizes previously encoded lookup table keys and replaces them with decoded values.
func (ext *DecoderExtension) DecorateDecoder(typ reflect2.Type, decoder jsoniter.ValDecoder) jsoniter.ValDecoder {
	if !ext.isOpen() {
		ext.log.Debug("decoder is not open, cannot decorate decoder")
		return decoder
	}
	if ext.lookupCtx == nil || ext.lookupCtx.Token == lookup.InvalidToken {
		ext.log.Debug("decoding is not sparse, do not add lookup decoder")
		return decoder
	}
	ext.log.Debugf("added lookup decoder for type: %s", typ)
	decoder = lookup.NewLookupDecoder(ext.lookupCtx, typ, decoder)
	if dec, ok := decoder.(*lookup.Decoder); ok {
		dec.SetLogger(log.Named(ext.log, typ.String()))
	}
	return decoder
}

// Open should be called before Unmarshal to prepare the decoder.
func (ext *DecoderExtension) Open() error {
	ext.mu.Lock()
	defer ext.mu.Unlock()
	ext.log.Debug("opening decoder")
	if ext.open {
		return ext.logError(errors.New("decoder already open"))
	}
	if err := ext.openLookupStream(); err != nil {
		err = fmt.Errorf("failed to open decoder %w", err)
		return ext.logError(err)
	}
	ext.open = true
	ext.log.Debug("decoder open")
	return nil
}

func (ext *DecoderExtension) isOpen() bool {
	ext.mu.RLock()
	defer ext.mu.RUnlock()
	return ext.open
}

// Close should be called after Unmarshal.
func (ext *DecoderExtension) Close() {
	ext.mu.Lock()
	defer ext.mu.Unlock()
	ext.close()
}

// close is the non-locking internal close call.
func (ext *DecoderExtension) close() {
	ext.log.Debug("closing decoder")
	ext.closeLookupStream()
	ext.open = false
	ext.log.Debug("decoder closed")
}

func (ext *DecoderExtension) openLookupStream() error {
	ext.log.Debug("opening lookup stream")
	stream := ext.encoder.BorrowStream(nil)
	if stream == nil {
		return ext.logError(errors.New("lookup stream is nil"))
	}
	if err := stream.Flush(); err != nil {
		err = fmt.Errorf("cannot flush lookup stream %w", err)
		return ext.logError(err)
	}
	// setup the lookup context
	ext.setupLookupContext(stream)
	if !ext.validLookupContext() {
		return ext.logError(errors.New("invalid lookup context"))
	}
	ext.log.Debug("lookup stream open")
	return nil
}

func (ext *DecoderExtension) setupLookupContext(stream *jsoniter.Stream) {
	ext.log.Debugf("setup lookup context: %s", ext.lookupCtx.Token)
	stream.Attachment = ext.encoder.Get(ext.lookupBuffer)
	ext.lookupCtx.Stream = stream
	ext.lookupBuffer = nil
}

func (ext *DecoderExtension) validLookupContext() bool {
	if ext.lookupCtx == nil {
		ext.log.Error(errors.New("lookup context is nil"))
		return false
	}
	if ext.lookupCtx.Stream == nil {
		ext.log.Error(errors.New("lookup stream is nil"))
		return false
	}
	if ext.lookupCtx.Token != lookup.InvalidToken &&
		ext.lookupCtx.Stream.Attachment == nil {
		ext.log.Error(errors.New("lookup table is nil"))
		return false
	}
	return true
}

func (ext *DecoderExtension) closeLookupStream() {
	ext.log.Debug("closing lookup stream")
	ext.lookupBuffer = nil
	if ext.lookupCtx == nil {
		ext.log.Warn("lookup context is nil")
		return
	}
	stream := ext.lookupCtx.Stream
	if stream == nil {
		ext.log.Warn("lookup stream is nil")
		return
	}
	stream.Attachment = nil
	ext.encoder.ReturnStream(stream)
	ext.lookupCtx.Token = lookup.InvalidToken
	ext.lookupCtx.Stream = nil
	ext.log.Debug("lookup stream closed")
}

func (ext *DecoderExtension) logError(e error) error {
	if e == nil {
		return e
	}
	ext.log.Error(e)
	return e
}
