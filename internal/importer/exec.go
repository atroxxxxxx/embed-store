package importer

import (
	"context"
	"time"

	database "github.com/atroxxxxxx/embed-store/internal/db"
	"github.com/atroxxxxxx/embed-store/internal/runcfg"
	"go.uber.org/zap"
)

func Exec(ctx context.Context, db *database.Database, cfg runcfg.RunConfig, log *zap.Logger, tickTime time.Duration) {
	stats := &Stats{}
	importCtx, cansel := context.WithCancel(ctx)

	go func() {
		defer cansel()
		log.Info("import started",
			zap.String("file", cfg.ImportCfg.FilePath),
			zap.Int("workers", cfg.ImportCfg.Workers),
			zap.Int("batch size", cfg.ImportCfg.BatchSize),
			zap.Int("limit", cfg.ImportCfg.Limit),
		)

		start := time.Now()
		err := Run(ctx, db, cfg.ImportCfg, stats)
		duration := time.Since(start)
		if err != nil {
			log.Error("import failed", zap.Duration("duration", duration), zap.Error(err))
		}

		log.Info("import finished",
			zap.Duration("duration", duration),
			zap.Int64("read", stats.Read.Load()),
			zap.Int64("inserted", stats.Inserted.Load()),
			zap.Int64("duplicates", stats.Duplicates.Load()),
			zap.Int64("failed", stats.Failed.Load()),
		)
	}()

	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-importCtx.Done():
				return
			case <-ticker.C:
				log.Info("import progres",
					zap.Int64("read", stats.Read.Load()),
					zap.Int64("inserted", stats.Inserted.Load()),
					zap.Int64("duplicates", stats.Duplicates.Load()),
					zap.Int64("failed", stats.Failed.Load()),
				)

			}
		}
	}()

}
