package cluster

import (
	"context"
	"time"

	database "github.com/atroxxxxxx/embed-store/internal/db"
	"go.uber.org/zap"
)

func ExecCluster(
	ctx context.Context,
	db *database.Database,
	cfg ClusterConfig,
	log *zap.Logger,
) error {

	log.Info("clusterization started",
		zap.Int("clusters", cfg.Clusters),
		zap.Int("iters", cfg.Iters),
		zap.Int("workers", cfg.Workers),
	)

	clusterCtx, cancel := context.WithTimeout(ctx, 30*time.Minute)
	defer cancel()

	if err := Run(clusterCtx, db, cfg, log); err != nil {
		log.Error("clusterization failed", zap.Error(err))
		return err
	}

	log.Info("clusterization finished")
	return nil
}
