package log

import (
	"log"
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestWrapper(t *testing.T) {
	const (
		testName  = "test"
		emptyName = ""
	)
	tests := []struct {
		logger    interface{}
		name      string
		assertNil assert.ValueAssertionFunc
	}{
		{nil, emptyName, assert.Nil},
		{logrus.New(), emptyName, assert.NotNil},
		{logrus.New(), testName, assert.NotNil},
		{logrus.New().WithContext(nil), emptyName, assert.NotNil},
		{logrus.New().WithContext(nil), testName, assert.NotNil},
		{NewLogrusLoggerWithLevel(ErrorLevel), emptyName, assert.NotNil},
		{NewLogrusLoggerWithLevel(ErrorLevel), testName, assert.NotNil},
		{log.New(os.Stderr, "", 0), emptyName, assert.NotNil},
		{log.New(os.Stderr, "", 0), testName, assert.NotNil},
		{NewStdLoggerWithLevel(ErrorLevel), emptyName, assert.NotNil},
		{NewStdLoggerWithLevel(ErrorLevel), testName, assert.NotNil},
		{zap.NewExample(), emptyName, assert.NotNil},
		{zap.NewExample(), testName, assert.NotNil},
		{zap.NewExample().Sugar(), emptyName, assert.NotNil},
		{zap.NewExample().Sugar(), testName, assert.NotNil},
		{NewZapLoggerWithLevel(ErrorLevel), emptyName, assert.NotNil},
		{NewZapLoggerWithLevel(ErrorLevel), testName, assert.NotNil},
	}

	for _, test := range tests {
		logger := Named(test.logger, "name")
		test.assertNil(t, logger)
		if logger != nil {
			_, ok := logger.(Logger)
			assert.True(t, ok)
			// error
			logger.Error(testName)
			logger.Errorf("%s", testName)
		}
	}
}
