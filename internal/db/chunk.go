package db

import (
	"time"

	"github.com/pgvector/pgvector-go"
)

const VectorSize = 384

type Chunk struct {
	ID        int64
	DocID     int64
	Title     *string
	Author    *string
	Text      string
	Time      time.Time
	Type      string
	Score     int32
	Deleted   bool
	Dead      bool
	Embedding pgvector.Vector
	Info      Info
}

type Info struct {
	Number uint32
	Start  uint64
	End    uint64
}
