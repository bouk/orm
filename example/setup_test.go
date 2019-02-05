package example

import (
	"bou.ke/tempdb/sqlite3"
	"context"
	"fmt"

	"os"
	"testing"
)

func Ctx() context.Context {
}

func TestMain(m *testing.M) {
	db, c, err := sqlite3.New()
	if err != nil {
		fmt.Printf("Failed to setup test DB: %v\n", err)
		os.Exit(1)
	}
	defer c()
	os.Exit(m.Run())
}
