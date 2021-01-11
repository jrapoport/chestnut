package secure

import (
	"encoding/hex"
	"errors"
	"fmt"
	"reflect"
	"sync"

	"github.com/jrapoport/chestnut/encoding/json/encoders"
	"github.com/jrapoport/chestnut/encoding/json/encoders/hash"
	"github.com/jrapoport/chestnut/encoding/json/encoders/lookup"
	"github.com/jrapoport/chestnut/encoding/json/packager"
	"github.com/jrapoport/chestnut/encoding/tags"
	"github.com/jrapoport/chestnut/log"
	"github.com/json-iterator/go"
)

// SecureLookupPrefix will format the secure lookup token to "[prefix]-[encoder id]-[index]".
const SecureLookupPrefix = "cn"

// EncryptionFunction defines the prototype for the encryption callback.
// See WARNING regarding use of PassthroughEncryption.
type EncryptionFunction func(plaintext []byte) (ciphertext []byte, err error)

// PassthroughEncryption is a dummy function for development and testing *ONLY*.
/*
*   WARNING: DO NOT USE IN PRODUCTION.
*	PassthroughEncryption is *NOT* encryption and *DOES NOT* encrypt data.
 */
var PassthroughEncryption EncryptionFunction = func(plaintext []byte) ([]byte, error) {
	return []byte(hex.EncodeToString(plaintext)), nil
}

// EncoderExtension is a JSON encoder extension for the encryption and decryption of JSON
// encoded data. It supports full encryption / decryption of the encoded block in in
// addition to sparse encryption and hashing of structs on a per field basis via supplementary
// JSON struct field tag options. For additional information on sparse encryption & hashing, please
// SEE: https://github.com/jrapoport/chestnut/blob/master/README.md
//
// For additional information on json-iterator extensions, please
// SEE: https://github.com/json-iterator/go/wiki/Extension
type EncoderExtension struct {
	jsoniter.EncoderExtension
	opts         Options
	encoderID    string
	encoder      jsoniter.API
	lookupCtx    *lookup.Context
	lookupBuffer []byte
	open         bool
	encryptFunc  EncryptionFunction
	log          log.Logger
	mu           sync.RWMutex
}

// NewSecureEncoderExtension returns a new EncoderExtension using the supplied
// EncryptionFunction. If no encoder id is supplied, a new random encoder id will be used.
func NewSecureEncoderExtension(encoderID string, efn EncryptionFunction, opt ...Option) *EncoderExtension {
	const encoderName = "encoder"
	if encoderID == encoders.InvalidID {
		encoderID = encoders.NewEncoderID()
	}
	opts := DefaultOptions
	opts = applyOptions(opts, opt...)
	encoder := encoders.NewEncoder()
	logName := fmt.Sprintf("%s [%s]", encoderName, encoderID)
	token := lookup.NewLookupToken(SecureLookupPrefix, encoderID)
	ext := new(EncoderExtension)
	ext.opts = opts
	ext.log = log.Named(opts.log, logName)
	ext.encoderID = encoderID
	ext.encryptFunc = efn
	ext.encoder = encoder
	ext.lookupCtx = &lookup.Context{Token: token}
	if encoder == nil {
		ext.log.Fatal(errors.New("encoder not found"))
	}
	if efn == nil {
		ext.log.Panic(errors.New("encryption required"))
	}
	return ext
}

