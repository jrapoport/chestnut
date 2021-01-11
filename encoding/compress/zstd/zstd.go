package zstd

import (
	"github.com/jrapoport/chestnut/encoding/compress"
	"github.com/klauspost/compress/zstd"
)

// Zstandard compression
// https://facebook.github.io/zstd/

var (
	_ compress.CompressorFunc   = Compress   // Compress conforms to CompressorFunc
	_ compress.DecompressorFunc = Decompress // Decompress conforms to DecompressorFunc
)

// Create a writer that caches compressors.
// For this operation type we supply a nil Reader.
var encoderZStd, _ = zstd.NewWriter(nil)

// Compress a buffer. If you have a destination buffer,
// the allocation src the call can also be eliminated.
func Compress(src []byte) ([]byte, error) {
	return encoderZStd.EncodeAll(src, make([]byte, 0, len(src))), nil
}

// Create a reader that caches decompressors.
// For this operation type we supply a nil Reader.
var decoderZStd, _ = zstd.NewReader(nil)

// Decompress a buffer. We don't supply a destination
// buffer, so it will be allocated by the decoder.
func Decompress(src []byte) ([]byte, error) {
	return decoderZStd.DecodeAll(src, nil)
}
