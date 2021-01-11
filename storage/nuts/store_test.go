package nuts

import (
	"fmt"
	"sort"
	"testing"

	"github.com/google/uuid"
	"github.com/jrapoport/chestnut/log"
	"github.com/jrapoport/chestnut/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type testCase struct {
	name  string
	key   string
	value string
	err   assert.ErrorAssertionFunc
	has   assert.BoolAssertionFunc
}

type TestObject struct {
	Value string
}

var (
	testName  = "test-name"
	testKey   = "test-key"
	testValue = "test-value"
	testObj   = &TestObject{"hello"}
)

var putTests = []testCase{
	{"", "", "", assert.Error, assert.False},
	{"a", testKey, "", assert.Error, assert.False},
	{"b", testKey, testValue, assert.NoError, assert.True},
	{"c/c", testKey, testValue, assert.NoError, assert.True},
	{".d", testKey, testValue, assert.NoError, assert.True},
	{testName, "", "", assert.Error, assert.False},
	{testName, "a", "", assert.Error, assert.False},
	{testName, "b", testValue, assert.NoError, assert.True},
	{testName, "c/c", testValue, assert.NoError, assert.True},
	{testName, ".d", testValue, assert.NoError, assert.True},
	{testName, testKey, testValue, assert.NoError, assert.True},
}

var tests = append(putTests,
	testCase{testName, "not-found", "", assert.Error, assert.False},
)

type StoreTestSuite struct {
	suite.Suite
	store *Store
}

func TestStore(t *testing.T) {
	suite.Run(t, new(StoreTestSuite))
}

func (ts *StoreTestSuite) SetupTest() {
	ts.store = NewStore(ts.T().TempDir())
	err := ts.store.Open()
	assert.NoError(ts.T(), err)
}

func (ts *StoreTestSuite) TearDownTest() {
	err := ts.store.Close()
	assert.NoError(ts.T(), err)
}

func (ts *StoreTestSuite) BeforeTest(_, testName string) {
	switch testName {
	case "TestStore_Put",
		"TestStore_Save",
		"TestStore_Load",
		"TestStore_List",
		"TestStore_ListAll":
		break
	default:
		ts.TestStore_Put()
	}
}

func (ts *StoreTestSuite) TestStore_Put() {
	for i, test := range putTests {
		err := ts.store.Put(test.name, []byte(test.key), []byte(test.value))
		test.err(ts.T(), err, "%d test name: %s key: %s", i, test.name, test.key)
	}
}

func (ts *StoreTestSuite) TestStore_Save() {
	err := ts.store.Save(testName, []byte(testKey), testObj)
	assert.NoError(ts.T(), err)
}

func (ts *StoreTestSuite) TestStore_Load() {
	ts.T().Run("Setup", func(t *testing.T) {
		ts.TestStore_Save()
	})
	to := &TestObject{}
	err := ts.store.Load(testName, []byte(testKey), to)
	assert.NoError(ts.T(), err)
	assert.Equal(ts.T(), testObj, to)
}

func (ts *StoreTestSuite) TestStore_Get() {
	for i, test := range tests {
		value, err := ts.store.Get(test.name, []byte(test.key))
		test.err(ts.T(), err, "%d test name: %s key: %s", i, test.name, test.key)
		assert.Equal(ts.T(), test.value, string(value),
			"%d test key: %s", i, test.key)
	}
}

func (ts *StoreTestSuite) TestStore_Has() {
	for i, test := range tests {
		has, _ := ts.store.Has(test.name, []byte(test.key))
		test.has(ts.T(), has, "%d test key: %s", i, test.key)
	}
}

func (ts *StoreTestSuite) TestStore_List() {
	const listLen = 100
	list := make([]string, listLen)
	for i := 0; i < listLen; i++ {
		list[i] = uuid.New().String()
		err := ts.store.Put(testName, []byte(list[i]), []byte(testValue))
		assert.NoError(ts.T(), err)
	}
	keys, err := ts.store.List(testName)
	assert.NoError(ts.T(), err)
	assert.Len(ts.T(), keys, listLen)
	// put both lists in the same order so we can compare them
	strKeys := make([]string, len(keys))
	for i, k := range keys {
		strKeys[i] = string(k)
	}
	sort.Strings(list)
	sort.Strings(strKeys)
	assert.Equal(ts.T(), list, strKeys)
}

func (ts *StoreTestSuite) TestStore_ListAll() {
	const listLen = 100
	list := make([]string, listLen)
	for i := 0; i < listLen; i++ {
		list[i] = uuid.New().String()
		ns := fmt.Sprintf("%s%d", testName, i)
		err := ts.store.Put(ns, []byte(list[i]), []byte(testValue))
		assert.NoError(ts.T(), err)
	}
	keyMap, err := ts.store.ListAll()
	assert.NoError(ts.T(), err)
	var keys []string
	for _, ks := range keyMap {
		for _, k := range ks {
			keys = append(keys, string(k))
		}
	}
	assert.Len(ts.T(), keys, listLen)
	sort.Strings(list)
	sort.Strings(keys)
	assert.Equal(ts.T(), list, keys)
}

func (ts *StoreTestSuite) TestStore_Delete() {
	var deleteTests = []struct {
		key string
		err assert.ErrorAssertionFunc
	}{
		{"", assert.Error},
		{"a", assert.NoError},
		{"b", assert.NoError},
		{"c/c", assert.NoError},
		{".d", assert.NoError},
		{"eee", assert.NoError},
		{"not-found", assert.NoError},
	}
	for i, test := range deleteTests {
		err := ts.store.Delete(testName, []byte(test.key))
		test.err(ts.T(), err, "%d test key: %s", i, test.key)
	}
}

func (ts *StoreTestSuite) TestStore_Export() {
	err := ts.store.Export("")
	assert.Error(ts.T(), err)
	err = ts.store.Export(ts.store.path)
	assert.Error(ts.T(), err)
	err = ts.store.Export(ts.T().TempDir())
	assert.NoError(ts.T(), err)
}

func TestStore_WithLogger(t *testing.T) {
	levels := []log.Level{
		log.DebugLevel,
		log.InfoLevel,
		log.WarnLevel,
		log.ErrorLevel,
		log.PanicLevel,
	}
	type LoggerOpt func(log.Level) storage.StoreOption
	logOpts := []LoggerOpt{
		storage.WithLogrusLogger,
		storage.WithStdLogger,
		storage.WithZapLogger,
	}
	path := t.TempDir()
	for _, level := range levels {
		for _, logOpt := range logOpts {
			opt := logOpt(level)
			store := NewStore(path, opt)
			assert.NotNil(t, store)
			err := store.Open()
			assert.NoError(t, err)
			err = store.Close()
			assert.NoError(t, err)
		}
	}
}
