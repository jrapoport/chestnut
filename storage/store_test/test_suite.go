package store_test

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

type testObject struct {
	Value string
}

var (
	testName  = "test-name"
	testKey   = "test-key"
	testValue = "test-value"
	testObj   = &testObject{"hello"}
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

type storeFunc = func(string, ...storage.StoreOption) storage.Storage

type storeTestSuite struct {
	suite.Suite
	storeFunc
	store storage.Storage
	path  string
}

// TestStore tests a store
func TestStore(t *testing.T, fn storeFunc) {
	ts := new(storeTestSuite)
	ts.storeFunc = fn
	suite.Run(t, ts)
}

// SetupTest
func (ts *storeTestSuite) SetupTest() {
	ts.path = ts.T().TempDir()
	ts.store = ts.storeFunc(ts.path)
	err := ts.store.Open()
	ts.NoError(err)
}

// TearDownTest
func (ts *storeTestSuite) TearDownTest() {
	err := ts.store.Close()
	ts.NoError(err)
}

// BeforeTest
func (ts *storeTestSuite) BeforeTest(_, testName string) {
	switch testName {
	case "TestStorePut",
		"TestStoreSave",
		"TestStoreLoad",
		"TestStoreList",
		"TestStoreListAll",
		"TestStoreWithLogger":
		break
	default:
		ts.TestStorePut()
	}
}

// TestStorePut
func (ts *storeTestSuite) TestStorePut() {
	for i, test := range putTests {
		err := ts.store.Put(test.name, []byte(test.key), []byte(test.value))
		test.err(ts.T(), err, "%d test name: %s key: %s", i, test.name, test.key)
	}
}

// TestStoreSave
func (ts *storeTestSuite) TestStoreSave() {
	err := ts.store.Save(testName, []byte(testKey), testObj)
	ts.NoError(err)
}

// TestStoreLoad
func (ts *storeTestSuite) TestStoreLoad() {
	ts.T().Run("Setup", func(t *testing.T) {
		ts.TestStoreSave()
	})
	to := &testObject{}
	err := ts.store.Load(testName, []byte(testKey), to)
	ts.NoError(err)
	ts.Equal(testObj, to)
}

// TestStoreGet
func (ts *storeTestSuite) TestStoreGet() {
	for i, test := range tests {
		value, err := ts.store.Get(test.name, []byte(test.key))
		test.err(ts.T(), err, "%d test name: %s key: %s", i, test.name, test.key)
		ts.Equal(test.value, string(value),
			"%d test key: %s", i, test.key)
	}
}

// TestStoreHas
func (ts *storeTestSuite) TestStoreHas() {
	for i, test := range tests {
		has, _ := ts.store.Has(test.name, []byte(test.key))
		test.has(ts.T(), has, "%d test key: %s", i, test.key)
	}
}

// TestStoreList
func (ts *storeTestSuite) TestStoreList() {
	const listLen = 100
	list := make([]string, listLen)
	for i := 0; i < listLen; i++ {
		list[i] = uuid.New().String()
		err := ts.store.Put(testName, []byte(list[i]), []byte(testValue))
		ts.NoError(err)
	}
	keys, err := ts.store.List(testName)
	ts.NoError(err)
	ts.Len(keys, listLen)
	// put both lists in the same order so we can compare them
	strKeys := make([]string, len(keys))
	for i, k := range keys {
		strKeys[i] = string(k)
	}
	sort.Strings(list)
	sort.Strings(strKeys)
	ts.Equal(list, strKeys)
}

// TestStoreListAll
func (ts *storeTestSuite) TestStoreListAll() {
	const listLen = 100
	list := make([]string, listLen)
	for i := 0; i < listLen; i++ {
		list[i] = uuid.New().String()
		ns := fmt.Sprintf("%s%d", testName, i)
		err := ts.store.Put(ns, []byte(list[i]), []byte(testValue))
		ts.NoError(err)
	}
	keyMap, err := ts.store.ListAll()
	ts.NoError(err)
	var keys []string
	for _, ks := range keyMap {
		for _, k := range ks {
			keys = append(keys, string(k))
		}
	}
	ts.Len(keys, listLen)
	sort.Strings(list)
	sort.Strings(keys)
	ts.Equal(list, keys)
}

// TestStoreDelete
func (ts *storeTestSuite) TestStoreDelete() {
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

// TestStoreExport
func (ts *storeTestSuite) TestStoreExport() {
	exTests := []struct {
		path string
		Err  assert.ErrorAssertionFunc
	}{
		{"", assert.Error},
		{ts.path, assert.Error},
		{ts.T().TempDir(), assert.NoError},
	}
	for _, test := range exTests {
		err := ts.store.Export(test.path)
		test.Err(ts.T(), err)
		if err == nil {
			s2 := ts.storeFunc(test.path)
			ts.NotNil(s2)
			err = s2.Open()
			ts.NoError(err)
			keys, err := s2.ListAll()
			ts.NoError(err)
			ts.NotEmpty(keys)
			err = s2.Close()
			ts.NoError(err)
		}
	}
}

// TestStoreWithLogger
func (ts *storeTestSuite) TestStoreWithLogger() {
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
	path := ts.T().TempDir()
	for _, level := range levels {
		for _, logOpt := range logOpts {
			opt := logOpt(level)
			store := ts.storeFunc(path, opt)
			ts.NotNil(store)
			err := store.Open()
			ts.NoError(err)
			err = store.Close()
			ts.NoError(err)
		}
	}
}
