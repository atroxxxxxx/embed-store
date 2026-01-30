package importer

import (
	"context"
	"errors"
	"sync"

	"github.com/atroxxxxxx/embed-store/internal/db"
)

var (
	ErrInvalidArgs = errors.New("invalid function args")
)

func Run(ctx context.Context, repo Repo, config Config, stats *Stats) error {
	if repo == nil || stats == nil || config.Workers <= 0 || config.FilePath == "" {
		return ErrInvalidArgs
	}

	jobs := make(chan *db.Chunk, config.Workers*2)
	var waitGroup sync.WaitGroup
	waitGroup.Add(config.Workers)

	for range config.Workers {
		go func() {
			defer waitGroup.Done()
			runWorker(ctx, repo, jobs, stats, config.BatchSize)
		}()
	}

	err := ParseCSV(config.FilePath, jobs, stats, config.Limit)
	close(jobs)

	waitGroup.Wait()

	if err != nil {
		return err
	}

	return nil
}

func runWorker(ctx context.Context, repo Repo, jobs <-chan *db.Chunk, stats *Stats, batchSize int) {
	if batchSize == 0 {
		batchSize = 1
	}
	batch := make([]*db.Chunk, 0, batchSize)
	flush := func() {
		if len(batch) == 0 {
			return
		}

		inserted, err := repo.InsertBatch(ctx, batch)
		if err != nil {
			stats.Failed.Add(int64(len(batch)))
			batch = batch[:0]
			return
		}

		stats.Inserted.Add(inserted)
		stats.Duplicates.Add(int64(len(batch)) - inserted)
		batch = batch[:0]
	}

	for chunk := range jobs {
		if chunk == nil {
			stats.Failed.Add(1)
			continue
		}

		batch = append(batch, chunk)
		if len(batch) >= batchSize {
			flush()
		}
	}
	flush()
}
