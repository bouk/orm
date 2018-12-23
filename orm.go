package main // import "bou.ke/orm"

import (
	"github.com/gobuffalo/flect"
)

type Input struct {
	Tables  Tables
	Package string
}

type TableName string

func (t TableName) Singular() string {
	return flect.Singularize(string(t))
}

func (t TableName) StructName() string {
	return flect.New(string(t)).Singularize().Pascalize().String()
}

func (t TableName) RelationName() string {
	return flect.Pascalize(string(t))
}

type Tables []Table
type Table struct {
	Name      string      `json:"name"`
	Columns   []Column    `json:"columns"`
	BelongsTo []TableName `json:"belongs_to"`
	HasMany   []TableName `json:"has_many"`
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
