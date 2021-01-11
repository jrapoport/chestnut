package log

import (
	"log"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var _ Logger = (*zap.SugaredLogger)(nil)

// NewZapLoggerWithLevel returns a new production zap logger with the log level.
func NewZapLoggerWithLevel(lvl Level) Logger {
	zlvl := levelToZapLevel(lvl)
	opt := zap.IncreaseLevel(zlvl)
	l, err := zap.NewProduction(opt)
	if err != nil {
		log.Fatal(err.Error())
		return nil
	}
	return l.Sugar()
}

func levelToZapLevel(lvl Level) zapcore.Level {
	switch lvl {
	case DebugLevel:
		return zapcore.DebugLevel
	case InfoLevel:
		return zapcore.InfoLevel
	case WarnLevel:
		return zapcore.WarnLevel
	case ErrorLevel:
		return zapcore.ErrorLevel
	case PanicLevel:
		return zapcore.PanicLevel
	case FatalLevel:
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}
