package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/pkg/errors"

	"bou.ke/orm/rel"
)

// DB is a general interface for sql.Conn, sql.DB, and sql.Tx
type DB interface {
	// ExecContext executes a query without returning any rows. The args are for any placeholder parameters in the query.
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)

	// QueryContext executes a query that returns rows, typically a SELECT. The args are for any placeholder parameters in the query.
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)

	// QueryRowContext executes a query that is expected to return at most one row. QueryRowContext always returns a non-nil value. Errors are deferred until Row's Scan method is called. If the query selects no rows, the *Row's Scan will return ErrNoRows. Otherwise, the *Row's Scan scans the first selected row and discards the rest.
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

var ErrNotFound error = errors.New("not found")

type Relation interface {
	// Count ...
	Count(ctx context.Context, db DB) (int64, error)

	// DeleteAll ...
	DeleteAll(ctx context.Context, db DB) (int64, error)

	// UpdateAll
	UpdateAll(ctx context.Context, db DB, query string, args ...interface{}) error
}

type UserFields struct {
	// ID ...
	ID int64

	// FirstName ...
	FirstName string

	// LastName ...
	LastName string
}

type User struct {
	UserFields

	// If true, then this record exists in the DB
	persisted bool
	deleted   bool

	old UserFields
}

func (o *User) Posts() PostRelation {
	return Posts().Where("user_id = ?", o.ID)
}

func (o *User) Save(ctx context.Context, db DB) error {
	if o.deleted {
		return fmt.Errorf("record deleted")
	}

	if o.persisted {
		stmt := &rel.UpdateStatement{
			Table: "users",
			Wheres: []rel.Expr{
				rel.Assignment{
					Field: "id",
					Value: rel.BindParam{Value: o.old.ID},
				},
			},
		}

		if o.ID != o.old.ID {
			stmt.Values = append(stmt.Values, rel.Assignment{
				Field: "id",
				Value: &rel.BindParam{
					Value: o.ID,
				},
			})
		}

		if o.FirstName != o.old.FirstName {
			stmt.Values = append(stmt.Values, rel.Assignment{
				Field: "first_name",
				Value: &rel.BindParam{
					Value: o.FirstName,
				},
			})
		}

		if o.LastName != o.old.LastName {
			stmt.Values = append(stmt.Values, rel.Assignment{
				Field: "last_name",
				Value: &rel.BindParam{
					Value: o.LastName,
				},
			})
		}

		query, values := stmt.Build()
		_, err := db.ExecContext(ctx, query, values...)
		if err != nil {
			return errors.Wrapf(err, "executing %q", query)
		}
	} else {
		stmt := &rel.InsertStatement{
			Table: "users",
		}

		if o.ID != 0 {
			stmt.Columns = append(stmt.Columns, "id")
			stmt.Values = append(stmt.Values, &rel.BindParam{
				Value: o.ID,
			})
		}
		stmt.Columns = append(stmt.Columns, "first_name")
		stmt.Values = append(stmt.Values, &rel.BindParam{
			Value: o.FirstName,
		})
		stmt.Columns = append(stmt.Columns, "last_name")
		stmt.Values = append(stmt.Values, &rel.BindParam{
			Value: o.LastName,
		})

		query, values := stmt.Build()
		res, err := db.ExecContext(ctx, query, values...)
		if err != nil {
			return errors.Wrapf(err, "executing %q", query)
		}
		o.persisted = true

		if o.ID == 0 {
			o.ID, err = res.LastInsertId()
			if err != nil {
				return err
			}
		}
	}

	o.old = o.UserFields

	return nil
}

func (o *User) Delete(ctx context.Context, db DB) error {
	_, err := Users().Where("id = ?", o.ID).DeleteAll(ctx, db)
	if err != nil {
		return err
	}
	o.deleted = true
	return err
}

func (o *User) fieldPointerForColumn(column string) interface{} {
	switch column {
	case "id":
		return &o.ID
	case "first_name":
		return &o.FirstName
	case "last_name":
		return &o.LastName
	default:
		return nil
	}
}

