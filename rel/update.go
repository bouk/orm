package rel

type InsertStatement struct {
	Table   string
	Columns []string
	Values  []Expr
}

func (s *InsertStatement) Build() (string, []interface{}) {
	var c collector
	c.WriteString("INSERT INTO ")
	c.WriteString(s.Table)
	c.WriteString(" ")
	if len(s.Columns) > 0 {
		c.WriteString("(")
		for i, value := range s.Columns {
			if i > 0 {
				c.WriteString(", ")
			}
			c.WriteString(value)
		}
		c.WriteString(")")
	}
	c.WriteString(" VALUES ")
	for i, value := range s.Values {
		if i > 0 {
			c.WriteString(", ")
		}
		value.writeTo(&c)
	}

	return c.String(), c.values
}
