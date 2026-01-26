package logger

import (
	"errors"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	Info  string = "info"
	Warn  string = "warn"
	Error string = "error"
	Debug string = "debug"
)

var ErrUnknownLogLevel = errors.New("unknown log level")

func New(level string) (*zap.Logger, error) {
	var logLevel zapcore.Level
	level = strings.ToLower(level)
	switch level {
	case Info:
		logLevel = zapcore.InfoLevel
	case Warn:
		logLevel = zapcore.WarnLevel
	case Error:
		logLevel = zapcore.ErrorLevel
	case Debug:
		logLevel = zapcore.DebugLevel
	default:
		return nil, ErrUnknownLogLevel
	}
	config := zap.NewProductionConfig()

	config.Level.SetLevel(logLevel)
	if logLevel == zapcore.DebugLevel {
		config.Development = true
	} else {
		config.DisableCaller = true
		config.DisableStacktrace = true
	}
	logger, err := config.Build()
	if err != nil {
		return nil, err
	}
	return logger, nil
}
