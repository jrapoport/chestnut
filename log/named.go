package log

import (
	"log"

	"github.com/sirupsen/logrus"
	"go.uber.org/zap"
)

// logrusField matches zap
const logrusField = "logger"

// Named adds a name string to the logger. How the name is added is
// logger specific i.e. a logrus field or std logger prefix, etc.
func Named(logger interface{}, name string) Logger {
	switch l := logger.(type) {
	case *logrus.Logger:
		return l.WithField(logrusField, name)
	case *logrus.Entry:
		return l.WithField(logrusField, name)
	case *log.Logger:
		l.SetPrefix(name + " ")
		return &stdLogger{l, InfoLevel}
	case *stdLogger:
		l.SetPrefix(name + " ")
		return l
	case *zap.SugaredLogger:
		return l.Named(name)
	case *zap.Logger:
		return l.Sugar().Named(name)
	}
	return nil
}
