package rel

import (
	"bytes"
	"fmt"
)

type collector struct {
	bytes.Buffer
	values []interface{}
}

type SelectStatement struct {
	Table   string
	Columns []Expr
	Wheres  []Expr
	Orders  []Expr
	Limit   uint64
	Offset  uint64
}

func (s *SelectStatement) Build() (string, []interface{}) {
	var c collector
	c.WriteString("SELECT ")

	for i, col := range s.Columns {
		if i > 0 {
			c.WriteString(", ")
		}
		col.writeTo(&c)
	}

	c.WriteString(" FROM ")
	c.WriteString(s.Table)

	if len(s.Wheres) > 0 {
		c.WriteString(" WHERE ")
		for i, where := range s.Wheres {
			if i > 0 {
				c.WriteString(" AND ")
			}
			where.writeTo(&c)
		}
	}

	if len(s.Orders) > 0 {
		c.WriteString(" ORDER BY ")
		for i, order := range s.Orders {
			if i > 0 {
				c.WriteString(", ")
			}
			order.writeTo(&c)
		}
	}

	if s.Limit != 0 {
		fmt.Fprintf(&c, " LIMIT %d", s.Limit)
	}
	if s.Offset != 0 {
		fmt.Fprintf(&c, " OFFSET %d", s.Offset)
	}

	return c.String(), c.values
}

type Expr interface {
	writeTo(c *collector)
}

type Equality struct {
	Field string
	Expr  Expr
}

func (e *Equality) writeTo(c *collector) {
	c.WriteString(e.Field)
	c.WriteString(" = ")
	e.Expr.writeTo(c)
}

type BindParam struct {
	Value interface{}
}

func (b *BindParam) writeTo(c *collector) {
	c.WriteString("?")
	c.values = append(c.values, b.Value)
}

type Literal struct {
	Value string
}

func (l *Literal) writeTo(c *collector) {
	c.WriteString(l.Value)
}

type Ascending struct {
	Expr Expr
}

func (a *Ascending) writeTo(c *collector) {
	a.Expr.writeTo(c)
	c.WriteString(" ASC")
}

type Descending struct {
	Expr Expr
}

func (a *Descending) writeTo(c *collector) {
	a.Expr.writeTo(c)
	c.WriteString(" DESC")
}
