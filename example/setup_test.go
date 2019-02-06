package example

import (
	"context"
	"fmt"
	"os"
	"testing"

	"bou.ke/tempdb/sqlite3"

	"bou.ke/orm/ctxdb"
)

var ctx context.Context

func TestMain(m *testing.M) {
	db, c, err := sqlite3.New()
	if err != nil {
		fmt.Printf("Failed to setup test DB: %v\n", err)
		os.Exit(1)
	}
	defer c()
	_, err = db.Exec(`
CREATE TABLE users (
  id INTEGER PRIMARY KEY,
  first_name TEXT NOT NULL,
  last_name  TEXT NOT NULL
);

CREATE TABLE posts (
  id INTEGER PRIMARY KEY,
  user_id INTEGER NOT NULL,
  body TEXT NOT NULL
);
`)
	if err != nil {
		fmt.Printf("Failed to run setup SQL: %v\n", err)
		os.Exit(2)
	}

	ctx = ctxdb.With(context.Background(), db)
	os.Exit(m.Run())
}
