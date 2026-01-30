package importer

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/atroxxxxxx/embed-store/internal/db"
)

type Parser struct{}

const csvTimeLayout = "2006-01-02 15:04:05.000"

func (Parser) NullableString(fieldName string, value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func (Parser) Int64(fieldName string, value string) (int64, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return 0, fmt.Errorf("%s: empty integer", fieldName)
	}
	parsed, err := strconv.ParseInt(trimmed, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", fieldName, err)
	}
	return parsed, nil
}

func (Parser) Bool01(fieldName string, value string) (bool, error) {
	trimmed := strings.TrimSpace(value)
	switch trimmed {
	case "0":
		return false, nil
	case "1":
		return true, nil
	default:
		return false, fmt.Errorf("%s: expected 0/1, got %q", fieldName, trimmed)
	}
}

func (Parser) Time(fieldName string, value string) (time.Time, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return time.Time{}, fmt.Errorf("%s: empty time", fieldName)
	}
	parsed, err := time.Parse(csvTimeLayout, trimmed)
	if err != nil {
		return time.Time{}, fmt.Errorf("%s: %w", fieldName, err)
	}
	return parsed, nil
}

func (Parser) TypeFromInt(fieldName string, value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	switch trimmed {
	case "1":
		return "story", nil
	case "2":
		return "comment", nil
	case "3":
		return "poll", nil
	case "4":
		return "pollopt", nil
	case "5":
		return "job", nil
	default:
		return "", fmt.Errorf("%s: unknown type %q (expected 1 or 2)", fieldName, trimmed)
	}
}

func (Parser) Vector384(fieldName string, value string) ([]float32, error) {
	trimmed := strings.TrimSpace(value)
	if len(trimmed) < 2 || trimmed[0] != '[' || trimmed[len(trimmed)-1] != ']' {
		return nil, fmt.Errorf("%s: vector not in [..] format", fieldName)
	}

	body := trimmed[1 : len(trimmed)-1]
	if body == "" {
		return nil, fmt.Errorf("%s: empty vector", fieldName)
	}

	parts := strings.Split(body, ",")
	if len(parts) != db.VectorSize {
		return nil, fmt.Errorf("%s: vector length %d != %d", fieldName, len(parts), db.VectorSize)
	}

	vector := make([]float32, db.VectorSize)
	for index := 0; index < db.VectorSize; index++ {
		number, err := strconv.ParseFloat(parts[index], 32)
		if err != nil {
			return nil, fmt.Errorf("%s: vector[%d]: %w", fieldName, index, err)
		}
		vector[index] = float32(number)
	}
	return vector, nil
}