func (o *User) pointersForFields(fields []string) ([]interface{}, error) {
	pointers := make([]interface{}, len(fields))
	for i, field := range fields {
		ptr := o.fieldPointerForColumn(field)
		if ptr == nil {
			return nil, fmt.Errorf("unknown column %q", field)
		}
		pointers[i] = ptr
	}
	return pointers, nil
}

// assignField sets the field to the value.
// It panics if the field doesn't exist or the value is the wrong type.
func (o *User) assignField(name string, value interface{}) {
	switch name {
	case "id":
		o.ID = value.(int64)
	case "first_name":
		o.FirstName = value.(string)
	case "last_name":
		o.LastName = value.(string)
	default:
		panic("unknown field: " + name)
	}
}

type UserRelation interface {
	Relation

	// All ...
	All(ctx context.Context, db DB) ([]*User, error)

	// TODO: Create(ctx context.Context, db DB, query string, args ...interface{}) (*User, error)

	// Find ...
	Find(ctx context.Context, db DB, id int64) (*User, error)

	// FindBy ...
	FindBy(ctx context.Context, db DB, query string, args ...interface{}) (*User, error)

	// First ...
	First(ctx context.Context, db DB) (*User, error)

	// Last ...
	Last(ctx context.Context, db DB) (*User, error)

	// Limit ...
	Limit(limit int64) UserRelation

	// New creates a User populated with the scope of the relation
	New() *User

	// Offset ...
	Offset(offset int64) UserRelation

	// Order ...
	Order(query string, args ...string) UserRelation

	// Select ...
	Select(fields ...string) UserRelation

	// Take ...
	Take(ctx context.Context, db DB) (*User, error)

	// Where ...
	Where(query string, args ...interface{}) UserRelation
}

// Users returns a UserRelation, allowing you to build a query.
// Note: the intermediate result of calls to the Relation can not be reused.
func Users() UserRelation {
	return &userRelation{}
}

type userRelation struct {
	fields      []string
	whereClause []rel.Expr
	orderValues []rel.Expr
	limit       int64
	offset      int64
}

func userFindBySQL(ctx context.Context, db DB, query string, args ...interface{}) ([]*User, error) {
	var users []*User
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	row := &User{}
	row.persisted = true
	fields, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	ptrs, err := row.pointersForFields(fields)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		if err := rows.Scan(ptrs...); err != nil {
			return nil, err
		}

		o := &User{}
		*o = *row

		o.old = o.UserFields

		users = append(users, o)
	}

	return users, rows.Err()
}

func (q *userRelation) UpdateAll(ctx context.Context, db DB, query string, args ...interface{}) error {
	clauses, err := rel.ParseAssignment(query, args...)
	if err != nil {
		return err
	}

	stmt := &rel.UpdateStatement{
		Table:  "users",
		Wheres: q.whereClause,
		Values: clauses,
	}

	query, values := stmt.Build()
	_, err = db.ExecContext(ctx, query, values...)
	return err
}

func (q *userRelation) ToSQL() (query string, args []interface{}) {
	fields := q.columnFields()
	columns := make([]rel.Expr, 0, len(fields))
	for _, field := range fields {
		columns = append(columns, &rel.Literal{Text: field})
	}
	s := rel.SelectStatement{
		Columns: columns,
		Table:   "users",
		Wheres:  q.whereClause,
		Orders:  q.orderValues,
		Limit:   q.limit,
		Offset:  q.offset,
	}
	return s.Build()
}

func (q *userRelation) Count(ctx context.Context, db DB) (int64, error) {
	var count int64
	q.fields = []string{"COUNT(*)"}

	query, args := q.ToSQL()
	err := db.QueryRowContext(ctx, query, args...).Scan(&count)

	return count, err
}

