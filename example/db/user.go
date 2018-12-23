package db

import (
	"context"
	"database/sql"

	"bou.ke/orm/rel"
)

type User struct {
	ID        uint64
	FirstName string
	LastName  string

	orm_original *User
}

func (u *User) Posts() PostRelation {
	return Posts.Where("user_id", u.ID)
}

/*
func (u *User) Update(ctx context.Context, query string, value interface{}, args ...interface{}) error
func (u *User) Delete(ctx context.Context) error
func (u *User) Save(ctx context.Context) error
*/

type UserRelation interface {
	All(ctx context.Context) ([]*User, error)
	Count(ctx context.Context) (uint64, error)
	Find(ctx context.Context, id uint64) (*User, error)
	FindBy(ctx context.Context, query string, args ...interface{}) (*User, error)
	First(ctx context.Context) (*User, error)
	Last(ctx context.Context) (*User, error)
	Limit(limit uint64) UserRelation
	Offset(offset uint) UserRelation
	Order(query string, args ...string) UserRelation
	Take(ctx context.Context) (*User, error)
	Where(query string, args ...interface{}) UserRelation

	queryRow(ctx context.Context, fields string, dest ...interface{}) error
	query(ctx context.Context, fields string) (*sql.Rows, error)

	/*
		Create(ctx context.Context, args ...interface{}) (*User, error)
		Each(ctx context.Context, f func(idx uint, user *User) error) error
		Exists(ctx context.Context) (bool, error)
		FindOrCreateBy(ctx context.Context, field string, value interface{}, args ...interface{}) (*User, error)

		DeleteAll(ctx context.Context) error
		UpdateAll(ctx context.Context, query string, value interface{}, args ...interface{}) error
	*/
}

type copyingUserRelation struct {
	*userRelation
}

func (c copyingUserRelation) Limit(limit uint64) UserRelation {
	rel := *c.userRelation
	return rel.Limit(limit)
}

func (c copyingUserRelation) Offset(offset uint64) UserRelation {
	rel := *c.userRelation
	return rel.Offset(offset)
}

func (c copyingUserRelation) Order(query string, args ...string) UserRelation {
	rel := *c.userRelation
	return rel.Order(query, args...)
}

func (c copyingUserRelation) Where(query string, args ...interface{}) UserRelation {
	rel := *c.userRelation
	return rel.Where(query, args...)
}

var Users = copyingUserRelation{&userRelation{}}

type userRelation struct {
	whereClause []rel.Expr
	orderValues []rel.Expr
	limit       uint64
	offset      uint64
}

func (q *userRelation) buildQuery(fields string) (query string, args []interface{}) {
	s := rel.SelectStatement{
		Columns: []rel.Expr{&rel.Literal{Value: fields}},
		Table:   "users",
		Wheres:  q.whereClause,
		Orders:  q.orderValues,
		Limit:   q.limit,
		Offset:  q.offset,
	}
	return s.Build()
}

func (q *userRelation) queryRow(ctx context.Context, fields string, dest ...interface{}) error {
	db := getDB(ctx)
	query, args := q.buildQuery(fields)

	return db.QueryRowContext(ctx, query, args).Scan(dest...)
}

func (q *userRelation) query(ctx context.Context, fields string) (*sql.Rows, error) {
	db := getDB(ctx)
	query, args := q.buildQuery(fields)

	return db.QueryContext(ctx, query, args)
}

func (q *userRelation) Count(ctx context.Context) (uint64, error) {
	var count uint64
	err := q.queryRow(ctx, "COUNT(*)", &count)
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

func (q *userRelation) Offset(offset uint64) UserRelation {
	q.offset = offset
	return q
}

func (q *userRelation) All(ctx context.Context) ([]*User, error) {
	rows, err := q.query(ctx, `id, first_name, last_name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		user := &User{}

		if err := rows.Scan(
			&user.ID,
			&user.FirstName,
			&user.LastName,
		); err != nil {
			return nil, err
		}

		users = append(users, user)
	}

	return users, rows.Err()
}

func (q *userRelation) Take(ctx context.Context) (*User, error) {
	var user User
	err := q.Limit(1).queryRow(
		ctx,
		`id, first_name, last_name`,
		&user.ID,
		&user.FirstName,
		&user.LastName,
	)

	return &user, err
}

func (q *userRelation) Find(ctx context.Context, id uint64) (*User, error) {
	return q.FindBy(ctx, "id", id)
}

func (q *userRelation) FindBy(ctx context.Context, query string, args ...interface{}) (*User, error) {
	return q.Where(query, args...).Take(ctx)
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

// TODO(bouk): validate order args
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

func (q *userRelation) First(ctx context.Context) (*User, error) {
	return q.Order("id", "ASC").Take(ctx)
}

func (q *userRelation) Last(ctx context.Context) (*User, error) {
	return q.Order("id", "DESC").Take(ctx)
}
