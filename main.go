package main

import (
	"bytes"
	"go/format"
	"os"
	"text/template"
)

func main() {
	tpl := template.Must(template.ParseFiles("template.tmpl"))
	tables := Tables{
		{
			Name: "users",
			Columns: Columns{
				{
					Name: "id",
					Type: "uint64",
				},
				{
					Name: "first_name",
					Type: "string",
				},
				{
					Name: "last_name",
					Type: "string",
				},
			},
			HasMany: []TableName{"posts"},
		},
		{
			Name: "posts",
			Columns: Columns{
				{
					Name: "id",
					Type: "uint64",
				},
				{
					Name: "user_id",
					Type: "uint64",
				},
				{
					Name: "body",
					Type: "string",
				},
			},
			BelongsTo: []TableName{"users"},
		},
	}

	var b bytes.Buffer
	err := tpl.Execute(&b, Input{
		Tables:  tables,
		Package: "db",
	})
	if err != nil {
		panic(err)
	}
	output, err := format.Source(b.Bytes())
	if err != nil {
		b.WriteTo(os.Stdout)
		panic(err)
	}
	os.Stdout.Write(output)
}