func (q *userRelation) DeleteAll(ctx context.Context, db DB) (int64, error) {
	s := rel.DeleteStatement{
		Table:  "users",
		Wheres: q.whereClause,
	}

	query, args := s.Build()

	res, err := db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (q *userRelation) Where(query string, args ...interface{}) UserRelation {
	clauses, err := rel.ParseAssignment(query, args...)

	// TODO(bouk): return error relation
	if err != nil {
		panic(err)
	}

	for _, clause := range clauses {
		q.whereClause = append(q.whereClause, clause)
	}

	return q
}

func (q *userRelation) Limit(limit int64) UserRelation {
	q.limit = limit
	return q
}

func (q *userRelation) New() *User {
	o := &User{}
	for _, w := range q.whereClause {
		if eq, ok := w.(rel.Assignment); ok {
			if bind, ok := eq.Value.(rel.BindParam); ok {
				o.assignField(eq.Field, bind.Value)
			}
		}
	}

	return o
}

func (q *userRelation) Select(fields ...string) UserRelation {
	q.fields = append(q.fields, fields...)
	return q
}

func (q *userRelation) Offset(offset int64) UserRelation {
	q.offset = offset
	return q
}

func (q *userRelation) columnFields() []string {
	if q.fields == nil {
		return []string{
			"id",
			"first_name",
			"last_name",
		}
	} else {
		return q.fields
	}
}

func (q *userRelation) All(ctx context.Context, db DB) ([]*User, error) {
	query, args := q.ToSQL()
	return userFindBySQL(ctx, db, query, args...)
}

func (q *userRelation) Take(ctx context.Context, db DB) (*User, error) {
	q.limit = 1
	os, err := q.All(ctx, db)
	if err != nil {
		return nil, err
	}

	if len(os) == 0 {
		return nil, ErrNotFound
	}

	return os[0], nil
}

func (q *userRelation) Find(ctx context.Context, db DB, id int64) (*User, error) {
	return q.FindBy(ctx, db, "id = ?", id)
}

func (q *userRelation) FindBy(ctx context.Context, db DB, query string, args ...interface{}) (*User, error) {
	return q.Where(query, args...).Take(ctx, db)
}

func (q *userRelation) First(ctx context.Context, db DB) (*User, error) {
	return q.Order("id ASC").Take(ctx, db)
}

func (q *userRelation) Last(ctx context.Context, db DB) (*User, error) {
	return q.Order("id DESC").Take(ctx, db)
}

func (q *userRelation) Order(query string, args ...string) UserRelation {
	q.orderValues = append(q.orderValues, &rel.Literal{Text: query})

	for i := 0; i < len(args); i++ {
		q.orderValues = append(q.orderValues, &rel.Literal{Text: args[i]})
	}

	return q
}

type PostFields struct {
	// ID ...
	ID int64

	// UserID ...
	UserID int64

	// Body ...
	Body string
}

type Post struct {
	PostFields

	// If true, then this record exists in the DB
	persisted bool
	deleted   bool

	old PostFields
}

func (o *Post) User(ctx context.Context, db DB) (*User, error) {
	return Users().Find(ctx, db, o.UserID)
}

func (o *Post) Save(ctx context.Context, db DB) error {
	if o.deleted {
		return fmt.Errorf("record deleted")
	}

	if o.persisted {
		stmt := &rel.UpdateStatement{
			Table: "posts",
			Wheres: []rel.Expr{
				rel.Assignment{
					Field: "id",
					Value: rel.BindParam{Value: o.old.ID},
				},
			},
		}

		if o.ID != o.old.ID {
			stmt.Values = append(stmt.Values, rel.Assignment{
				Field: "id",
				Value: &rel.BindParam{
					Value: o.ID,
				},
			})
		}

		if o.UserID != o.old.UserID {
			stmt.Values = append(stmt.Values, rel.Assignment{
				Field: "user_id",
				Value: &rel.BindParam{
					Value: o.UserID,
				},
			})
		}

		if o.Body != o.old.Body {
			stmt.Values = append(stmt.Values, rel.Assignment{
				Field: "body",
				Value: &rel.BindParam{
					Value: o.Body,
				},
			})
		}

		query, values := stmt.Build()
		_, err := db.ExecContext(ctx, query, values...)
		if err != nil {
			return errors.Wrapf(err, "executing %q", query)
		}
	} else {
		stmt := &rel.InsertStatement{
			Table: "posts",
		}

		if o.ID != 0 {
			stmt.Columns = append(stmt.Columns, "id")
			stmt.Values = append(stmt.Values, &rel.BindParam{
				Value: o.ID,
			})
		}
		stmt.Columns = append(stmt.Columns, "user_id")
		stmt.Values = append(stmt.Values, &rel.BindParam{
			Value: o.UserID,
		})
		stmt.Columns = append(stmt.Columns, "body")
		stmt.Values = append(stmt.Values, &rel.BindParam{
			Value: o.Body,
		})

		query, values := stmt.Build()
		res, err := db.ExecContext(ctx, query, values...)
		if err != nil {
			return errors.Wrapf(err, "executing %q", query)
		}
		o.persisted = true

		if o.ID == 0 {
			o.ID, err = res.LastInsertId()
			if err != nil {
				return err
			}
		}
	}

	o.old = o.PostFields

	return nil
}

func (o *Post) Delete(ctx context.Context, db DB) error {
	_, err := Posts().Where("id = ?", o.ID).DeleteAll(ctx, db)
	if err != nil {
		return err
	}
	o.deleted = true
	return err
}

func (o *Post) fieldPointerForColumn(column string) interface{} {
	switch column {
	case "id":
		return &o.ID
	case "user_id":
		return &o.UserID
	case "body":
		return &o.Body
	default:
		return nil
	}
}

func (o *Post) pointersForFields(fields []string) ([]interface{}, error) {
	pointers := make([]interface{}, len(fields))
	for i, field := range fields {
		ptr := o.fieldPointerForColumn(field)
		if ptr == nil {
			return nil, fmt.Errorf("unknown column %q", field)
		}
		pointers[i] = ptr
	}
	return pointers, nil
}

// assignField sets the field to the value.
// It panics if the field doesn't exist or the value is the wrong type.
func (o *Post) assignField(name string, value interface{}) {
	switch name {
	case "id":
		o.ID = value.(int64)
	case "user_id":
		o.UserID = value.(int64)
	case "body":
		o.Body = value.(string)
	default:
		panic("unknown field: " + name)
	}
}

type PostRelation interface {
	Relation

	// All ...
	All(ctx context.Context, db DB) ([]*Post, error)

	// TODO: Create(ctx context.Context, db DB, query string, args ...interface{}) (*Post, error)

	// Find ...
	Find(ctx context.Context, db DB, id int64) (*Post, error)

	// FindBy ...
	FindBy(ctx context.Context, db DB, query string, args ...interface{}) (*Post, error)

	// First ...
	First(ctx context.Context, db DB) (*Post, error)

	// Last ...
	Last(ctx context.Context, db DB) (*Post, error)

	// Limit ...
	Limit(limit int64) PostRelation

	// New creates a Post populated with the scope of the relation
	New() *Post

	// Offset ...
	Offset(offset int64) PostRelation

	// Order ...
	Order(query string, args ...string) PostRelation

	// Select ...
	Select(fields ...string) PostRelation

	// Take ...
	Take(ctx context.Context, db DB) (*Post, error)

	// Where ...
	Where(query string, args ...interface{}) PostRelation
}

// Posts returns a PostRelation, allowing you to build a query.
// Note: the intermediate result of calls to the Relation can not be reused.
func Posts() PostRelation {
	return &postRelation{}
}

type postRelation struct {
	fields      []string
	whereClause []rel.Expr
	orderValues []rel.Expr
	limit       int64
	offset      int64
}

func postFindBySQL(ctx context.Context, db DB, query string, args ...interface{}) ([]*Post, error) {
	var posts []*Post
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	row := &Post{}
	row.persisted = true
	fields, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	ptrs, err := row.pointersForFields(fields)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		if err := rows.Scan(ptrs...); err != nil {
			return nil, err
		}

		o := &Post{}
		*o = *row

		o.old = o.PostFields

		posts = append(posts, o)
	}

	return posts, rows.Err()
}

