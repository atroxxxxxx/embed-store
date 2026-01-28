package db

import (
	"context"
	"database/sql"
	"time"
)

type Database struct {
	DB *sql.DB
}

const SQLDriver string = "pgx"

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
