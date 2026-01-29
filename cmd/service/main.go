package main

import (
	"context"
	"fmt"
	"os"
	"time"

	database "github.com/atroxxxxxx/embed-store/internal/db"
	"github.com/atroxxxxxx/embed-store/internal/logger"
	"github.com/atroxxxxxx/embed-store/internal/runcfg"
	_ "github.com/golang-migrate/migrate/v4"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pgvector/pgvector-go"
	"go.uber.org/zap"
)

func main() {
	cfg, err := runcfg.Parse()
	if err != nil { //TODO: log.fatal
		panic("flag parsing: " + err.Error())
	}

	log, err := logger.New(cfg.LogLevel)
	if err != nil {
		panic("log creation: " + err.Error())
	}

	defer func(log *zap.Logger) {
		err := log.Sync()
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
		}
	}(log)

	log.Debug("log init")
	log.Info("trying to connect to database")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	start := time.Now()

	db, err := database.Connect(cfg.DSN, ctx)

	if err != nil {
		log.Fatal("failed to connect database: ", zap.Error(err))
	}
	defer db.DB.Close()

	duration := time.Since(start)
	log.Info("database successfully connected", zap.Duration("duration: ", duration))

	if err = db.DB.Ping(); err != nil {
		log.Error("database connection failed", zap.Error(err))
	}

	c := database.Chunk{
		DocID:     1,
		Title:     nil,
		Author:    nil,
		Text:      "hello",
		Time:      start,
		Type:      "comment",
		Score:     -1,
		Deleted:   false,
		Dead:      true,
		Embedding: pgvector.NewVector(make([]float32, database.VectorSize)),
		Info: database.Metadata{
			Number: 1,
			Start:  1,
			End:    2,
		},
	}

	id, err := db.InsertChunk(context.Background(), &c)
	if err != nil {
		log.Error("can't insert chunk", zap.Error(err))
	} else {
		log.Info("successfully inserted", zap.Int64("ID: ", id))
	}
	chunk, err := db.ChunkById(context.Background(), 1)
	if err != nil {
		log.Error("can't find chunk", zap.Error(err))
	} else {
		log.Info("chunk found", zap.Any("chunk: ", *chunk))
	}
}
