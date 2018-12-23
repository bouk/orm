package db

import (
	"context"
	"database/sql"
)

type dbKey struct{}

func getDB(ctx context.Context) *sql.DB {
	db, _ := ctx.Value(dbKey{}).(*sql.DB)
	return db
}
