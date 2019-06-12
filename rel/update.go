package rel

type UpdateStatement struct {
	Table  string
	Values []Expr
	Wheres []Expr
}

func (s *UpdateStatement) Build() (string, []interface{}) {
	var c collector
	c.WriteString("UPDATE ")
	c.WriteString(s.Table)
	c.WriteString(" SET ")
	for i, value := range s.Values {
		if i > 0 {
			c.WriteString(", ")
		}
		value.writeTo(&c)
	}

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
