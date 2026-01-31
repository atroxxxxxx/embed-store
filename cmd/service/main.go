package main

import (
	"context"
	"fmt"
	golog "log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/atroxxxxxx/embed-store/internal/cluster"
	database "github.com/atroxxxxxx/embed-store/internal/db"
	"github.com/atroxxxxxx/embed-store/internal/httpapi"
	"github.com/atroxxxxxx/embed-store/internal/importer"
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

	rootCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	connectCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := database.Connect(cfg.DSN, connectCtx)

	if err != nil {
		log.Fatal("failed to connect database: ", zap.Error(err))
	}
	defer db.DB.Close()
	log.Info("database successfully connected")

	handler, err := httpapi.New(&db, log)
	if err != nil {
		log.Fatal("handler error", zap.Error(err))
	}

	go func() {
		if cfg.RunImport {
			if err = importer.ExecImporter(rootCtx, &db, cfg, log, 20*time.Second); err != nil {
				log.Error("import aborted", zap.Error(err))
				return
			}
		}
		if cfg.RunCluster {
			if err = cluster.ExecCluster(rootCtx, &db, cfg.ClusterCfg, log); err != nil {
				log.Error("cluster aborted", zap.Error(err))
				return
			}
		}
	}()

	log.Info("http server started", zap.String("addr", cfg.HTTPAddr))
	if err = http.ListenAndServe(cfg.HTTPAddr, handler.Routes()); err != nil {
		log.Fatal("http server failed", zap.Error(err))
	}
}
