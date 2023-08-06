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
    * `json` for request bodies and responses in JSON
    * `query`, `path` for parameters in URL
    * `header`, `cookie`, `formData`, `file` for other parameters
    * `form` acts as `query` and `formData`
    * [field tags](https://github.com/swaggest/jsonschema-go#field-tags) named after JSON Schema/OpenAPI 3 Schema constraints
    * `collectionFormat` to unpack slices from string
        * `csv` comma-separated values,
        * `ssv` space-separated values,
        * `pipes` pipe-separated values (`|`),
        * `multi` ampersand-separated values (`&`), 
* Flexible schema control with [`jsonschema-go`](https://github.com/swaggest/jsonschema-go#implementing-interfaces-on-a-type)

## Example

[Other examples](https://pkg.go.dev/github.com/swaggest/openapi-go/openapi3#pkg-examples).

```go
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

putOp := openapi3.Operation{}

handleError(reflector.SetRequest(&putOp, new(req), http.MethodPut))
handleError(reflector.SetJSONResponse(&putOp, new(resp), http.StatusOK))
handleError(reflector.SetJSONResponse(&putOp, new([]resp), http.StatusConflict))
handleError(reflector.Spec.AddOperation(http.MethodPut, "/things/{id}", putOp))

getOp := openapi3.Operation{}

handleError(reflector.SetRequest(&getOp, new(req), http.MethodGet))
handleError(reflector.SetJSONResponse(&getOp, new(resp), http.StatusOK))
handleError(reflector.Spec.AddOperation(http.MethodGet, "/things/{id}", getOp))

schema, err := reflector.Spec.MarshalYAML()
if err != nil {
    log.Fatal(err)
}

fmt.Println(string(schema))
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
                $ref: '#/components/schemas/Openapi3TestResp'
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
              $ref: '#/components/schemas/Openapi3TestReq'
      responses:
        "200":
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Openapi3TestResp'
          description: OK
        "409":
          content:
            application/json:
              schema:
                items:
                  $ref: '#/components/schemas/Openapi3TestResp'
                type: array
          description: Conflict
components:
  schemas:
    Openapi3TestReq:
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
          type: array
        string:
          type: string
      type: object
    Openapi3TestResp:
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
          type: array
        updated_at:
          format: date-time
          type: string
      type: object
```
