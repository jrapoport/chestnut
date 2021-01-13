package compress

import (
	"bytes"
	"encoding/hex"
)

// Format is the supporter compression algorithm type.
type Format string

// TODO: support additional compression algorithms besides
//  Zstandard from https://github.com/klauspost/compress
const (
	// None no compression.
	None Format = ""

	// Custom a custom compression format is being used.
	Custom Format = "custom"

	// Zstd Zstandard compression https://facebook.github.io/zstd/.
	Zstd Format = "zstd"
)

func (f Format) Valid() bool {
	switch f {
	case None:
		break
	case Custom:
		break
	case Zstd:
		break
	default:
		return false
	}
	return true
}



// CompressorFunc is the function the prototype for compression.
type CompressorFunc func(data []byte) (compressed []byte, err error)

// PassthroughCompressor is a dummy function for development and testing *ONLY*.
/*
*   WARNING: DO NOT USE IN PRODUCTION.
*	PassthroughCompressor is *NOT* compression and *DOES NOT* compress data.
 */
var PassthroughCompressor CompressorFunc = func(data []byte) ([]byte, error) {
	return []byte(hex.EncodeToString(data)), nil
}

// DecompressorFunc is the function the prototype for decompression.
type DecompressorFunc func(compressed []byte) (data []byte, err error)

// PassthroughDecompressor is a dummy function for development and testing *ONLY*.
/*
*   WARNING: DO NOT USE IN PRODUCTION.
*	PassthroughDecompressor is *NOT* decompression and *DOES NOT* decompress data.
 */
var PassthroughDecompressor DecompressorFunc = func(compressed []byte) ([]byte, error) {
	return hex.DecodeString(string(compressed))
}

var (
	formatTag = []byte{0xB, 0xA, 0xD, 0xA, 0x5, 0x5, 0x5, 0xB}
	formatSep = []byte{0x1e} // US-ASCII Record Separator
)

// EncodeFormat adds the compression format to the compressed data.
func EncodeFormat(data []byte, f Format) []byte {
	if f == None || len(data) <= 0 {
		return data
	}
	return bytes.Join([][]byte{formatTag, []byte(f), data}, formatSep)
}

// DecodeFormat removes and returns the compression format from the compressed data.
// If no compression format is found DecodeFormat returns the original the data.
func DecodeFormat(data []byte) ([]byte, Format) {
	if len(data) <= 0 {
		return data, None
	}
	if !bytes.HasPrefix(data, formatTag) {
		return data, None
	}
	parts := bytes.SplitN(data, formatSep, 3)
	if len(parts) < 3 {
		return data, None
	}
	// double check
	if !bytes.Equal(parts[0], formatTag) {
		return data, None
	}
	format := Format(parts[1])
	switch format {
	case Zstd:
		break
	case Custom:
		break
	default:
		return data, None
	}
	return parts[2], format
}
