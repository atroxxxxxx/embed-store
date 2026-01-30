package importer

import (
	"fmt"
	"strings"
)

type Column struct {
	DocID      int
	Title      int
	Author     int
	Text       int
	Time       int
	Type       int
	Score      int
	Dead       int
	Deleted    int
	Vector     int
	ChunkStart int
	ChunkEnd   int
	ChunkNo    int
}

func IndexColumns(header []string) (Column, error) {
	columnIndex := make(map[string]int, len(header))
	for pos, name := range header {
		columnIndex[strings.TrimSpace(name)] = pos
	}

	var column Column
	need := map[string]*int{
		"doc_id":      &column.DocID,
		"title":       &column.Title,
		"author":      &column.Author,
		"text":        &column.Text,
		"time":        &column.Time,
		"type":        &column.Type,
		"score":       &column.Score,
		"dead":        &column.Dead,
		"deleted":     &column.Deleted,
		"vector":      &column.Vector,
		"chunk_start": &column.ChunkStart,
		"chunk_end":   &column.ChunkEnd,
		"chunk_no":    &column.ChunkNo,
	}

	for name, dst := range need {
		idx, ok := columnIndex[name]
		if !ok {
			return Column{}, fmt.Errorf("missing column %q", name)
		}
		*dst = idx
	}

	return column, nil
}
