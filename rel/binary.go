package rel

type Assignment struct {
	Field string
	Value Expr
}

func (a Assignment) writeTo(c *collector) {
	c.WriteString(a.Field)
	c.WriteString(" = ")
	a.Value.writeTo(c)
}
