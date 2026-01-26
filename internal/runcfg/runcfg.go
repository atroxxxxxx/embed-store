package runcfg

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/atroxxxxxx/embed-store/internal/logger"
)

const DefaultHTTP = ":8080"

var ErrDSNEmpty = errors.New("incomplete db config")

type RunConfig struct {
	DSN        string
	HTTPAddr   string
	LogLevel   string
	RunImport  bool
	RunCluster bool
}

func Parse() (RunConfig, error) {
	temp, err := parseEnv()
	if err != nil {
		return temp, fmt.Errorf(".env parsing failed: %w", err)
	}

	var (
		addr       = flag.String("http-addr", temp.HTTPAddr, "http address")
		logLevel   = flag.String("log-level", temp.LogLevel, "log message level")
		runImport  = flag.Bool("run-import", temp.RunImport, "")
		runCluster = flag.Bool("run-cluster", temp.RunCluster, "")
		isDebug    = flag.Bool("d", false, "debug log level shortcut")
	)
	flag.Parse()

	if *isDebug {
		*logLevel = logger.Debug
	}

	return RunConfig{
			DSN:        temp.DSN,
			HTTPAddr:   *addr,
			LogLevel:   *logLevel,
			RunImport:  *runImport,
			RunCluster: *runCluster,
		},
		nil
}

func parseEnv() (RunConfig, error) {
	var cfg RunConfig

	cfg.HTTPAddr = os.Getenv("HTTP_ADDR")
	if cfg.HTTPAddr == "" {
		cfg.HTTPAddr = DefaultHTTP
	}

	cfg.LogLevel = os.Getenv("LOG_LEVEL")
	if cfg.LogLevel == "" {
		cfg.LogLevel = logger.Info
	}

	// TODO: remove code repeating
	if envFlag := os.Getenv("RUN_IMPORT"); envFlag != "" {
		boolFlag, err := strconv.ParseBool(envFlag)
		if err != nil {
			return cfg, fmt.Errorf("invalid RUN_IMPORT: %w", err)
		}
		cfg.RunImport = boolFlag
	}

	if envFlag := os.Getenv("RUN_CLUSTER"); envFlag != "" {
		boolFlag, err := strconv.ParseBool(envFlag)
		if err != nil {
			return cfg, fmt.Errorf("invalid RUN_CLUSTER: %w", err)
		}
		cfg.RunCluster = boolFlag
	}

	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	name := os.Getenv("DB_NAME")
	sslmode := os.Getenv("DB_SSLMODE")

	if host == "" || port == "" || user == "" || name == "" {
		return cfg, ErrDSNEmpty
	}

	if sslmode == "" {
		sslmode = "disable"
	}

	cfg.DSN = fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, name, sslmode,
	)

	return cfg, nil
}
