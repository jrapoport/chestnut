package keystore

import (
	"log"
	"sort"
	"testing"

	"github.com/google/uuid"
	"github.com/jrapoport/chestnut"
	"github.com/jrapoport/chestnut/encryptor/aes"
	"github.com/jrapoport/chestnut/encryptor/crypto"
	"github.com/jrapoport/chestnut/storage"
	"github.com/jrapoport/chestnut/storage/nuts"
	ci "github.com/libp2p/go-libp2p-core/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

var (
	testName     = uuid.New().String()
	textSecret   = crypto.TextSecret("i-am-a-good-secret")
	encryptorOpt = chestnut.WithAES(crypto.Key256, aes.CFB, textSecret)
	privateKey   = func() ci.PrivKey {
		pk, _, err := ci.GenerateKeyPair(ci.ECDSA, 512)
		if err != nil {
			log.Fatal(err)
		}
		return pk
	}()
)

type testCase struct {
	name   string
	key    ci.PrivKey
	err    assert.ErrorAssertionFunc
	exists bool
}

var tests = []testCase{
	{"", nil, assert.Error, false},
	{"", nil, assert.Error, false},
	{"f", nil, assert.Error, false},
	{"g", privateKey, assert.NoError, true},
	{"h", privateKey, assert.NoError, true},
	{"i/i", privateKey, assert.NoError, true},
	{".j", privateKey, assert.NoError, true},
	{testName, privateKey, assert.NoError, true},
}

var testCaseNotFound = testCase{"not-found", nil, assert.Error, false}

type KeystoreTestSuite struct {
	suite.Suite
	keystore *Keystore
}

func newNutsDBStore(t *testing.T) storage.Storage {
	path := t.TempDir()
	store := nuts.NewStore(path)
	assert.NotNil(t, store)
	return store
}

func TestKeystore(t *testing.T) {
	suite.Run(t, new(KeystoreTestSuite))
}

func (ts *KeystoreTestSuite) SetupTest() {
	store := newNutsDBStore(ts.T())
	ts.keystore = NewKeystore(store, encryptorOpt)
	ts.NotNil(ts.keystore)
	err := ts.keystore.Open()
	ts.NoError(err)
}

func (ts *KeystoreTestSuite) TearDownTest() {
	err := ts.keystore.Close()
	ts.NoError(err)
}

func (ts *KeystoreTestSuite) BeforeTest(_, testName string) {
	switch testName {
	case "TestKeystore_Encryptor",
		"TestKeystore_Put",
		"TestKeystore_List":
		break
	default:
		ts.TestKeystore_Put()
	}
}

func TestInvalidConfig(t *testing.T) {
	assert.Panics(t, func() {
		NewKeystore(nil, encryptorOpt)
	})
}

func (ts *KeystoreTestSuite) TestKeystore_Encryptor() {
	err := ts.keystore.Put(testName, privateKey)
	ts.NoError(err)
	pk, err := ts.keystore.Get(testName)
	ts.NotNil(pk)
	ts.NoError(err)
	ts.Equal(privateKey.Type().String(), pk.Type().String())
}

func (ts *KeystoreTestSuite) TestKeystore_Put() {
	for i, test := range tests {
		err := ts.keystore.Put(test.name, test.key)
		test.err(ts.T(), err, "%d test name: %s", i, test.name)
	}
	err := ts.keystore.Put(testName, privateKey)
	ts.Error(err)
}

func (ts *KeystoreTestSuite) TestKeystore_Get() {
	getTests := append(tests, testCaseNotFound)
	for i, test := range getTests {
		key, err := ts.keystore.Get(test.name)
		test.err(ts.T(), err, "%d test name: %s", i, test.name)
		ts.Equal(test.key, key, "%d test name: %s", i, test.name)
	}
}

func (ts *KeystoreTestSuite) TestKeystore_Has() {
	for _, test := range tests {
		has, _ := ts.keystore.Has(test.name)
		ts.Equal(test.exists, has)
	}
}

func (ts *KeystoreTestSuite) TestKeystore_List() {
	const listLen = 100
	list := make([]string, listLen)
	for i := 0; i < listLen; i++ {
		list[i] = uuid.New().String()
		err := ts.keystore.Put(list[i], privateKey)
		ts.NoError(err)
	}
	keys, err := ts.keystore.List()
	ts.NoError(err)
	ts.Len(keys, listLen)
	// put both lists in the same order so we can compare them
	sort.Strings(list)
	sort.Strings(keys)
	ts.Equal(list, keys)
}

func (ts *KeystoreTestSuite) TestKeystore_Delete() {
	for i, test := range tests {
		if test.exists == false {
			continue
		}
		err := ts.keystore.Delete(test.name)
		test.err(ts.T(), err, "%d test key: %s", i, test.key)
	}
}

func (ts *KeystoreTestSuite) TestKeystore_Export() {
	err := ts.keystore.Export(ts.T().TempDir())
	ts.NoError(err)
}

func TestKeystore_OpenErr(t *testing.T) {
	ks := &Keystore{}
	err := ks.Open()
	assert.Error(t, err)
}
