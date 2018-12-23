package db

import (
	"context"
	"database/sql"
	"fmt"

	"bou.ke/orm/rel"
)

type User struct {
	// ID ...
	ID uint64

	// FirstName ...
	FirstName string

	// LastName ...
	LastName string
}

func (o *User) Posts() PostRelation {
	return Posts().Where("user_id", o.ID)
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

type UserRelation interface {
	All(ctx context.Context) ([]*User, error)
	Count(ctx context.Context) (uint64, error)
	Find(ctx context.Context, id uint64) (*User, error)
	FindBy(ctx context.Context, query string, args ...interface{}) (*User, error)
	First(ctx context.Context) (*User, error)
	Last(ctx context.Context) (*User, error)
	Limit(limit uint64) UserRelation
	Offset(offset uint64) UserRelation
	Order(query string, args ...string) UserRelation
	Select(fields ...string) UserRelation
	Take(ctx context.Context) (*User, error)
	Where(query string, args ...interface{}) UserRelation

	queryRow(ctx context.Context, fields []string, dest []interface{}) error
	query(ctx context.Context, fields []string) (*sql.Rows, error)
}

func Users() UserRelation {
	return &userRelation{}
}

type userRelation struct {
	fields      []string
	whereClause []rel.Expr
	orderValues []rel.Expr
	limit       uint64
	offset      uint64
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
	db := getDB(ctx)
	query, args := q.buildQuery(fields)

	return db.QueryRowContext(ctx, query, args).Scan(dest...)
}

func (q *userRelation) query(ctx context.Context, fields []string) (*sql.Rows, error) {
	db := getDB(ctx)
	query, args := q.buildQuery(fields)

	return db.QueryContext(ctx, query, args)
}

func (q *userRelation) Count(ctx context.Context) (uint64, error) {
	var count uint64
	err := q.queryRow(ctx, []string{"COUNT(*)"}, []interface{}{&count})
	return count, err
}

func (q *userRelation) Where(query string, args ...interface{}) UserRelation {
	if len(args)%2 != 1 {
		panic("invalid where call")
	}

	q.whereClause = append(q.whereClause, &rel.Equality{
		Field: query,
		Expr:  &rel.BindParam{Value: args[0]},
	})

	for i := 1; i <= len(args); i += 2 {
		q.whereClause = append(q.whereClause, &rel.Equality{
			Field: args[i].(string),
			Expr:  &rel.BindParam{Value: args[i+1]},
		})
	}

	return q
}

func (q *userRelation) Limit(limit uint64) UserRelation {
	q.limit = limit
	return q
}

func (q *userRelation) Select(fields ...string) UserRelation {
	q.fields = append(q.fields, fields...)
	return q
}

func (q *userRelation) Offset(offset uint64) UserRelation {
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
	ptrs, err := row.pointersForFields(fields)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		if err := rows.Scan(ptrs...); err != nil {
			return nil, err
		}

		user := &User{}
		*user = *row
		users = append(users, user)
	}

	return users, rows.Err()
}

func (q *userRelation) Take(ctx context.Context) (*User, error) {
	fields := q.columnFields()

	user := &User{}
	ptrs, err := user.pointersForFields(fields)
	if err != nil {
		return nil, err
	}

	err = q.Limit(1).queryRow(ctx, fields, ptrs)

	return user, err
}

func (q *userRelation) Find(ctx context.Context, id uint64) (*User, error) {
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

	q.orderValues = append(q.orderValues, orderDirection(&rel.Literal{Value: query}, args[0]))

	for i := 1; i <= len(args); i += 2 {
		q.orderValues = append(q.orderValues, orderDirection(&rel.Literal{Value: args[i]}, args[i+1]))
	}

	return q
}

type Post struct {
	// ID ...
	ID uint64

	// UserID ...
	UserID uint64

	// Body ...
	Body string
}

func (o *Post) User(ctx context.Context) (*User, error) {
	return Users().Find(ctx, o.UserID)
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

type PostRelation interface {
	All(ctx context.Context) ([]*Post, error)
	Count(ctx context.Context) (uint64, error)
	Find(ctx context.Context, id uint64) (*Post, error)
	FindBy(ctx context.Context, query string, args ...interface{}) (*Post, error)
	First(ctx context.Context) (*Post, error)
	Last(ctx context.Context) (*Post, error)
	Limit(limit uint64) PostRelation
	Offset(offset uint64) PostRelation
	Order(query string, args ...string) PostRelation
	Select(fields ...string) PostRelation
	Take(ctx context.Context) (*Post, error)
	Where(query string, args ...interface{}) PostRelation

	queryRow(ctx context.Context, fields []string, dest []interface{}) error
	query(ctx context.Context, fields []string) (*sql.Rows, error)
}

func Posts() PostRelation {
	return &postRelation{}
}

type postRelation struct {
	fields      []string
	whereClause []rel.Expr
	orderValues []rel.Expr
	limit       uint64
	offset      uint64
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
	db := getDB(ctx)
	query, args := q.buildQuery(fields)

	return db.QueryRowContext(ctx, query, args).Scan(dest...)
}

func (q *postRelation) query(ctx context.Context, fields []string) (*sql.Rows, error) {
	db := getDB(ctx)
	query, args := q.buildQuery(fields)

	return db.QueryContext(ctx, query, args)
}

func (q *postRelation) Count(ctx context.Context) (uint64, error) {
	var count uint64
	err := q.queryRow(ctx, []string{"COUNT(*)"}, []interface{}{&count})
	return count, err
}

func (q *postRelation) Where(query string, args ...interface{}) PostRelation {
	if len(args)%2 != 1 {
		panic("invalid where call")
	}

	q.whereClause = append(q.whereClause, &rel.Equality{
		Field: query,
		Expr:  &rel.BindParam{Value: args[0]},
	})

	for i := 1; i <= len(args); i += 2 {
		q.whereClause = append(q.whereClause, &rel.Equality{
			Field: args[i].(string),
			Expr:  &rel.BindParam{Value: args[i+1]},
		})
	}

	return q
}

func (q *postRelation) Limit(limit uint64) PostRelation {
	q.limit = limit
	return q
}

func (q *postRelation) Select(fields ...string) PostRelation {
	q.fields = append(q.fields, fields...)
	return q
}

func (q *postRelation) Offset(offset uint64) PostRelation {
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
	ptrs, err := row.pointersForFields(fields)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		if err := rows.Scan(ptrs...); err != nil {
			return nil, err
		}

		post := &Post{}
		*post = *row
		posts = append(posts, post)
	}

	return posts, rows.Err()
}

func (q *postRelation) Take(ctx context.Context) (*Post, error) {
	fields := q.columnFields()

	post := &Post{}
	ptrs, err := post.pointersForFields(fields)
	if err != nil {
		return nil, err
	}

	err = q.Limit(1).queryRow(ctx, fields, ptrs)

	return post, err
}

func (q *postRelation) Find(ctx context.Context, id uint64) (*Post, error) {
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

	q.orderValues = append(q.orderValues, orderDirection(&rel.Literal{Value: query}, args[0]))

	for i := 1; i <= len(args); i += 2 {
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
