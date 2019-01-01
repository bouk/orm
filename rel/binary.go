package rel

type Assignment struct {
	Column string
	Value  Expr
}

func (a *Assignment) writeTo(c *collector) {
	c.WriteString(a.Column)
	c.WriteString(" = ")
	a.Value.writeTo(c)
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
