package cluster

import (
	"context"
	"fmt"
	"time"

	"github.com/atroxxxxxx/embed-store/internal/db"
	"go.uber.org/zap"
)

func Run(ctx context.Context, database *db.Database, cfg ClusterConfig, log *zap.Logger) error {
	if cfg.Clusters <= 0 {
		cfg.Clusters = 64
	}
	if cfg.Iters <= 0 {
		cfg.Iters = 10
	}
	if cfg.Workers <= 0 {
		cfg.Workers = 4
	}
	if cfg.Limit <= 0 {
		cfg.Limit = 20000
	}
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = 1000
	}

	log.Info("clusterization started",
		zap.Int("clusters", cfg.Clusters),
		zap.Int("iters", cfg.Iters),
		zap.Int("workers", cfg.Workers),
		zap.Int("limit", cfg.Limit),
		zap.Int("batch_size", cfg.BatchSize),
	)

	start := time.Now()
	points, err := database.ClusterSource(ctx, cfg.Limit)
	if err != nil {
		return fmt.Errorf("cluster source: %w", err)
	}
	if len(points) == 0 {
		log.Warn("cluster source returned 0 rows")
		return nil
	}

	ids := make([]int64, len(points))
	vectors := make([][]float32, len(points))
	for i, point := range points {
		ids[i] = point.ID
		vectors[i] = point.Embedding.Slice()
	}

	assignments, err := kMeans(vectors, cfg)
	if err != nil {
		return fmt.Errorf("kmeans: %w", err)
	}

	for startIdx := 0; startIdx < len(ids); startIdx += cfg.BatchSize {
		endIdx := startIdx + cfg.BatchSize
		if endIdx > len(ids) {
			endIdx = len(ids)
		}

		if err := database.UpdateClusterIDs(ctx, ids[startIdx:endIdx], assignments[startIdx:endIdx]); err != nil {
			return fmt.Errorf("update cluster ids [%d:%d]: %w", startIdx, endIdx, err)
		}

		log.Debug("cluster ids updated", zap.Int("from", startIdx), zap.Int("to", endIdx))
	}

	log.Info("clusterization finished",
		zap.Int("rows", len(ids)),
		zap.Duration("duration", time.Since(start)),
	)
	return nil
}