// Seal encrypts and returns the encoded value as a sealed package.
func (ext *EncoderExtension) Seal(encoded []byte) ([]byte, error) {
	ext.mu.Lock()
	defer ext.mu.Unlock()
	ext.log.Debugf("sealing %d encoded bytes: %s", len(encoded), string(encoded))
	/// must do this first
	if ext.open {
		ext.log.Debug("encoder is open, closing it")
		ext.close()
	}
	token := ext.lookupCtx.Token
	ext.log.Debugf("package token: %s", token)
	plaintext := ext.lookupBuffer
	if ext.isSparse() {
		ext.log.Debug("sparse encoding data")
	} else {
		ext.log.Debug("secure encoding data")
		plaintext = encoded
		token = ""
		encoded = nil
	}
	if ext.hasCompressor() {
		var err error
		ext.log.Debugf("compress %d plaintext bytes", len(plaintext))
		if plaintext, err = ext.compress(plaintext); err != nil {
			return nil, ext.logError(err)
		}
		ext.log.Debugf("compressed %d plaintext bytes", len(plaintext))
		ext.log.Debugf("compress %d encoded bytes", len(encoded))
		if encoded, err = ext.compress(encoded); err != nil {
			return nil, ext.logError(err)
		}
		ext.log.Debugf("compressed %d encoded bytes", len(encoded))
	}
	ext.log.Debugf("encrypting %d plaintext bytes: %s",
		len(plaintext), string(plaintext))
	// encrypt the blocks
	ciphertext, err := ext.encrypt(plaintext)
	if err != nil {
		return nil, ext.logError(err)
	}
	ext.log.Debugf("encrypted %d bytes", len(ciphertext))
	comp := ext.hasCompressor()
	ext.log.Debug("sealing package")
	pkg, err := packager.EncodePackage(ext.encoderID, token, ciphertext, encoded, comp)
	if err != nil {
		return nil, ext.logError(err)
	}
	ext.log.Debugf("sealed %d encoded bytes", len(pkg))
	return pkg, nil
}

func (ext *EncoderExtension) hasCompressor() bool {
	return ext.opts.compressor != nil
}

func (ext *EncoderExtension) compress(data []byte) ([]byte, error) {
	if len(data) <= 0 {
		return nil, nil
	}
	if !ext.hasCompressor() {
		return data, nil
	}
	return ext.opts.compressor(data)
}

// encrypt calls the EncryptionFunction if set, otherwise panic.
// See WARNING regarding the use of PassthroughEncryption.
func (ext *EncoderExtension) encrypt(plaintext []byte) ([]byte, error) {
	if ext.encryptFunc == nil {
		ext.log.Panic(errors.New("encryption function required"))
	}
	return ext.encryptFunc(plaintext)
}

// UpdateStructDescriptor customizes the encoding by specifying alternate
// lookup encoder for secure struct field tags and hash struct field strings.
func (ext *EncoderExtension) UpdateStructDescriptor(structDescriptor *jsoniter.StructDescriptor) {
	if !ext.isOpen() {
		ext.log.Debug("encoder is not open, cannot update struct descriptor")
		return
	}
	ext.log.Debugf("updating struct: %s", structDescriptor.Type)
	for _, binding := range structDescriptor.Fields {
		field := binding.Field
		typ := field.Type()
		ext.log.Debugf("updating struct field %s.%s", structDescriptor.Type, field.Name())
		tag, has := binding.Field.Tag().Lookup(tags.JSONTag)
		if !has {
			ext.log.Debug("json tag not found, ignore")
			continue
		}
		name, opts := tags.ParseJSONTag(tag)
		ext.log.Debugf("json tag name: %s options: %s", name, opts)
		if tags.IgnoreField(name) {
			ext.log.Debugf("json tag name %s, ignore", name)
			binding.ToNames = []string{}
			continue
		}
		hashName := tags.HashName(opts)
		secure := tags.IsSecure(opts)
		if !secure && hashName == tags.HashNone {
			ext.log.Debug("tag options not found, ignore")
			continue
		}
		encoder := binding.Encoder
		if secure {
			ext.log.Debugf("added lookup encoder to secure field %s", field.Name())
			encoder = lookup.NewLookupEncoder(ext.lookupCtx, typ, encoder)
			if enc, ok := encoder.(*lookup.Encoder); ok {
				enc.SetLogger(log.Named(ext.log, typ.String()))

			}
		}
		if hashName != tags.HashNone && typ.Kind() == reflect.String {
			// if the hash name is unsupported hashFn will be nil
			if hashFn := hash.FunctionForName(hashName); hashFn != nil {
				ext.log.Debugf("added %s hash encoder for field %s", field.Name(), hashName)
				encoder = hash.NewHashEncoder(hashName.String(), hashFn, encoder)
				if enc, ok := encoder.(*hash.Encoder); ok {
					enc.SetLogger(log.Named(ext.log, hashName.String()))
				}
			} else {
				ext.log.Warnf("%s hash encoder not found", hashName)
			}
		}
		binding.Encoder = encoder
	}
}

