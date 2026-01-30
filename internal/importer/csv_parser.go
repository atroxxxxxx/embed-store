package importer

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"

	"github.com/atroxxxxxx/embed-store/internal/db"
)

func ParseCSV(path string, out chan<- *db.Chunk, stats *Stats, limit int) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open csv: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1

	header, err := reader.Read()
	if err != nil {
		return fmt.Errorf("read header: %w", err)
	}

	columns, err := IndexColumns(header)
	if err != nil {
		return fmt.Errorf("index columns: %w", err)
	}

	parser := Parser{}

	var parsed int
	for {
		if limit > 0 && parsed >= limit {
			return nil
		}

		record, err := reader.Read()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return fmt.Errorf("read row: %w", err)
		}

		chunk, err := ParseRow(record, columns, parser)
		if err != nil {
			return fmt.Errorf("parse row %d: %w", parsed+1, err)
		}

		stats.Read.Add(1)
		out <- chunk
		parsed++
	}
}
