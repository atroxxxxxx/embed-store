package config

import (
	"errors"
	"flag"

	"github.com/atroxxxxxx/embed-store/internal/logger"
)

const DefaultHTTP = ":8080"

var (
	ErrDSNEmpty        = errors.New("empty dsn")
	ErrUnknownLogLevel = errors.New("unknown log level")
)

type Config struct {
	DSN        string
	HTTPAddr   string
	LogLevel   string
	RunImport  bool
	RunCluster bool
}

func New() (Config, error) {
	var ( // TODO: norm descriptions
		dsn        = flag.String("dsn", "", "database path")
		addr       = flag.String("http-addr", DefaultHTTP, "http address")
		logLevel   = flag.String("log-level", logger.Info, "log message level")
		runImport  = flag.Bool("run-import", true, "")
		runCluster = flag.Bool("run-cluster", false, "")
		isDebug    = flag.Bool("d", false, "debug log level shortcut")
	)
	flag.Parse()
	if *dsn == "" {
		return Config{}, ErrDSNEmpty
	}
	if *isDebug {
		*logLevel = logger.Debug
	} else {
		switch *logLevel {
		case logger.Info:
		case logger.Warn:
		case logger.Error:
		case logger.Debug:
		default:
			return Config{}, ErrUnknownLogLevel
		}
	}
	return Config{*dsn, *addr, *logLevel, *runImport, *runCluster}, nil
}
