package importer

import (
	"context"
	"time"

	database "github.com/atroxxxxxx/embed-store/internal/db"
	"github.com/atroxxxxxx/embed-store/internal/runcfg"
	"go.uber.org/zap"
)

func ExecImporter(
	ctx context.Context,
	db *database.Database,
	cfg runcfg.RunConfig,
	log *zap.Logger,
	tickTime time.Duration,
) error {

	stats := &Stats{}

	log.Info("import started",
		zap.String("file", cfg.ImportCfg.FilePath),
		zap.Int("workers", cfg.ImportCfg.Workers),
		zap.Int("batch size", cfg.ImportCfg.BatchSize),
		zap.Int("limit", cfg.ImportCfg.Limit),
	)

	importCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		ticker := time.NewTicker(tickTime)
		defer ticker.Stop()

		for {
			select {
			case <-importCtx.Done():
				return
			case <-ticker.C:
				log.Info("import progress",
					zap.Int64("read", stats.Read.Load()),
					zap.Int64("inserted", stats.Inserted.Load()),
					zap.Int64("duplicates", stats.Duplicates.Load()),
					zap.Int64("failed", stats.Failed.Load()),
				)
			}
		}
	}()

	start := time.Now()
	err := Run(importCtx, db, cfg.ImportCfg, stats)
	duration := time.Since(start)

	if err != nil {
		log.Error("import failed",
			zap.Duration("duration", duration),
			zap.Error(err),
		)
		return err
	}

	log.Info("import finished",
		zap.Duration("duration", duration),
		zap.Int64("read", stats.Read.Load()),
		zap.Int64("inserted", stats.Inserted.Load()),
		zap.Int64("duplicates", stats.Duplicates.Load()),
		zap.Int64("failed", stats.Failed.Load()),
	)

	return nil
}
