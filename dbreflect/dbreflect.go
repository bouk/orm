package dbreflect

import (
	"context"
)

type Reflecter interface {
	DescribeTable(ctx context.Context, q querier, tableName string) (*Table, error)
	ListTables() []string
}

type Table struct {
	Name    string
	Columns []Column

	// IDColumn is the optional automatically assigned integer ID column
	// It's SERIAL PRIMARY KEY in PostgreSQL, INTEGER PRIMARY KEY in SQLite, and AUTO_INCREMENT PRIMARY KEY in MySQL
	IDColumn *Column
}

type Column struct {
	Name     string
	Nullable bool
	Type     string
}
