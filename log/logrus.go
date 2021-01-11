package log

import "github.com/sirupsen/logrus"

var _ Logger = (*logrus.Logger)(nil)
var _ Logger = (*logrus.Entry)(nil)

// NewLogrusLoggerWithLevel returns a new production logrus logger with the log level.
func NewLogrusLoggerWithLevel(lvl Level) Logger {
	l := logrus.New()
	l.SetLevel(levelToLogrusLevel(lvl))
	return l.WithContext(nil)
}

// NOTE: for logrus panic is a higher level than fatal.
func levelToLogrusLevel(lvl Level) logrus.Level {
	switch lvl {
	case DebugLevel:
		return logrus.DebugLevel
	case InfoLevel:
		return logrus.InfoLevel
	case WarnLevel:
		return logrus.WarnLevel
	case ErrorLevel:
		return logrus.ErrorLevel
	case PanicLevel:
		return logrus.PanicLevel
	case FatalLevel:
		return logrus.FatalLevel
	default:
		return logrus.InfoLevel
	}
}
