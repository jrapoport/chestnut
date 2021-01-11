package log

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLevel(t *testing.T) {
	levels := []Level{
		DebugLevel,
		InfoLevel,
		WarnLevel,
		ErrorLevel,
		PanicLevel,
		FatalLevel,
	}
	type NewLoggerFunc func(Level) Logger
	tests := []struct {
		name  string
		logFn NewLoggerFunc
	}{
		{"logrus", NewLogrusLoggerWithLevel},
		{"std", NewStdLoggerWithLevel},
		{"zap", NewZapLoggerWithLevel},
	}
	for _, level := range levels {
		for _, test := range tests {
			logger := test.logFn(level)
			// debug
			logger.Debug(test.name, " ", "debug")
			logger.Debugf("%s %s", test.name, "debug")
			// info
			logger.Info(test.name, " ", "info")
			logger.Infof("%s %s", test.name, "info")
			// warn
			logger.Warn(test.name, " ", "warn")
			logger.Warnf("%s %s", test.name, "warn")
			// error
			logger.Error(test.name, " ", "error")
			logger.Errorf("%s %s", test.name, "error")
			// panic
			assert.Panics(t, func() {
				logger.Panic(test.name, " ", "panic")
			})
			assert.Panics(t, func() {
				logger.Panicf("%s %s", test.name, "panic")
			})
		}
	}
}
