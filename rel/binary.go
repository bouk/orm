package rel

type Assignment struct {
	Field Field
	Value Expr
}

func (a Assignment) writeTo(c *collector) {
	a.Field.writeTo(c)
	c.WriteString(" = ")
	a.Value.writeTo(c)
}

type Field struct {
	Name string
}

func (f Field) writeTo(c *collector) {
	c.WriteString(f.Name)
}

// In is an SQL IN expression
type In struct {
	Left  Expr
	Right ExprList
}

func (i In) writeTo(c *collector) {
	if len(i.Right) == 0 {
		c.WriteString("1=0")
		return
	}
	i.Left.writeTo(c)
	c.WriteString(" IN (")
	i.Right.writeTo(c)
	c.WriteString(")")
}
