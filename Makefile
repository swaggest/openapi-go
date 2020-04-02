gen:
	cd resources/schema/ && json-cli gen-go openapi3.json --output ../../openapi3/entities.go --package-name openapi3 --with-zero-values --fluent-setters --root-name Spec
	gofmt -w ./openapi3/entities.go

lint:
	golangci-lint run --enable-all --disable gochecknoglobals,funlen,gomnd,gocognit ./...
