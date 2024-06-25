#GOLANGCI_LINT_VERSION := "v1.59.1" # Optional configuration to pinpoint golangci-lint version.

# The head of Makefile determines location of dev-go to include standard targets.
GO ?= go
export GO111MODULE = on

ifneq "$(GOFLAGS)" ""
  $(info GOFLAGS: ${GOFLAGS})
endif

ifneq "$(wildcard ./vendor )" ""
  $(info Using vendor)
  modVendor =  -mod=vendor
  ifeq (,$(findstring -mod,$(GOFLAGS)))
      export GOFLAGS := ${GOFLAGS} ${modVendor}
  endif
  ifneq "$(wildcard ./vendor/github.com/bool64/dev)" ""
  	DEVGO_PATH := ./vendor/github.com/bool64/dev
  endif
endif

ifeq ($(DEVGO_PATH),)
	DEVGO_PATH := $(shell GO111MODULE=on $(GO) list ${modVendor} -f '{{.Dir}}' -m github.com/bool64/dev)
	ifeq ($(DEVGO_PATH),)
    	$(info Module github.com/bool64/dev not found, downloading.)
    	DEVGO_PATH := $(shell export GO111MODULE=on && $(GO) get github.com/bool64/dev && $(GO) list -f '{{.Dir}}' -m github.com/bool64/dev)
	endif
endif

JSON_CLI_VERSION := "v1.8.6"
JSON_CLI_VERSION_31 := "v1.11.1"

-include $(DEVGO_PATH)/makefiles/main.mk
-include $(DEVGO_PATH)/makefiles/lint.mk
-include $(DEVGO_PATH)/makefiles/test-unit.mk
-include $(DEVGO_PATH)/makefiles/reset-ci.mk

# Add your custom targets here.

## Run tests
test: test-unit

## Generate entities from schema
gen-3.0:
	@test -s $(GOPATH)/bin/json-cli-$(JSON_CLI_VERSION) || (curl -sSfL https://github.com/swaggest/json-cli/releases/download/$(JSON_CLI_VERSION)/json-cli -o $(GOPATH)/bin/json-cli-$(JSON_CLI_VERSION) && chmod +x $(GOPATH)/bin/json-cli-$(JSON_CLI_VERSION))
	@cd resources/schema/ && $(GOPATH)/bin/json-cli-$(JSON_CLI_VERSION) gen-go openapi3.json --output ../../openapi3/entities.go --package-name openapi3 --with-tests --with-zero-values --validate-required --fluent-setters --root-name Spec
	@gofmt -w ./openapi3/entities.go ./openapi3/entities_test.go


## Generate entities from schema
gen-3.1:
	@test -s $(GOPATH)/bin/json-cli-$(JSON_CLI_VERSION_31) || (curl -sSfL https://github.com/swaggest/json-cli/releases/download/$(JSON_CLI_VERSION_31)/json-cli -o $(GOPATH)/bin/json-cli-$(JSON_CLI_VERSION_31) && chmod +x $(GOPATH)/bin/json-cli-$(JSON_CLI_VERSION_31))
	@cd resources/schema/ && $(GOPATH)/bin/json-cli-$(JSON_CLI_VERSION_31)  gen-go openapi31-patched.json --config openapi31-config.json --output ../../openapi31/entities.go --package-name openapi31 --def-ptr '#/$$defs' --with-zero-values --validate-required --fluent-setters --root-name Spec
	@gofmt -w ./openapi31/entities.go
