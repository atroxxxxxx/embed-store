package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/atroxxxxxx/embed-store/internal/config"
	database "github.com/atroxxxxxx/embed-store/internal/db"
	"github.com/atroxxxxxx/embed-store/internal/logger"
	_ "github.com/jackc/pgx/v5/stdlib"
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
			_, _ = fmt.Fprintln(os.Stderr, err)
		}
	}(log)

	log.Debug("log init")
	log.Info("trying to connect to database")

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	start := time.Now()

	db, err := database.Connect(cfg.DSN, ctx)

	if err != nil {
		log.Fatal("failed to connect database: ", zap.Error(err))
	}
	defer db.Close()

	duration := time.Since(start)
	log.Info("database successfully connected", zap.Duration("duration: ", duration))

	if err = db.Ping(); err != nil {
		log.Error("database connection failed", zap.Error(err))
	}
}
