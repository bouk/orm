example/db/db.generated.go: template.tmpl
	go run main.go orm.go > example/db/db.generated.go

.PHONY: example/db/db.generated.go
