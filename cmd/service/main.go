package main

import (
	"context"
	"fmt"
	golog "log"
	"net/http"
	"os"
	"time"

	database "github.com/atroxxxxxx/embed-store/internal/db"
	"github.com/atroxxxxxx/embed-store/internal/httpapi"
	"github.com/atroxxxxxx/embed-store/internal/logger"
	"github.com/atroxxxxxx/embed-store/internal/runcfg"
	_ "github.com/golang-migrate/migrate/v4"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

func main() {
	cfg, err := runcfg.Parse()
	if err != nil {
		golog.Fatal("flag parsing", err)
	}

	log, err := logger.New(cfg.LogLevel)
	if err != nil {
		golog.Fatal("log init", err)
	}

	defer func(log *zap.Logger) {
		err := log.Sync()
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
		}
	}(log)

	log.Debug("log init")
	log.Info("trying to connect to database")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	start := time.Now()

	db, err := database.Connect(cfg.DSN, ctx)

	if err != nil {
		log.Fatal("failed to connect database: ", zap.Error(err))
	}
	defer db.DB.Close()

	duration := time.Since(start)
	log.Info("database successfully connected", zap.Duration("duration", duration))

	handler, err := httpapi.New(&db, log)
	if err != nil {
		log.Fatal("handler error", zap.Error(err))
	}
	log.Info("http server started", zap.String("addr", cfg.HTTPAddr))
	err = http.ListenAndServe(cfg.HTTPAddr, handler.Routes())
	if err != nil {
		log.Fatal("something went wrong :(", zap.Error(err))
	}
}
