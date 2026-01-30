package importer

import (
	"encoding/csv"
	"fmt"
	"os"
)

func ReadFile(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("file not open: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1

	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("read header: %w", err)
	}
	return header, nil
}
