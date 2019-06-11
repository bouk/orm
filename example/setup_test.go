package example

import (
	"context"
	"database/sql"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"bou.ke/tempdb/sqlite3"
)

var d *sql.DB
var ctx = context.Background()

func TestMain(m *testing.M) {
	var c func()
	var err error
	d, c, err = sqlite3.New()
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
	_, err = d.Exec(string(migration))
	if err != nil {
		fmt.Printf("Failed to run setup SQL: %v\n", err)
		os.Exit(2)
	}

	os.Exit(m.Run())
}

func clear() {
	d.Exec("DELETE FROM users")
	d.Exec("DELETE FROM posts")
}
