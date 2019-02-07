package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/pkg/errors"

	"bou.ke/ctxdb"
	"bou.ke/orm/rel"
)

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
	return Posts().Where("user_id", o.ID)
}

func (o *User) Save(ctx context.Context) error {
	if o.deleted {
		return fmt.Errorf("record deleted")
	}

	if o.persisted {
		stmt := &rel.UpdateStatement{
			Table: "users",
			Wheres: []rel.Expr{
				&rel.Equality{
					Field: "id",
					Expr:  &rel.BindParam{Value: o.old.ID},
				},
			},
		}

		if o.ID != o.old.ID {
			stmt.Values = append(stmt.Values, &rel.Assignment{
				Column: "id",
				Value: &rel.BindParam{
					Value: o.ID,
				},
			})
		}

		if o.FirstName != o.old.FirstName {
			stmt.Values = append(stmt.Values, &rel.Assignment{
				Column: "first_name",
				Value: &rel.BindParam{
					Value: o.FirstName,
				},
			})
		}

		if o.LastName != o.old.LastName {
			stmt.Values = append(stmt.Values, &rel.Assignment{
				Column: "last_name",
				Value: &rel.BindParam{
					Value: o.LastName,
				},
			})
		}

		query, values := stmt.Build()
		_, err := ctxdb.Exec(ctx, query, values...)
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
		res, err := ctxdb.Exec(ctx, query, values...)
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

	o.old.ID = o.ID
	o.old.FirstName = o.FirstName
	o.old.LastName = o.LastName

	return nil
}

func (o *User) Delete(ctx context.Context) error {
	_, err := Users().Where("id", o.ID).DeleteAll(ctx)
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
	// All ...
	All(ctx context.Context) ([]*User, error)

	// Count ...
	Count(ctx context.Context) (int64, error)

	// Create
	// TODO

	// DeleteAll ...
	DeleteAll(ctx context.Context) (int64, error)

	// Find ...
	Find(ctx context.Context, id int64) (*User, error)

	// FindBy ...
	FindBy(ctx context.Context, query string, args ...interface{}) (*User, error)

	// First ...
	First(ctx context.Context) (*User, error)

	// Last ...
	Last(ctx context.Context) (*User, error)

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
	Take(ctx context.Context) (*User, error)

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

func (q *userRelation) buildQuery(fields []string) (query string, args []interface{}) {
	columns := make([]rel.Expr, 0, len(fields))
	for _, field := range fields {
		columns = append(columns, &rel.Literal{Value: field})
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

func (q *userRelation) queryRow(ctx context.Context, fields []string, dest []interface{}) error {
	query, args := q.buildQuery(fields)

	return ctxdb.QueryRow(ctx, query, args...).Scan(dest...)
}

func (q *userRelation) query(ctx context.Context, fields []string) (*sql.Rows, error) {
	query, args := q.buildQuery(fields)

	return ctxdb.Query(ctx, query, args...)
}

func (q *userRelation) Count(ctx context.Context) (int64, error) {
	var count int64
	err := q.queryRow(ctx, []string{"COUNT(*)"}, []interface{}{&count})
	return count, err
}

func (q *userRelation) DeleteAll(ctx context.Context) (int64, error) {
	s := rel.DeleteStatement{
		Table:  "users",
		Wheres: q.whereClause,
	}

	query, args := s.Build()

	res, err := ctxdb.Exec(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (q *userRelation) Where(query string, args ...interface{}) UserRelation {
	if len(args)%2 != 1 {
		panic("invalid where call")
	}

	q.whereClause = append(q.whereClause, &rel.Equality{
		Field: query,
		Expr:  &rel.BindParam{Value: args[0]},
	})

	for i := 1; i < len(args); i += 2 {
		q.whereClause = append(q.whereClause, &rel.Equality{
			Field: args[i].(string),
			Expr:  &rel.BindParam{Value: args[i+1]},
		})
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
		if eq, ok := w.(*rel.Equality); ok {
			if bind, ok := eq.Expr.(*rel.BindParam); ok {
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

func (q *userRelation) All(ctx context.Context) ([]*User, error) {
	var users []*User

	fields := q.columnFields()
	rows, err := q.query(ctx, fields)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	row := &User{}
	row.persisted = true
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

		o.old.ID = o.ID
		o.old.FirstName = o.FirstName
		o.old.LastName = o.LastName

		users = append(users, o)
	}

	return users, rows.Err()
}

func (q *userRelation) Take(ctx context.Context) (*User, error) {
	fields := q.columnFields()

	o := &User{}
	o.persisted = true
	ptrs, err := o.pointersForFields(fields)
	if err != nil {
		return nil, err
	}

	q.limit = 1
	err = q.queryRow(ctx, fields, ptrs)
	if err == sql.ErrNoRows {
		return nil, nil
	}

	o.old.ID = o.ID
	o.old.FirstName = o.FirstName
	o.old.LastName = o.LastName

	return o, err
}

func (q *userRelation) Find(ctx context.Context, id int64) (*User, error) {
	return q.FindBy(ctx, "id", id)
}

func (q *userRelation) FindBy(ctx context.Context, query string, args ...interface{}) (*User, error) {
	return q.Where(query, args...).Take(ctx)
}

func (q *userRelation) First(ctx context.Context) (*User, error) {
	return q.Order("id", "ASC").Take(ctx)
}

func (q *userRelation) Last(ctx context.Context) (*User, error) {
	return q.Order("id", "DESC").Take(ctx)
}

func (q *userRelation) Order(query string, args ...string) UserRelation {
	if len(args) == 0 {
		args = []string{"ASC"}
	}

	if len(args)%2 != 1 {
		panic("invalid where call")
	}

	q.orderValues = append(q.orderValues, orderDirection(&rel.Literal{Value: query}, args[0]))

	for i := 1; i < len(args); i += 2 {
		q.orderValues = append(q.orderValues, orderDirection(&rel.Literal{Value: args[i]}, args[i+1]))
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

func (o *Post) User(ctx context.Context) (*User, error) {
	return Users().Find(ctx, o.UserID)
}

func (o *Post) Save(ctx context.Context) error {
	if o.deleted {
		return fmt.Errorf("record deleted")
	}

	if o.persisted {
		stmt := &rel.UpdateStatement{
			Table: "posts",
			Wheres: []rel.Expr{
				&rel.Equality{
					Field: "id",
					Expr:  &rel.BindParam{Value: o.old.ID},
				},
			},
		}

		if o.ID != o.old.ID {
			stmt.Values = append(stmt.Values, &rel.Assignment{
				Column: "id",
				Value: &rel.BindParam{
					Value: o.ID,
				},
			})
		}

		if o.UserID != o.old.UserID {
			stmt.Values = append(stmt.Values, &rel.Assignment{
				Column: "user_id",
				Value: &rel.BindParam{
					Value: o.UserID,
				},
			})
		}

		if o.Body != o.old.Body {
			stmt.Values = append(stmt.Values, &rel.Assignment{
				Column: "body",
				Value: &rel.BindParam{
					Value: o.Body,
				},
			})
		}

		query, values := stmt.Build()
		_, err := ctxdb.Exec(ctx, query, values...)
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
		res, err := ctxdb.Exec(ctx, query, values...)
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

	o.old.ID = o.ID
	o.old.UserID = o.UserID
	o.old.Body = o.Body

	return nil
}

func (o *Post) Delete(ctx context.Context) error {
	_, err := Posts().Where("id", o.ID).DeleteAll(ctx)
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
	// All ...
	All(ctx context.Context) ([]*Post, error)

	// Count ...
	Count(ctx context.Context) (int64, error)

	// Create
	// TODO

	// DeleteAll ...
	DeleteAll(ctx context.Context) (int64, error)

	// Find ...
	Find(ctx context.Context, id int64) (*Post, error)

	// FindBy ...
	FindBy(ctx context.Context, query string, args ...interface{}) (*Post, error)

	// First ...
	First(ctx context.Context) (*Post, error)

	// Last ...
	Last(ctx context.Context) (*Post, error)

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
	Take(ctx context.Context) (*Post, error)

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

func (q *postRelation) buildQuery(fields []string) (query string, args []interface{}) {
	columns := make([]rel.Expr, 0, len(fields))
	for _, field := range fields {
		columns = append(columns, &rel.Literal{Value: field})
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

func (q *postRelation) queryRow(ctx context.Context, fields []string, dest []interface{}) error {
	query, args := q.buildQuery(fields)

	return ctxdb.QueryRow(ctx, query, args...).Scan(dest...)
}

func (q *postRelation) query(ctx context.Context, fields []string) (*sql.Rows, error) {
	query, args := q.buildQuery(fields)

	return ctxdb.Query(ctx, query, args...)
}

func (q *postRelation) Count(ctx context.Context) (int64, error) {
	var count int64
	err := q.queryRow(ctx, []string{"COUNT(*)"}, []interface{}{&count})
	return count, err
}

func (q *postRelation) DeleteAll(ctx context.Context) (int64, error) {
	s := rel.DeleteStatement{
		Table:  "posts",
		Wheres: q.whereClause,
	}

	query, args := s.Build()

	res, err := ctxdb.Exec(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (q *postRelation) Where(query string, args ...interface{}) PostRelation {
	if len(args)%2 != 1 {
		panic("invalid where call")
	}

	q.whereClause = append(q.whereClause, &rel.Equality{
		Field: query,
		Expr:  &rel.BindParam{Value: args[0]},
	})

	for i := 1; i < len(args); i += 2 {
		q.whereClause = append(q.whereClause, &rel.Equality{
			Field: args[i].(string),
			Expr:  &rel.BindParam{Value: args[i+1]},
		})
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
		if eq, ok := w.(*rel.Equality); ok {
			if bind, ok := eq.Expr.(*rel.BindParam); ok {
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

func (q *postRelation) All(ctx context.Context) ([]*Post, error) {
	var posts []*Post

	fields := q.columnFields()
	rows, err := q.query(ctx, fields)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	row := &Post{}
	row.persisted = true
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

		o.old.ID = o.ID
		o.old.UserID = o.UserID
		o.old.Body = o.Body

		posts = append(posts, o)
	}

	return posts, rows.Err()
}

func (q *postRelation) Take(ctx context.Context) (*Post, error) {
	fields := q.columnFields()

	o := &Post{}
	o.persisted = true
	ptrs, err := o.pointersForFields(fields)
	if err != nil {
		return nil, err
	}

	q.limit = 1
	err = q.queryRow(ctx, fields, ptrs)
	if err == sql.ErrNoRows {
		return nil, nil
	}

	o.old.ID = o.ID
	o.old.UserID = o.UserID
	o.old.Body = o.Body

	return o, err
}

func (q *postRelation) Find(ctx context.Context, id int64) (*Post, error) {
	return q.FindBy(ctx, "id", id)
}

func (q *postRelation) FindBy(ctx context.Context, query string, args ...interface{}) (*Post, error) {
	return q.Where(query, args...).Take(ctx)
}

func (q *postRelation) First(ctx context.Context) (*Post, error) {
	return q.Order("id", "ASC").Take(ctx)
}

func (q *postRelation) Last(ctx context.Context) (*Post, error) {
	return q.Order("id", "DESC").Take(ctx)
}

func (q *postRelation) Order(query string, args ...string) PostRelation {
	if len(args) == 0 {
		args = []string{"ASC"}
	}

	if len(args)%2 != 1 {
		panic("invalid where call")
	}

	q.orderValues = append(q.orderValues, orderDirection(&rel.Literal{Value: query}, args[0]))

	for i := 1; i < len(args); i += 2 {
		q.orderValues = append(q.orderValues, orderDirection(&rel.Literal{Value: args[i]}, args[i+1]))
	}

	return q
}

func orderDirection(e rel.Expr, direction string) rel.Expr {
	switch direction {
	case "ASC", "asc":
		return &rel.Ascending{
			Expr: e,
		}
	case "DESC", "desc":
		return &rel.Descending{
			Expr: e,
		}
	}

	panic("fail")
}
