package importer

import (
	"fmt"

	"github.com/atroxxxxxx/embed-store/internal/db"
	"github.com/pgvector/pgvector-go"
)

func ParseRow(record []string, columns Column, parser Parser) (*db.Chunk, error) {
	docID, err := parser.Int64("doc_id", record[columns.DocID])
	if err != nil {
		return nil, err
	}

	title := parser.NullableString("title", record[columns.Title])
	author := parser.NullableString("author", record[columns.Author])

	text := record[columns.Text]
	if text == "" {
		return nil, fmt.Errorf("text: empty")
	}

	parsedTime, err := parser.Time("time", record[columns.Time])
	if err != nil {
		return nil, err
	}

	parsedType, err := parser.TypeFromInt("type", record[columns.Type])
	if err != nil {
		return nil, err
	}

	score64, err := parser.Int64("score", record[columns.Score])
	if err != nil {
		return nil, err
	}

	dead, err := parser.Bool01("dead", record[columns.Dead])
	if err != nil {
		return nil, err
	}

	deleted, err := parser.Bool01("deleted", record[columns.Deleted])
	if err != nil {
		return nil, err
	}

	vectorSlice, err := parser.Vector384("vector", record[columns.Vector])
	if err != nil {
		return nil, err
	}

	chunkStart, err := parser.Int64("chunk_start", record[columns.ChunkStart])
	if err != nil {
		return nil, err
	}

	chunkEnd, err := parser.Int64("chunk_end", record[columns.ChunkEnd])
	if err != nil {
		return nil, err
	}

	chunkNo64, err := parser.Int64("chunk_no", record[columns.ChunkNo])
	if err != nil {
		return nil, err
	}

	return &db.Chunk{
		DocID:     docID,
		Title:     title,
		Author:    author,
		Text:      text,
		Time:      parsedTime,
		Type:      parsedType,
		Score:     int32(score64),
		Deleted:   deleted,
		Dead:      dead,
		Embedding: pgvector.NewVector(vectorSlice),
		Info: db.Metadata{
			Number: int32(chunkNo64),
			Start:  chunkStart,
			End:    chunkEnd,
		},
	}, nil
}
