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
	DSN       string
	HTTPAddr  string
	LogLevel  string
	RunImport bool
	ImportCfg struct {
		FilePath  string
		Workers   int
		BatchSize int
		Limit     int
	}
	RunCluster bool
	ClusterCfg struct {
		Clusters  int
		Iters     int
		Workers   int
		Limit     int
		BatchSize int
	}
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
			ImportCfg:  temp.ImportCfg,
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

	if envFlag := os.Getenv("RUN_IMPORT"); envFlag != "" {
		boolFlag, err := strconv.ParseBool(envFlag)
		if err != nil {
			return cfg, fmt.Errorf("invalid RUN_IMPORT: %w", err)
		}
		cfg.RunImport = boolFlag
	}

	if cfg.RunImport {
		cfg.ImportCfg.FilePath = os.Getenv("IMPORT_FILE")

		cfg.ImportCfg.Workers = getEnvCount("IMPORT_WORKERS", 4)
		cfg.ImportCfg.BatchSize = getEnvCount("IMPORT_BATCH_SIZE", 200)
		cfg.ImportCfg.Limit = getEnvCount("IMPORT_LIMIT", 0)
	}

	if envFlag := os.Getenv("RUN_CLUSTER"); envFlag != "" {
		boolFlag, err := strconv.ParseBool(envFlag)
		if err != nil {
			return cfg, fmt.Errorf("invalid RUN_CLUSTER: %w", err)
		}
		cfg.RunCluster = boolFlag
	}

	if cfg.RunCluster {
		cfg.ClusterCfg.Clusters = getEnvCount("CLUSTER_COUNT", 64)
		cfg.ClusterCfg.Iters = getEnvCount("CLUSTER_ITERS", 10)
		cfg.ClusterCfg.Workers = getEnvCount("CLUSTER_WORKERS", 6)
		cfg.ClusterCfg.Limit = getEnvCount("CLUSTER_LIMIT", 20000)
		cfg.ClusterCfg.BatchSize = getEnvCount("CLUSTER_BATCH_SIZE", 1000)
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

func getEnvCount(key string, def int) int {
	value := os.Getenv(key)
	if value == "" {
		return def
	}
	count, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return def
	}
	return int(count)
}
