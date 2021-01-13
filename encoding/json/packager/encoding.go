package packager

import (
	"bytes"
	"encoding/gob"
	"errors"
	"github.com/jrapoport/chestnut/encoding/json/encoders"
)

// EncodePackage returns a valid binary enc package for storage.
func EncodePackage(encoderID, token string, cipher, encoded []byte, compressed bool) ([]byte, error) {
	if encoderID == encoders.InvalidID {
		return nil, errors.New("invalid encoder id")
	}
	format := Secure
	// are we sparse?
	sparse := len(encoded) >= minSparse
	if sparse {
		format = Sparse
	}
	// start the package
	pkg := &Package{
		Version:    currentVer.String(),
		Format:     format,
		Compressed: compressed,
		EncoderID:  encoderID,
		Token:      token,
		Cipher:     cipher,
		Encoded:    encoded,
	}
	if err := pkg.Valid(); err != nil {
		return nil, err
	}
	return encode(pkg)
}

func encode(pkg *Package) ([]byte, error) {
	if err := pkg.Valid(); err != nil {
		return nil, err
	}
	b := bytes.Buffer{}
	e := gob.NewEncoder(&b)
	if err := e.Encode(pkg); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

// DecodePackage takes packaged data and returns the ciphertext and encoding block.
func DecodePackage(bytes []byte) (*Package, error) {
	pkg, err := decode(bytes)
	if err != nil {
		return nil, err
	}
	// check the package ver
	if err = pkg.checkVersion(); err != nil {
		return nil, err
	}
	if err = pkg.Valid(); err != nil {
		return nil, err
	}
	return pkg, err
}

func decode(data []byte) (*Package, error) {
	pkg := &Package{}
	buf := bytes.Buffer{}
	buf.Write(data)
	d := gob.NewDecoder(&buf)
	return pkg, d.Decode(pkg)
}
