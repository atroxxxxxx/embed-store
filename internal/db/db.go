package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type Database struct {
	DB *sql.DB
}

const (
	SQLDriver     string = "pgx"
	uniqueErrCode        = "23505"
)

var (
	ErrChunkNil     = errors.New("chunk is null")
	ErrDuplicateKey = errors.New("duplicate key")
)

func Connect(dsn string, ctx context.Context) (Database, error) {
	db, err := sql.Open(SQLDriver, dsn)
	if err != nil {
		return Database{}, err
	}
	for {
		err = db.PingContext(ctx)
		if err == nil {
			return Database{DB: db}, nil
		}
		select {
		case <-time.After(500 * time.Millisecond):
		case <-ctx.Done():
			db.Close()
			return Database{}, ctx.Err()
		}
	}
}

func (obj *Database) InsertChunk(ctx context.Context, chunk *Chunk) (int64, error) {
	const sqlResponse = "INSERT INTO hackernews " +
		"(doc_id, title, author, text, time, type, score, deleted, dead, embedding, chunk_no, chunk_start, chunk_end) " +
		"VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13) RETURNING id"
	if chunk == nil {
		return 0, ErrChunkNil
	}

	row := obj.DB.QueryRowContext(ctx, sqlResponse,
		chunk.DocID, chunk.Title, chunk.Author, chunk.Text, chunk.Time, chunk.Type, chunk.Score, chunk.Deleted, chunk.Dead,
		chunk.Embedding, chunk.Info.Number, chunk.Info.Start, chunk.Info.End)
	if err := row.Scan(&chunk.ID); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == uniqueErrCode {
				return 0, ErrDuplicateKey
			}
		}
		return 0, fmt.Errorf("insert failed: %w", err)
	}
	return chunk.ID, nil
}

func (obj *Database) InsertBatch(ctx context.Context, batch []*Chunk) (int64, error) {
	if len(batch) == 0 {
		return 0, nil
	}

	const columnsPerRow = 13
	var queryBuilder strings.Builder
	queryBuilder.Grow(256 + len(batch)*columnsPerRow*6)
	queryBuilder.WriteString(
		"INSERT INTO hackernews " +
			"(doc_id, title, author, text, time, type, score, deleted, dead, embedding, " +
			"chunk_no, chunk_start, chunk_end) VALUES ",
	)

	args := make([]any, 0, len(batch)*columnsPerRow)
	argIdx := 1
	for rowIdx, chunk := range batch {
		if chunk == nil {
			return 0, fmt.Errorf("nil chunk in batch. index: %d", argIdx)
		}
		if rowIdx > 0 {
			queryBuilder.WriteByte(',')
		}

		queryBuilder.WriteByte('(')
		for idx := range columnsPerRow {
			if idx > 0 {
				queryBuilder.WriteByte(',')
			}
			queryBuilder.WriteByte('$')
			queryBuilder.WriteString(fmt.Sprint(argIdx))
			argIdx++
		}
		queryBuilder.WriteByte(')')

		args = append(args,
			chunk.DocID,
			chunk.Title,
			chunk.Author,
			chunk.Text,
			chunk.Time,
			chunk.Type,
			chunk.Score,
			chunk.Deleted,
			chunk.Dead,
			chunk.Embedding,
			chunk.Info.Number,
			chunk.Info.Start,
			chunk.Info.End,
		)
	}

	queryBuilder.WriteString(" ON CONFLICT (doc_id, chunk_no) DO NOTHING")

	result, err := obj.DB.ExecContext(ctx, queryBuilder.String(), args...)
	if err != nil {
		return 0, fmt.Errorf("batch insert failed: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("rows affected: %w", err)
	}

	return affected, nil
}

func (obj *Database) ChunkByID(ctx context.Context, id int64) (*Chunk, error) {
	const response = "SELECT doc_id, title, author, text, time, type, score, " +
		"deleted, dead, embedding, chunk_no, chunk_start, chunk_end FROM hackernews WHERE id = $1"
	row := obj.DB.QueryRowContext(ctx, response, id)
	chunk := Chunk{}
	if err := row.Scan(&chunk.DocID, &chunk.Title, &chunk.Author, &chunk.Text,
		&chunk.Time, &chunk.Type, &chunk.Score, &chunk.Deleted, &chunk.Dead, &chunk.Embedding,
		&chunk.Info.Number, &chunk.Info.Start, &chunk.Info.End); err != nil {
		return nil, fmt.Errorf("id %d not found: %w", id, err)
	}
	chunk.ID = id
	return &chunk, nil
}
