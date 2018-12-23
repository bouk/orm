package main

import (
	"github.com/gobuffalo/flect"
)

type Input struct {
	Tables  Tables
	Package string
}

type Tables []Table
type Table struct {
	Name    string   `json:"name"`
	Columns []Column `json:"columns"`
}

func (t *Table) Singular() string {
	return flect.Singularize(t.Name)
}

func (t *Table) StructName() string {
	return flect.New(t.Name).Singularize().Pascalize().String()
}

func (t *Table) RelationName() string {
	return flect.Pascalize(t.Name)
}

type Columns []Column
type Column struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

func (c *Column) FieldName() string {
	return flect.Pascalize(c.Name)
}
