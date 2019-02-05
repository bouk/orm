package ctxdb

import (
	"context"
	"database/sql"
	"errors"
)

type connKey struct{}
type rollback struct{}

func (r rollback) Error() string {
	return "rolling back"
}

var (
	missingDBErr       = errors.New("missing Databaser in context")
	cantTxErr          = errors.New("can't begin transaction on Databaser in context")
	Rollback     error = rollback{}
)

// Databaser is a general interface for sql.Conn, sql.DB, and sql.Tx
type Databaser interface {
	// ExecContext executes a query without returning any rows. The args are for any placeholder parameters in the query.
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)

	// QueryContext executes a query that returns rows, typically a SELECT. The args are for any placeholder parameters in the query.
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)

	// QueryRowContext executes a query that is expected to return at most one row. QueryRowContext always returns a non-nil value. Errors are deferred until Row's Scan method is called. If the query selects no rows, the *Row's Scan will return ErrNoRows. Otherwise, the *Row's Scan scans the first selected row and discards the rest.
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

// With will return a new context that includes Databaser
func With(ctx context.Context, db Databaser) context.Context {
	return context.WithValue(ctx, connKey{}, db)
}

func getDB(ctx context.Context) Databaser {
	db, _ := ctx.Value(connKey{}).(Databaser)
	return db
}

type beginTxer interface {
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
}

// Tx creates a new transaction in the Conn or DB in the context, and executes f with this transaction. It does a rollback if f returns an error, and returns that error. It will rollback and return nil if the error is ctxdb.Rollback. If f does not return an error, it will commit.
func Tx(ctx context.Context, f func(ctx context.Context) error) error {
	db := getDB(ctx)
	if db == nil {
		return missingDBErr
	}

	txer, ok := db.(beginTxer)
	if !ok {
		return cantTxErr
	}

	tx, err := txer.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	err = f(With(ctx, tx))
	if err != nil {
		tx.Rollback()
		if err == Rollback {
			return nil
		}

		return err
	}

	return tx.Commit()
}

// Exec executes a query without returning any rows. The args are for any placeholder parameters in the query.
func Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	db := getDB(ctx)
	if db == nil {
		return nil, missingDBErr
	}
	return db.ExecContext(ctx, query, args...)
}

// Query executes a query that returns rows, typically a SELECT. The args are for any placeholder parameters in the query.
func Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	db := getDB(ctx)
	if db == nil {
		return nil, missingDBErr
	}
	return db.QueryContext(ctx, query, args...)
}

// QueryRow executes a query that is expected to return at most one row. QueryRow always returns a non-nil value. Errors are deferred until Row's Scan method is called. If the query selects no rows, the *Row's Scan will return ErrNoRows. Otherwise, the *Row's Scan scans the first selected row and discards the rest.
func QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	db := getDB(ctx)
	if db == nil {
		return (&row{err: missingDBErr}).intoDBRow()
	}
	return db.QueryRowContext(ctx, query, args...)
}
