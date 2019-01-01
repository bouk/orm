package rel

import (
	"fmt"
)

type SelectStatement struct {
	Table   string
	Columns []Expr
	Wheres  []Expr
	Orders  []Expr
	Limit   int64
	Offset  int64
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
