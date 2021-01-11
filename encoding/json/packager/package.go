package packager

import (
	"errors"
	"fmt"

	"github.com/hashicorp/go-version"
	"github.com/jrapoport/chestnut/encoding/json/encoders"
)

const (
	// Version is the current package fmt ver.
	Version = "0.0.1"

	// InvalidToken is an invalid sparse token.
	InvalidToken = ""

	// minCipher is the min length of base 64 enc ciphertext.
	minCipher = 4

	// minSparse is the min length of sparse enc data.
	minSparse = 2 // "{}" is an empty JSON object

	// minCompressed is the min length of compressed base 64 enc data.
	minCompressed = 8
)

// Format is the package fmt. Currently
// only secure & sparse formats are supported.
type Format string

const (
	// Secure indicates the package contains a fully encrypted JSON object.
	Secure Format = "secure"

	// Sparse indicates the package supports sparse decryption.
	Sparse Format = "sparse"
)

// Valid returns true if the fmt is valid.
func (f Format) Valid() bool {
	switch f {
	case Secure:
		return true
	case Sparse:
		return true
	default:
		return false
	}
}

// Package is returned by DecodePackage.
//
// - Secure: If the package fmt Format is Secure, Package contains the encrypted ciphertext
// Cipher containing a fully encoded JSON object.
//
// - Sparse: If the package Format is Sparse, the encrypted ciphertext Cipher contains a lookup
// table of secure values. Encoded contains a plaintext enc JSON with its secure fields
// removed and replaced with a secure lookup token consisting of a prefixed EncoderID with the
// fmt "[prefix]-[encoder id]" (SEE: NewLookupToken) and an index into the lookup table.
type Package struct {
	Version    string
	Format     Format
	Compressed bool
	EncoderID  string
	Token      string
	Cipher     []byte
	Encoded    []byte
}

// Valid returns true if the package fmt is valid.
func (p *Package) Valid() error {
	if len(p.Version) <= 0 {
		return errors.New("ver required")
	}
	_, err := version.NewVersion(p.Version)
	if err != nil {
		return fmt.Errorf("invalid ver %w", err)
	}
	if p.EncoderID == encoders.InvalidID {
		return errors.New("invalid encoder id")
	}
	if !p.Format.Valid() {
		return fmt.Errorf("invalid fmt %s", p.Format)
	}
	err = p.validateData()
	if err != nil {
		return err
	}
	sparse := len(p.Encoded) >= minSparse
	if sparse && p.Token == InvalidToken {
		return errors.New("invalid sparse token")
	}
	return nil
}

func (p *Package) validateData() error {
	if len(p.Cipher) < minCipher {
		return errors.New("invalid ciphertext")
	}
	if p.Compressed && len(p.Cipher) < minCompressed {
		return errors.New("invalid compressed ciphertext")
	}
	switch p.Format {
	case Secure:
		// this was handled above
		break
	case Sparse:
		if len(p.Encoded) < minSparse {
			return errors.New("invalid enc data")
		}
		if p.Compressed {
			if len(p.Encoded) < minCompressed {
				return errors.New("invalid compressed enc data")
			}
			break
		}
		// check that we have what looks like JSON
		if p.Encoded[0] != '{' {
			return errors.New("invalid enc data")
		}
	default:
		return fmt.Errorf("unsupported fmt: %s", p.Format)
	}
	return nil
}

// the currently supported package ver
var currentVer = version.Must(version.NewVersion(Version))

func (p *Package) checkVersion() error {
	if len(p.Version) <= 0 {
		return errors.New("ver required")
	}
	ver, err := version.NewVersion(p.Version)
	if err != nil {
		return err
	}
	if ver.GreaterThan(currentVer) {
		return fmt.Errorf("supported ver %s", ver)
	}
	return nil
}