func (q *postRelation) UpdateAll(ctx context.Context, db DB, query string, args ...interface{}) error {
	clauses, err := rel.ParseAssignment(query, args...)
	if err != nil {
		return err
	}

	stmt := &rel.UpdateStatement{
		Table:  "posts",
		Wheres: q.whereClause,
		Values: clauses,
	}

	query, values := stmt.Build()
	_, err = db.ExecContext(ctx, query, values...)
	return err
}

func (q *postRelation) ToSQL() (query string, args []interface{}) {
	fields := q.columnFields()
	columns := make([]rel.Expr, 0, len(fields))
	for _, field := range fields {
		columns = append(columns, &rel.Literal{Text: field})
	}
	s := rel.SelectStatement{
		Columns: columns,
		Table:   "posts",
		Wheres:  q.whereClause,
		Orders:  q.orderValues,
		Limit:   q.limit,
		Offset:  q.offset,
	}
	return s.Build()
}

func (q *postRelation) Count(ctx context.Context, db DB) (int64, error) {
	var count int64
	q.fields = []string{"COUNT(*)"}

	query, args := q.ToSQL()
	err := db.QueryRowContext(ctx, query, args...).Scan(&count)

	return count, err
}

func (q *postRelation) DeleteAll(ctx context.Context, db DB) (int64, error) {
	s := rel.DeleteStatement{
		Table:  "posts",
		Wheres: q.whereClause,
	}

	query, args := s.Build()

	res, err := db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (q *postRelation) Where(query string, args ...interface{}) PostRelation {
	clauses, err := rel.ParseAssignment(query, args...)

	// TODO(bouk): return error relation
	if err != nil {
		panic(err)
	}

	for _, clause := range clauses {
		q.whereClause = append(q.whereClause, clause)
	}

	return q
}

func (q *postRelation) Limit(limit int64) PostRelation {
	q.limit = limit
	return q
}

func (q *postRelation) New() *Post {
	o := &Post{}
	for _, w := range q.whereClause {
		if eq, ok := w.(rel.Assignment); ok {
			if bind, ok := eq.Value.(rel.BindParam); ok {
				o.assignField(eq.Field, bind.Value)
			}
		}
	}

	return o
}

func (q *postRelation) Select(fields ...string) PostRelation {
	q.fields = append(q.fields, fields...)
	return q
}

func (q *postRelation) Offset(offset int64) PostRelation {
	q.offset = offset
	return q
}

func (q *postRelation) columnFields() []string {
	if q.fields == nil {
		return []string{
			"id",
			"user_id",
			"body",
		}
	} else {
		return q.fields
	}
}

func (q *postRelation) All(ctx context.Context, db DB) ([]*Post, error) {
	query, args := q.ToSQL()
	return postFindBySQL(ctx, db, query, args...)
}

func (q *postRelation) Take(ctx context.Context, db DB) (*Post, error) {
	q.limit = 1
	os, err := q.All(ctx, db)
	if err != nil {
		return nil, err
	}

	if len(os) == 0 {
		return nil, ErrNotFound
	}

	return os[0], nil
}

func (q *postRelation) Find(ctx context.Context, db DB, id int64) (*Post, error) {
	return q.FindBy(ctx, db, "id = ?", id)
}

func (q *postRelation) FindBy(ctx context.Context, db DB, query string, args ...interface{}) (*Post, error) {
	return q.Where(query, args...).Take(ctx, db)
}

func (q *postRelation) First(ctx context.Context, db DB) (*Post, error) {
	return q.Order("id ASC").Take(ctx, db)
}

func (q *postRelation) Last(ctx context.Context, db DB) (*Post, error) {
	return q.Order("id DESC").Take(ctx, db)
}

func (q *postRelation) Order(query string, args ...string) PostRelation {
	q.orderValues = append(q.orderValues, &rel.Literal{Text: query})

	for i := 0; i < len(args); i++ {
		q.orderValues = append(q.orderValues, &rel.Literal{Text: args[i]})
	}

	return q
}
