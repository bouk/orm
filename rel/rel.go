package rel

import (
	"bytes"
)

type collector struct {
	bytes.Buffer
	values []interface{}
}

type Expr interface {
	writeTo(c *collector)
}

type BindParam struct {
	Value interface{}
}

func (b *BindParam) writeTo(c *collector) {
	c.WriteString("?")
	c.values = append(c.values, b.Value)
}

type Literal struct {
	Text   string
	Params []interface{}
}

func (l *Literal) writeTo(c *collector) {
	c.WriteString(l.Text)
	c.values = append(c.values, l.Params...)
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
