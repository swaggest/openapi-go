GOLANGCI_LINT_VERSION := "v1.27.0"

gen:
	cd resources/schema/ && json-cli gen-go openapi3.json --output ../../openapi3/entities.go --package-name openapi3 --with-zero-values --fluent-setters --root-name Spec
	gofmt -w ./openapi3/entities.go

lint:
	@test -s $(GOPATH)/bin/golangci-lint-$(GOLANGCI_LINT_VERSION) || (curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b /tmp $(GOLANGCI_LINT_VERSION) && mv /tmp/golangci-lint $(GOPATH)/bin/golangci-lint-$(GOLANGCI_LINT_VERSION))
	@$(GOPATH)/bin/golangci-lint-$(GOLANGCI_LINT_VERSION) run ./...
