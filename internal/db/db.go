package db

import (
	"context"
	"database/sql"
	"time"
)

const SQLDriver string = "pgx"

func Connect(dsn string, ctx context.Context) (*sql.DB, error) {
	db, err := sql.Open(SQLDriver, dsn)
	if err != nil {
		return nil, err
	}
	for {
		err = db.PingContext(ctx)
		if err == nil {
			return db, nil
		}
		select {
		case <-time.After(500 * time.Millisecond):
		case <-ctx.Done():
			db.Close()
			return nil, ctx.Err()
		}
	}
}
