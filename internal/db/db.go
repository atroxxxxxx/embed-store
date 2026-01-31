package db

import (
	"context"
	"database/sql"
	"errors"
	"time"

	_ "github.com/jackc/pgx/v5"
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
