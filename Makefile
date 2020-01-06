gen:
	cd resources/schema/ && json-cli gen-go openapi3.json --output ../../openapi3/entities.go --package-name openapi3 --with-zero-values --fluent-setters --enable-default-additional-properties --root-name Schema
	gofmt -w ./openapi3/entities.go
