package importer

import (
	"context"
	"sync/atomic"

	"github.com/atroxxxxxx/embed-store/internal/db"
)

type Repo interface {
	InsertChunk(ctx context.Context, chunk *db.Chunk) (int64, error)
}

type Config struct {
	FilePath  string
	Workers   int
	BatchSize int
	Limit     int
}

type Stats struct {
	Read       atomic.Int64
	Inserted   atomic.Int64
	Duplicates atomic.Int64
	Failed     atomic.Int64
}