// Open should be called before Marshal to prepare the encoder.
func (ext *EncoderExtension) Open() error {
	ext.mu.Lock()
	defer ext.mu.Unlock()
	ext.log.Debug("opening encoder")
	if ext.open {
		return ext.logError(errors.New("encoder already open"))
	}
	if err := ext.openLookupStream(); err != nil {
		err = fmt.Errorf("failed to open encoder %w", err)
		return ext.logError(err)
	}
	ext.open = true
	ext.log.Debug("encoder open")
	return nil
}

func (ext *EncoderExtension) isOpen() bool {
	ext.mu.RLock()
	defer ext.mu.RUnlock()
	return ext.open
}

// Close should be called after Marshal, but before Seal. Calling
// Seal before Close will call Close automatically if necessary.
func (ext *EncoderExtension) Close() {
	ext.mu.Lock()
	defer ext.mu.Unlock()
	ext.close()
}

// close is the non-locking internal close call.
func (ext *EncoderExtension) close() {
	ext.log.Debug("closing encoder")
	ext.open = false
	ext.closeLookupStream()
	ext.log.Debug("encoder closed")
}

func (ext *EncoderExtension) openLookupStream() error {
	ext.log.Debug("opening lookup stream")
	stream := ext.encoder.BorrowStream(nil)
	if stream == nil {
		return ext.logError(errors.New("lookup stream is nil"))
	}
	if err := stream.Flush(); err != nil {
		err = fmt.Errorf("cannot flush lookup stream %w", err)
		return ext.logError(err)
	}
	ext.setupLookupContext(stream)
	if !ext.validLookupContext() {
		return ext.logError(errors.New("invalid lookup context"))
	}
	ext.log.Debug("lookup stream open")
	return nil
}

func (ext *EncoderExtension) setupLookupContext(stream *jsoniter.Stream) {
	ext.log.Debugf("setup lookup context: %s", ext.lookupCtx.Token)
	// reset the lookup index to 0
	stream.Attachment = 0
	stream.WriteObjectStart()
	ext.lookupCtx.Stream = stream
}

func (ext *EncoderExtension) validLookupContext() bool {
	if ext.lookupCtx == nil {
		ext.log.Error(errors.New("lookup context is nil"))
		return false
	}
	if ext.lookupCtx.Stream == nil {
		ext.log.Error(errors.New("lookup stream is nil"))
		return false
	}
	if ext.lookupCtx.Token == lookup.InvalidToken {
		ext.log.Error(errors.New("lookup token is invalid"))
		return false
	}
	sa := ext.lookupCtx.Stream.Attachment
	if sa == nil || sa.(int) != 0 {
		ext.log.Error(errors.New("lookup index is invalid"))
		return false
	}
	return true
}

func (ext *EncoderExtension) closeLookupStream() {
	ext.log.Debug("closing lookup stream")
	if ext.lookupCtx == nil {
		ext.log.Warn("lookup context is nil")
		return
	}
	stream := ext.lookupCtx.Stream
	if stream == nil {
		ext.log.Warn("lookup stream is nil")
		return
	}
	stream.WriteObjectEnd()
	ext.lookupBuffer = stream.Buffer()
	stream.Attachment = nil
	ext.encoder.ReturnStream(stream)
	ext.lookupCtx.Stream = nil
	ext.log.Debug("lookup stream closed")
}

// isSparse checks to see if the value used sparse encryption. If the encoded struct
// used struct tags to secure specific fields, we should have a lookup table.
func (ext *EncoderExtension) isSparse() bool {
	const emptyBuffer = "{}"
	return len(ext.lookupBuffer) > len(emptyBuffer)
}

func (ext *EncoderExtension) logError(e error) error {
	if e == nil {
		return e
	}
	ext.log.Error(e)
	return e
}
