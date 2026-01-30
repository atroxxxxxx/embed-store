package httpapi

import (
	"context"
	"errors"
	"net/http"

	database "github.com/atroxxxxxx/embed-store/internal/db"
	"github.com/pgvector/pgvector-go"
	"go.uber.org/zap"
)

type Repo interface {
	// InsertChunk inserts chunk to db
	InsertChunk(ctx context.Context, chunk *database.Chunk) (int64, error)
	// ChunkByID searches chunk by ID in db
	ChunkByID(ctx context.Context, id int64) (*database.Chunk, error)
	Search(ctx context.Context, vec *pgvector.Vector, limit int) ([]*database.Chunk, error)
}

type Handler struct {
	db     Repo
	logger *zap.Logger
}

var (
	ErrNullArgs = errors.New("null constructor arguments")
)

func New(db Repo, logger *zap.Logger) (*Handler, error) {
	if db == nil || logger == nil {
		return nil, ErrNullArgs
	}

	return &Handler{
		db:     db,
		logger: logger,
	}, nil
}

func (obj *Handler) Routes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/chunks", obj.post)
	mux.HandleFunc("/chunks/", obj.get)
	mux.HandleFunc("/search", obj.search)
	return mux
}
