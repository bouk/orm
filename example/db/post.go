package db

import (
	"context"
	"database/sql"

	"bou.ke/orm/rel"
)

type Post struct {
	ID     uint64
	UserID uint64
	Body   string

	orm_original *Post
}

func (p *Post) User(ctx context.Context) (*User, error) {
	return Users.Find(ctx, p.UserID)
}

/*
func (u *Post) Update(ctx context.Context, query string, value interface{}, args ...interface{}) error
func (u *Post) Delete(ctx context.Context) error
func (u *Post) Save(ctx context.Context) error
*/

type PostRelation interface {
	All(ctx context.Context) ([]*Post, error)
	Count(ctx context.Context) (uint64, error)
	Find(ctx context.Context, id uint64) (*Post, error)
	FindBy(ctx context.Context, query string, args ...interface{}) (*Post, error)
	First(ctx context.Context) (*Post, error)
	Last(ctx context.Context) (*Post, error)
	Limit(limit uint64) PostRelation
	Offset(offset uint) PostRelation
	Order(query string, args ...string) PostRelation
	Take(ctx context.Context) (*Post, error)
	Where(query string, args ...interface{}) PostRelation

	queryRow(ctx context.Context, fields string, dest ...interface{}) error
	query(ctx context.Context, fields string) (*sql.Rows, error)

	/*
		Create(ctx context.Context, args ...interface{}) (*Post, error)
		Each(ctx context.Context, f func(idx uint, post *Post) error) error
		Exists(ctx context.Context) (bool, error)
		FindOrCreateBy(ctx context.Context, field string, value interface{}, args ...interface{}) (*Post, error)

		DeleteAll(ctx context.Context) error
		UpdateAll(ctx context.Context, query string, value interface{}, args ...interface{}) error
	*/
}

type copyingPostRelation struct {
	*postRelation
}

func (c copyingPostRelation) Limit(limit uint64) PostRelation {
	rel := *c.postRelation
	return rel.Limit(limit)
}

func (c copyingPostRelation) Offset(offset uint64) PostRelation {
	rel := *c.postRelation
	return rel.Offset(offset)
}

func (c copyingPostRelation) Order(query string, args ...string) PostRelation {
	rel := *c.postRelation
	return rel.Order(query, args...)
}

func (c copyingPostRelation) Where(query string, args ...interface{}) PostRelation {
	rel := *c.postRelation
	return rel.Where(query, args...)
}

var Posts = copyingPostRelation{&postRelation{}}

type postRelation struct {
	whereClause []rel.Expr
	orderValues []rel.Expr
	limit       uint64
	offset      uint64
}

func (q *postRelation) buildQuery(fields string) (query string, args []interface{}) {
	s := rel.SelectStatement{
		Columns: []rel.Expr{&rel.Literal{Value: fields}},
		Table:   "posts",
		Wheres:  q.whereClause,
		Orders:  q.orderValues,
		Limit:   q.limit,
		Offset:  q.offset,
	}
	return s.Build()
}

func (q *postRelation) queryRow(ctx context.Context, fields string, dest ...interface{}) error {
	db := getDB(ctx)
	query, args := q.buildQuery(fields)

	return db.QueryRowContext(ctx, query, args).Scan(dest...)
}

func (q *postRelation) query(ctx context.Context, fields string) (*sql.Rows, error) {
	db := getDB(ctx)
	query, args := q.buildQuery(fields)

	return db.QueryContext(ctx, query, args)
}

func (q *postRelation) Count(ctx context.Context) (uint64, error) {
	var count uint64
	err := q.queryRow(ctx, "COUNT(*)", &count)
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

func (q *postRelation) Offset(offset uint64) PostRelation {
	q.offset = offset
	return q
}

func (q *postRelation) All(ctx context.Context) ([]*Post, error) {
	rows, err := q.query(ctx, `id, body`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*Post
	for rows.Next() {
		post := &Post{}

		if err := rows.Scan(
			&post.ID,
			&post.Body,
		); err != nil {
			return nil, err
		}

		posts = append(posts, post)
	}

	return posts, rows.Err()
}

func (q *postRelation) Take(ctx context.Context) (*Post, error) {
	var post Post
	err := q.Limit(1).queryRow(
		ctx,
		`id, body`,
		&post.ID,
		&post.Body,
	)

	return &post, err
}

func (q *postRelation) Find(ctx context.Context, id uint64) (*Post, error) {
	return q.FindBy(ctx, "id", id)
}

func (q *postRelation) FindBy(ctx context.Context, query string, args ...interface{}) (*Post, error) {
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

func (q *postRelation) First(ctx context.Context) (*Post, error) {
	return q.Order("id", "ASC").Take(ctx)
}

func (q *postRelation) Last(ctx context.Context) (*Post, error) {
	return q.Order("id", "DESC").Take(ctx)
}
