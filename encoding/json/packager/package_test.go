package packager

import (
	"bytes"
	"encoding/gob"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

const (
	noComp = false
	comp   = true
	ver    = Version
)

var (
	empty     = ""
	id        = "c1ff7755"
	token     = "lookup-token"
	sec       = []byte("AAAAB3NzaC1yc2EAAAABJQAAAQB/nAmOjTmezNUDKYvEeIRf2Ynw")
	enc       = []byte("{\"test_object\":{\"" + token + "\":0}}")
	zstd      = []byte("KLUv/QQAAQEAeyJ0ZXN0X29iamVjdCI6eyJjbmMxZmY3NzU1IjowfX1hE1Nm")
	emptyZstd = []byte("KLUv/QQACQAAII1jaLY=")
	badVer    = "999.999.999"
	badVer2   = ".*"
	badFormat = Format("invalid")
	badData   = []byte("==")
	badZstd   = []byte("bm9wZQ")
	comps     = []bool{noComp, comp}
	secIns    = [][]byte{[]byte(nil), []byte(empty), badData, badZstd, sec, emptyZstd, zstd}
	encIns    = [][]byte{badData, enc, badZstd, emptyZstd, zstd}
)

type TestCase struct {
	ver       string
	fmt       Format
	id        string
	token     string
	comp      bool
	sec       []byte
	enc       []byte
	wrapErr   assert.ErrorAssertionFunc
	unwrapErr assert.ErrorAssertionFunc
}

var tests = []TestCase{
	// malformed packages
	{empty, "", empty, empty, noComp, nil, nil,
		assert.Error, assert.Error},
	{"0", "", empty, empty, noComp, nil, nil,
		assert.Error, assert.Error},
	{badVer, "", empty, empty, noComp, nil, nil,
		assert.Error, assert.Error},
	{badVer2, "", empty, empty, noComp, nil, nil,
		assert.Error, assert.Error},
	{ver, "", empty, empty, noComp, nil, nil,
		assert.Error, assert.Error},
	{ver, badFormat, empty, empty, noComp, nil, nil,
		assert.Error, assert.Error},
	{ver, badFormat, id, empty, noComp, nil, nil,
		assert.Error, assert.Error},
	{ver, Secure, id, empty, noComp, nil, nil,
		assert.Error, assert.Error},
	{ver, Sparse, empty, empty, noComp, nil, nil,
		assert.Error, assert.Error},
	{ver, Sparse, id, empty, noComp, nil, nil,
		assert.Error, assert.Error},
	// valid packages
	{ver, Secure, id, empty, noComp, sec, nil,
		assert.NoError, assert.NoError},
	{ver, Sparse, id, token, noComp, sec, enc,
		assert.NoError, assert.NoError},
	// valid compressed packages
	{ver, Secure, id, empty, comp, zstd, nil,
		assert.NoError, assert.NoError},
	{ver, Sparse, id, token, comp, zstd, zstd,
		assert.NoError, assert.NoError},
}

func genSecureTestCases() {
	for _, c := range comps {
		for secIdx, secIn := range secIns {
			wrapErr := assert.Error
			unwrapErr := assert.Error
			if c {
				if secIdx >= 4 {
					wrapErr = assert.NoError
					unwrapErr = assert.NoError
				}
			} else {
				if secIdx >= 3 {
					wrapErr = assert.NoError
					unwrapErr = assert.NoError
				}
			}
			tc := TestCase{
				ver:       ver,
				fmt:       Secure,
				id:        id,
				comp:      c,
				sec:       secIn,
				wrapErr:   wrapErr,
				unwrapErr: unwrapErr,
			}
			tests = append(tests, tc)
		}
	}
}

func genSparseTestCases() {
	for _, c := range comps {
		for secIdx, secIn := range secIns {
			for encIdx, encIn := range encIns {
				wrapErr := assert.Error
				unwrapErr := assert.Error
				if c {
					if encIdx == 1 {
						continue
					} else if secIdx >= 4 && encIdx > 2 {
						wrapErr = assert.NoError
						unwrapErr = assert.NoError
					}
				} else {
					if secIdx >= 3 && encIdx == 1 {
						wrapErr = assert.NoError
						unwrapErr = assert.NoError
					}
				}
				tc := TestCase{
					ver:       ver,
					fmt:       Sparse,
					id:        id,
					token:     token,
					comp:      c,
					sec:       secIn,
					enc:       encIn,
					wrapErr:   wrapErr,
					unwrapErr: unwrapErr,
				}
				tests = append(tests, tc)
			}
		}
	}
}

type PackageTestSuite struct {
	suite.Suite
}

func TestStore(t *testing.T) {
	suite.Run(t, new(PackageTestSuite))
}

func (ts *PackageTestSuite) SetupSuite() {
	genSecureTestCases()
	genSparseTestCases()
}

func (ts *PackageTestSuite) TestPackage_Encode() {
	for _, test := range tests {
		bytes, err := EncodePackage(test.id, test.token, test.sec, test.enc, test.comp)
		test.wrapErr(ts.T(), err)
		if err == nil {
			ts.NotEmpty(bytes)
		} else {
			ts.Empty(bytes)
		}
	}
}

func (ts *PackageTestSuite) TestPackage_Decode() {
	for _, test := range tests {
		testPkg := &Package{
			Version:    test.ver,
			Format:     test.fmt,
			Compressed: test.comp,
			EncoderID:  test.id,
			Token:      test.token,
			Cipher:     test.sec,
			Encoded:    test.enc,
		}
		_, err := encode(testPkg)
		test.unwrapErr(ts.T(), err)
	}

	for _, test := range tests {
		testPkg := &Package{
			Version:    test.ver,
			Format:     test.fmt,
			Compressed: test.comp,
			EncoderID:  test.id,
			Token:      test.token,
			Cipher:     test.sec,
			Encoded:    test.enc,
		}
		b := bytes.Buffer{}
		e := gob.NewEncoder(&b)
		err := e.Encode(testPkg)
		ts.NoError(err)
		pkg, err := DecodePackage(b.Bytes())
		test.unwrapErr(ts.T(), err)
		if err != nil {
			ts.Nil(pkg)
		} else {
			assertPackage(ts.T(), test, pkg)
		}
	}
}

func (ts *PackageTestSuite) TestPackage() {
	for _, test := range tests {
		bytes, err := EncodePackage(test.id, test.token, test.sec, test.enc, test.comp)
		test.wrapErr(ts.T(), err, string(bytes))
		if err != nil {
			ts.Empty(string(bytes))
			continue
		} else {
			ts.NotEmpty(string(bytes))
		}
		pkg, err := DecodePackage(bytes)
		test.unwrapErr(ts.T(), err)
		if err != nil {
			ts.Nil(pkg)
		} else {
			assertPackage(ts.T(), test, pkg)
		}
	}
}

func assertPackage(t *testing.T, test TestCase, pkg *Package) {
	assert.NotNil(t, pkg)
	assert.NoError(t, pkg.Valid())
	assert.Equal(t, test.ver, pkg.Version)
	assert.Equal(t, test.fmt, pkg.Format)
	assert.Equal(t, test.comp, pkg.Compressed)
	assert.Equal(t, test.id, pkg.EncoderID)
	assert.Equal(t, test.token, pkg.Token)
	assert.Equal(t, test.sec, pkg.Cipher)
	assert.Equal(t, test.enc, pkg.Encoded)
}
