package log

// Log is the same as the default standard logger from "log".
var Log = NewStdLoggerWithLevel(PanicLevel)

// Logger is a generic logger interface.
type Logger interface {
	// Debug logs args when the logger level is debug.
	Debug(v ...interface{})

	// Debugf formats args and logs the result when the logger level is debug.
	Debugf(format string, v ...interface{})

	// Info logs args when the logger level is info.
	Info(args ...interface{})

	// Infof formats args and logs the result when the logger level is info.
	Infof(format string, v ...interface{})

	// Warn logs args when the logger level is warn.
	Warn(v ...interface{})

	// Warnf formats args and logs the result when the logger level is warn.
	Warnf(format string, v ...interface{})

	// Error logs args when the logger level is error.
	Error(v ...interface{})

	// Errorf formats args and logs the result when the logger level is debug.
	Errorf(format string, v ...interface{})

	// Panic logs args on panic.
	Panic(v ...interface{})

	// Panicf formats args and logs the result on panic.
	Panicf(format string, v ...interface{})

	// Fatal logs args when the error is fatal.
	Fatal(v ...interface{})

	// Fatalf formats args and logs the result when the error is fatal.
	Fatalf(format string, v ...interface{})
}
