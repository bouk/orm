package example

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"bou.ke/ctxdb"
	"bou.ke/tempdb/sqlite3"
)

var ctx context.Context

func TestMain(m *testing.M) {
	db, c, err := sqlite3.New()
	if err != nil {
		fmt.Printf("Failed to setup test DB: %v\n", err)
		os.Exit(1)
	}
	defer c()
	migration, err := ioutil.ReadFile("./migrations/000000_schema/up.sql")
	if err != nil {
		fmt.Printf("Failed to read migration: %v\n", err)
		os.Exit(3)
	}
	_, err = db.Exec(string(migration))
	if err != nil {
		fmt.Printf("Failed to run setup SQL: %v\n", err)
		os.Exit(2)
	}

	ctx = ctxdb.With(context.Background(), db)
	os.Exit(m.Run())
}
