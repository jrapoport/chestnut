package log

import (
	"io"
	"log"
	"os"
)

// stdLogger is a wrapper of standard log.
type stdLogger struct {
	*log.Logger
	level Level
}

var _ Logger = (*stdLogger)(nil)

// NewStdLoggerWithLevel is a stderr logger with the log level.
func NewStdLoggerWithLevel(lvl Level) Logger {
	return NewStdLogger(lvl, os.Stderr, "", log.LstdFlags)
}

// NewStdLogger returns a new standard logger with the log level.
func NewStdLogger(lvl Level, out io.Writer, prefix string, flag int) Logger {
	return &stdLogger{log.New(out, prefix, flag), lvl}
}

// Debug logs args when the logger level is debug.
func (l *stdLogger) Debug(v ...interface{}) {
	if l.level > DebugLevel {
		return
	}
	l.Print(v...)
}

// Debugf formats args and logs the result when the logger level is debug.
func (l *stdLogger) Debugf(format string, v ...interface{}) {
	if l.level > DebugLevel {
		return
	}
	l.Printf(format, v...)
}

// Info logs args when the logger level is info.
func (l *stdLogger) Info(v ...interface{}) {
	if l.level > InfoLevel {
		return
	}
	l.Print(v...)
}

// Infof formats args and logs the result when the logger level is info.
func (l *stdLogger) Infof(format string, v ...interface{}) {
	if l.level > InfoLevel {
		return
	}
	l.Printf(format, v...)
}

// Warn logs args when the logger level is warn.
func (l *stdLogger) Warn(v ...interface{}) {
	if l.level > WarnLevel {
		return
	}
	l.Print(v...)
}

// Warnf formats args and logs the result when the logger level is warn.
func (l *stdLogger) Warnf(format string, v ...interface{}) {
	if l.level > WarnLevel {
		return
	}
	l.Printf(format, v...)
}

// Error logs args when the logger level is error.
func (l *stdLogger) Error(v ...interface{}) {
	if l.level > ErrorLevel {
		return
	}
	l.Print(v...)
}

// Errorf formats args and logs the result when the logger level is debug.
func (l *stdLogger) Errorf(format string, v ...interface{}) {
	if l.level > ErrorLevel {
		return
	}
	l.Printf(format, v...)
}

// Panic logs args on panic.
func (l *stdLogger) Panic(v ...interface{}) {
	l.Logger.Panic(v...)
}

// Panicf formats args and logs the result on panic.
func (l *stdLogger) Panicf(format string, v ...interface{}) {
	l.Logger.Panicf(format, v...)
}

// Fatal logs args when the error is fatal.
func (l *stdLogger) Fatal(v ...interface{}) {
	l.Logger.Fatal(v...)
}

// Fatalf formats args and logs the result when the error is fatal.
func (l *stdLogger) Fatalf(format string, v ...interface{}) {
	l.Logger.Fatalf(format, v...)
}
