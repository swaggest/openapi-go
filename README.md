# OpenAPI structures for Go

<img align="right" width="250px" src="/resources/logo.png">

This library provides Go structures to marshal/unmarshal and reflect [OpenAPI Schema](https://swagger.io/resources/open-api/) documents.

For automated HTTP REST service framework built with this library please check [`github.com/swaggest/rest`](https://github.com/swaggest/rest).

[![Build Status](https://github.com/swaggest/openapi-go/workflows/test/badge.svg)](https://github.com/swaggest/openapi-go/actions?query=branch%3Amaster+workflow%3Atest)
[![Coverage Status](https://codecov.io/gh/swaggest/openapi-go/branch/master/graph/badge.svg)](https://codecov.io/gh/swaggest/openapi-go)
[![GoDevDoc](https://img.shields.io/badge/dev-doc-00ADD8?logo=go)](https://pkg.go.dev/github.com/swaggest/openapi-go)
[![time tracker](https://wakatime.com/badge/github/swaggest/openapi-go.svg)](https://wakatime.com/badge/github/swaggest/openapi-go)
![Code lines](https://sloc.xyz/github/swaggest/openapi-go/?category=code)
![Comments](https://sloc.xyz/github/swaggest/openapi-go/?category=comments)

## Features

* Type safe mapping of OpenAPI 3 documents with Go structures generated from schema.
* Type-based reflection of Go structures to OpenAPI 3.0 or 3.1 schema.
* Schema control with field tags
    * `json` for request bodies and responses in JSON [[example](#example)]
    * `query`, `path` for parameters in URL [[example](#example)]
    * `header`, `cookie`, `formData`, `file` for other parameters [[example](#example-formdata-upload)]
    * `form` acts as `query` and `formData`
    * Customizable request and response content via `ContentUnit` [[example](#example-binary-upload)]
    * [field tags](https://github.com/swaggest/jsonschema-go#field-tags) named after JSON Schema/OpenAPI 3 Schema constraints
    * `collectionFormat` to unpack slices from string
        * `csv` comma-separated values,
        * `ssv` space-separated values,
        * `pipes` pipe-separated values (`|`),
        * `multi` ampersand-separated values (`&`),
        * `json` additionally to slices unpacks maps and structs,
* Flexible schema control with [`jsonschema-go`](https://github.com/swaggest/jsonschema-go#implementing-interfaces-on-a-type) [[example](#example-flexible-schema-control)]

## Example

[Other examples](https://pkg.go.dev/github.com/swaggest/openapi-go/openapi3#pkg-examples).

```go
package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/swaggest/openapi-go"
	"github.com/swaggest/openapi-go/openapi3"
)

func main() {
	reflector := openapi3.Reflector{}
	reflector.Spec = &openapi3.Spec{Openapi: "3.0.3"}
	reflector.Spec.Info.
		WithTitle("Things API").
		WithVersion("1.2.3").
		WithDescription("Put something here")

	type req struct {
		ID     string `path:"id" example:"XXX-XXXXX"`
		Locale string `query:"locale" pattern:"^[a-z]{2}-[A-Z]{2}$"`
		Title  string `json:"string"`
		Amount uint   `json:"amount"`
		Items  []struct {
			Count uint   `json:"count"`
			Name  string `json:"name"`
		} `json:"items"`
	}

	type resp struct {
		ID     string `json:"id" example:"XXX-XXXXX"`
		Amount uint   `json:"amount"`
		Items  []struct {
			Count uint   `json:"count"`
			Name  string `json:"name"`
		} `json:"items"`
		UpdatedAt time.Time `json:"updated_at"`
	}

	putOp, err := reflector.NewOperationContext(http.MethodPut, "/things/{id}")
	if err != nil {
		log.Fatal(err)
	}

	putOp.AddReqStructure(new(req))
	putOp.AddRespStructure(new(resp), func(cu *openapi.ContentUnit) { cu.HTTPStatus = http.StatusOK })
	putOp.AddRespStructure(new([]resp), func(cu *openapi.ContentUnit) { cu.HTTPStatus = http.StatusConflict })

	reflector.AddOperation(putOp)

	getOp, err := reflector.NewOperationContext(http.MethodGet, "/things/{id}")
	if err != nil {
		log.Fatal(err)
	}

	getOp.AddReqStructure(new(req))
	getOp.AddRespStructure(new(resp), func(cu *openapi.ContentUnit) { cu.HTTPStatus = http.StatusOK })

	err = reflector.AddOperation(getOp)
	if err != nil {
		log.Fatal(err)
	}

	schema, err := reflector.Spec.MarshalYAML()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(schema))
}
```

Output:

```yaml
openapi: 3.0.3
info:
  description: Put something here
  title: Things API
  version: 1.2.3
paths:
  /things/{id}:
    get:
      parameters:
      - in: query
        name: locale
        schema:
          pattern: ^[a-z]{2}-[A-Z]{2}$
          type: string
      - in: path
        name: id
        required: true
        schema:
          example: XXX-XXXXX
          type: string
      responses:
        "200":
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Resp'
          description: OK
    put:
      parameters:
      - in: query
        name: locale
        schema:
          pattern: ^[a-z]{2}-[A-Z]{2}$
          type: string
      - in: path
        name: id
        required: true
        schema:
          example: XXX-XXXXX
          type: string
      requestBody:
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/Req'
      responses:
        "200":
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Resp'
          description: OK
        "409":
          content:
            application/json:
              schema:
                items:
                  $ref: '#/components/schemas/Resp'
                type: array
          description: Conflict
components:
  schemas:
    Req:
      properties:
        amount:
          minimum: 0
          type: integer
        items:
          items:
            properties:
              count:
                minimum: 0
                type: integer
              name:
                type: string
            type: object
          nullable: true
          type: array
        string:
          type: string
      type: object
    Resp:
      properties:
        amount:
          minimum: 0
          type: integer
        id:
          example: XXX-XXXXX
          type: string
        items:
          items:
            properties:
              count:
                minimum: 0
                type: integer
              name:
                type: string
            type: object
          nullable: true
          type: array
        updated_at:
          format: date-time
          type: string
      type: object
```

## Example FormData Upload

```go
package main

import (
	"fmt"
	"log"
	"mime/multipart"
	"net/http"

	"github.com/swaggest/openapi-go/openapi31"
)

func main() {
	reflector := openapi31.NewReflector()
	reflector.Spec = &openapi31.Spec{Openapi: "3.1.0"}
	reflector.Spec.Info.
		WithTitle("Things API").
		WithVersion("1.2.3").
		WithDescription("Put something here")

	type req struct {
		SeqNum        string         `formData:"seqNum"`
		TotalChunks   string         `formData:"totalChunks"`
		UploadId      string         `formData:"uploadId"`
		ChunkedUpload bool           `formData:"chunkedUpload"`
		Message       string         `formData:"message" required:"true"`
		File          multipart.File `formData:"file" required:"true"`
	}

	postOp, err := reflector.NewOperationContext(http.MethodPost, "/upload")
	if err != nil {
		log.Fatal(err)
	}

	postOp.AddReqStructure(new(req))

	err = reflector.AddOperation(postOp)
	if err != nil {
		log.Fatal(err)
	}

	schema, err := reflector.Spec.MarshalYAML()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(schema))
}
```

## Example Binary Upload

```go
package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/swaggest/openapi-go"
	"github.com/swaggest/openapi-go/openapi31"
)

func main() {
	reflector := openapi31.NewReflector()
	reflector.Spec = &openapi31.Spec{Openapi: "3.1.0"}
	reflector.Spec.Info.
		WithTitle("Things API").
		WithVersion("1.2.3").
		WithDescription("Put something here")

	putOp, err := reflector.NewOperationContext(http.MethodPut, "/upload")
	if err != nil {
		log.Fatal(err)
	}

	putOp.AddReqStructure(nil, func(cu *openapi.ContentUnit) {
		cu.ContentType = "application/octet-stream"
		cu.Format = "binary"
		cu.Description = "Binary data bytes"
	})

	putOp.AddRespStructure(nil, func(cu *openapi.ContentUnit) { cu.HTTPStatus = 2 }) // Any 2XX status code

	err = reflector.AddOperation(putOp)
	if err != nil {
		log.Fatal(err)
	}

	schema, err := reflector.Spec.MarshalYAML()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(schema))
}
```

## Example Flexible Schema Control

```go
package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/swaggest/jsonschema-go"
	"github.com/swaggest/openapi-go/openapi31"
)

func main() {
	reflector := openapi31.NewReflector()
	reflector.Spec = &openapi31.Spec{Openapi: "3.1.0"}
	reflector.Spec.Info.
		WithTitle("Things API").
		WithVersion("1.2.3").
		WithDescription("Put something here")

	// Use existing `validate:"required"` struct tags to mark fields as required instead of using `required` tag.
	reflector.DefaultOptions = append(reflector.DefaultOptions, jsonschema.InterceptProp(func(params jsonschema.InterceptPropParams) error {
		if !params.Processed {
			return nil
		}

		if v, ok := params.Field.Tag.Lookup("validate"); ok {
			if strings.Contains(v, "required") {
				params.ParentSchema.Required = append(params.ParentSchema.Required, params.Name)
			}
		}

		return nil
	}))

	type req struct {
		ID string `json:"id" validate:"required"` // Even without explicit `required` tag this field will be marked as required.
	}

	postOp, err := reflector.NewOperationContext(http.MethodPost, "/things")
	if err != nil {
		log.Fatal(err)
	}
	postOp.AddReqStructure(new(req))

	err = reflector.AddOperation(postOp)
	if err != nil {
		log.Fatal(err)
	}

	schema, err := reflector.Spec.MarshalYAML()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(schema))
}
```