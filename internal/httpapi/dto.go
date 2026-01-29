package httpapi

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/atroxxxxxx/embed-store/internal/db"
	"github.com/pgvector/pgvector-go"
)

type Request struct {
	DocId      int64     `json:"doc_id"`
	Title      *string   `json:"title"`
	Author     *string   `json:"author"`
	Text       string    `json:"text"`
	Time       string    `json:"time"`
	Type       string    `json:"type"`
	Score      int32     `json:"score"`
	Deleted    bool      `json:"deleted"`
	Dead       bool      `json:"dead"`
	Embedding  []float32 `json:"embedding"`
	ChunkNo    int32     `json:"chunk_no"`
	ChunkStart int64     `json:"chunk_start"`
	ChunkEnd   int64     `json:"chunk_end"`
}

type Response struct {
	ID        int64       `json:"id"`
	DocId     int64       `json:"doc_id"`
	Title     *string     `json:"title,omitempty"`
	Author    *string     `json:"author,omitempty"`
	Text      string      `json:"text"`
	Time      string      `json:"time"`
	Type      string      `json:"type"`
	Score     int32       `json:"score"`
	Deleted   bool        `json:"deleted"`
	Dead      bool        `json:"dead"`
	Embedding *[]float32  `json:"embedding,omitempty"`
	Info      db.Metadata `json:"chunk_metadata"`
}

var (
	ErrInvalidType         = errors.New("undefined type")
	ErrInvalidEmbeddingLen = errors.New("invalid embedding length")
	ErrChunkNull           = errors.New("chunk is null")
	ErrRequestNull         = errors.New("request is null")
)

const (
	timeLayout = time.RFC3339
	story      = "story"
	comment    = "comment"
)

func Map(request *Request) (*db.Chunk, error) {
	if request == nil {
		return nil, ErrRequestNull
	}

	requestTime, err := time.Parse(timeLayout, request.Time)
	if err != nil {
		return nil, fmt.Errorf("time parsing: %w", err)
	}

	requestType := strings.TrimSpace(strings.ToLower(request.Type))
	if requestType != story && requestType != comment {
		return nil, ErrInvalidType
	}

	if len(request.Embedding) != db.VectorSize {
		return nil, ErrInvalidEmbeddingLen
	}

	return &db.Chunk{
		DocID:     request.DocId,
		Title:     request.Title,
		Author:    request.Author,
		Text:      request.Text,
		Time:      requestTime,
		Type:      requestType,
		Score:     request.Score,
		Deleted:   request.Deleted,
		Dead:      request.Dead,
		Embedding: pgvector.NewVector(request.Embedding),
		Info: db.Metadata{
			Number: request.ChunkNo,
			Start:  request.ChunkStart,
			End:    request.ChunkEnd,
		},
	}, nil
}

func Unmap(chunk *db.Chunk, withEmbedding bool) (Response, error) {
	if chunk == nil {
		return Response{}, ErrChunkNull
	}

	var vec *[]float32 = nil
	if withEmbedding {
		tmp := chunk.Embedding.Slice()
		vec = &tmp
	}

	return Response{
		ID:        chunk.ID,
		DocId:     chunk.DocID,
		Title:     chunk.Title,
		Author:    chunk.Author,
		Text:      chunk.Text,
		Time:      chunk.Time.Format(timeLayout),
		Type:      chunk.Type,
		Score:     chunk.Score,
		Deleted:   chunk.Deleted,
		Dead:      chunk.Dead,
		Embedding: vec,
		Info:      chunk.Info,
	}, nil
}
