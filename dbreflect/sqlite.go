package dbreflect

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"bou.ke/orm/types"
)

type querier interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
}

type sqliteReflecter struct{}

func (s sqliteReflecter) typeFromSqliteType(typ string) types.Type {
	typ = strings.ToLower(typ)
	switch typ {
	case "timestamp", "date", "datetime", "time":
		return types.Time
	case "bool", "boolean", "tinyint(1)":
		return types.Bool
	case "int8":
		return types.Int8
	case "int16", "int2":
		return types.Int16
	case "int32", "int4":
		return types.Int32
	case strings.Contains(typ, "int"):
		return types.Int64
	case "text", strings.HasPrefix(typ, "varchar"):
		return types.String
	default:
		return nil
	}
}

func (s sqliteReflecter) DescribeTable(ctx context.Context, q querier, tableName string) (*Table, error) {
	rows, err := q.QueryContext(ctx, fmt.Sprintf("PRAGMA table_info(%s)", tableName))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type sqliteColumnInfo struct {
		CID          int64
		Name         string
		Type         string
		NotNull      bool
		DefaultValue *string
		PrimaryKey   int64
	}
	var columns []sqliteColumnInfo
	for rows.Next() {
		var column sqliteColumnInfo
		if err = rows.Scan(
			&column.CID,
			&column.Name,
			&column.Type,
			&column.NotNull,
			&column.DefaultValue,
			&column.PrimaryKey); err != nil {
			return nil, err
		}
		columns = append(columns, column)
	}

	table := &Table{
		Name: tableName,
	}

	return nil, nil
}
