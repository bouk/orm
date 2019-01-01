package rel

type DeleteStatement struct {
	Table  string
	Wheres []Expr
}

func (s *DeleteStatement) Build() (string, []interface{}) {
	var c collector
	c.WriteString("DELETE FROM ")
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

	return c.String(), c.values
}
