package main

import (
	"fmt"
	"os"

	"github.com/atroxxxxxx/embed-store/internal/config"
	"github.com/atroxxxxxx/embed-store/internal/logger"
	"go.uber.org/zap"
)

func main() {
	cfg, err := config.New()
	if err != nil {
		panic("flag parsing: " + err.Error())
	}
	log, err := logger.New(cfg.LogLevel)
	if err != nil {
		panic("log creation: " + err.Error())
	}
	defer func(log *zap.Logger) {
		err := log.Sync()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}(log)
	log.Info("log init")
	log.Info("log level: " + cfg.LogLevel)
	log.Debug("debug")
}
