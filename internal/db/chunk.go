package db

import "time"

const VectorSize = 384

type Chunk struct {
	ID         int64
	DocID      int64
	Title      *string
	Author     *string
	Text       string
	Time       time.Time
	Type       string
	Score      int32
	Deleted    bool
	Dead       bool
	Embedding  [VectorSize]float32
	ChunkNo    int32
	ChunkStart int64
	ChunkEnd   int64
}
