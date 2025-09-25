package openapi3_test

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/swaggest/openapi-go"
	"github.com/swaggest/openapi-go/openapi3"
)

func handleError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func ExampleSpec_UnmarshalJSON() {
	yml := []byte(`{
  "openapi": "3.0.0",
  "info": {
    "version": "1.0.0",
    "title": "Swagger Petstore",
    "license": {
      "name": "MIT"
    }
  },
  "servers": [
    {
      "url": "http://petstore.swagger.io/v1"
    }
  ],
  "paths": {
    "/pets": {
      "get": {
        "summary": "List all pets",
        "operationId": "listPets",
        "tags": [
          "pets"
        ],
        "parameters": [
          {
            "name": "limit",
            "in": "query",
            "description": "How many items to return at one time (max 100)",
            "required": false,
            "schema": {
              "type": "integer",
              "format": "int32"
            }
          }
        ],
        "responses": {
          "200": {
            "description": "A paged array of pets",
            "headers": {
              "x-next": {
                "description": "A link to the next page of responses",
                "schema": {
                  "type": "string"
                }
              }
            },
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/Pets"
                }
              }
            }
          },
          "default": {
            "description": "unexpected error",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/Error"
                }
              }
            }
          }
        }
      },
      "post": {
        "summary": "Create a pet",
        "operationId": "createPets",
        "tags": [
          "pets"
        ],
        "responses": {
          "201": {
            "description": "Null response"
          },
          "default": {
            "description": "unexpected error",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/Error"
                }
              }
            }
          }
        }
      }
    },
    "/pets/{petId}": {
      "get": {
        "summary": "Info for a specific pet",
        "operationId": "showPetById",
        "tags": [
          "pets"
        ],
        "parameters": [
          {
            "name": "petId",
            "in": "path",
            "required": true,
            "description": "The id of the pet to retrieve",
            "schema": {
              "type": "string"
            }
          }
        ],
        "responses": {
          "200": {
            "description": "Expected response to a valid request",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/Pet"
                }
              }
            }
          },
          "default": {
            "description": "unexpected error",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/Error"
                }
              }
            }
          }
        }
      }
    }
  },
  "components": {
    "schemas": {
      "Pet": {
        "type": "object",
        "required": [
          "id",
          "name"
        ],
        "properties": {
          "id": {
            "type": "integer",
            "format": "int64"
          },
          "name": {
            "type": "string"
          },
          "tag": {
            "type": "string"
          }
        }
      },
      "Pets": {
        "type": "array",
        "items": {
          "$ref": "#/components/schemas/Pet"
        }
      },
      "Error": {
        "type": "object",
        "required": [
          "code",
          "message"
        ],
        "properties": {
          "code": {
            "x-foo": "bar",
            "type": "integer",
            "format": "int32"
          },
          "message": {
            "type": "string"
          }
        }
      }
    }
  }
}`)

	var s openapi3.Spec

	if err := s.UnmarshalYAML(yml); err != nil {
		log.Fatal(err)
	}

	fmt.Println(s.Info.Title)
	fmt.Println(s.Components.Schemas.MapOfSchemaOrRefValues["Error"].Schema.Properties["code"].Schema.MapOfAnything["x-foo"])

	// Output:
	// Swagger Petstore
	// bar
}

func ExampleSpec_UnmarshalYAML() {
	yml := []byte(`
openapi: "3.0.0"
info:
  version: 1.0.0
  title: Swagger Petstore
  license:
    name: MIT
servers:
  - url: http://petstore.swagger.io/v1
paths:
  /pets:
    get:
      summary: List all pets
      operationId: listPets
      tags:
        - pets
      parameters:
        - name: limit
          in: query
          description: How many items to return at one time (max 100)
          required: false
          schema:
            type: integer
            format: int32
      responses:
        '200':
          description: A paged array of pets
          headers:
            x-next:
              description: A link to the next page of responses
              schema:
                type: string
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/Pets"
        default:
          description: unexpected error
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/Error"
    post:
      summary: Create a pet
      operationId: createPets
      tags:
        - pets
      responses:
        '201':
          description: Null response
        default:
          description: unexpected error
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/Error"
  /pets/{petId}:
    get:
      summary: Info for a specific pet
      operationId: showPetById
      tags:
        - pets
      parameters:
        - name: petId
          in: path
          required: true
          description: The id of the pet to retrieve
          schema:
            type: string
      responses:
        '200':
          description: Expected response to a valid request
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/Pet"
        default:
          description: unexpected error
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/Error"
components:
  schemas:
    Pet:
      type: object
      required:
        - id
        - name
      properties:
        id:
          type: integer
          format: int64
        name:
          type: string
        tag:
          type: string
    Pets:
      type: array
      items:
        $ref: "#/components/schemas/Pet"
    Error:
      type: object
      required:
        - code
        - message
      properties:
        code:
          x-foo: bar
          type: integer
          format: int32
        message:
          type: string
`)

	var s openapi3.Spec

	if err := s.UnmarshalYAML(yml); err != nil {
		log.Fatal(err)
	}

	fmt.Println(s.Info.Title)
	fmt.Println(s.Components.Schemas.MapOfSchemaOrRefValues["Error"].Schema.Properties["code"].Schema.MapOfAnything["x-foo"])

	// Output:
	// Swagger Petstore
	// bar
}

func ExampleReflector_AddOperation() {
	reflector := openapi3.NewReflector()
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
		} `json:"items,omitempty"`
	}

	type resp struct {
		ID     string `json:"id" example:"XXX-XXXXX"`
		Amount uint   `json:"amount"`
		Items  []struct {
			Count uint   `json:"count"`
			Name  string `json:"name"`
		} `json:"items,omitempty"`
		UpdatedAt time.Time `json:"updated_at"`
	}

	putOp, _ := reflector.NewOperationContext(http.MethodPut, "/things/{id}")

	putOp.AddReqStructure(new(req))
	putOp.AddRespStructure(new(resp))
	putOp.AddRespStructure(new([]resp), openapi.WithHTTPStatus(http.StatusConflict))
	handleError(reflector.AddOperation(putOp))

	getOp, _ := reflector.NewOperationContext(http.MethodGet, "/things/{id}")
	getOp.AddReqStructure(new(req))
	getOp.AddRespStructure(new(resp))
	handleError(reflector.AddOperation(getOp))

	schema, err := reflector.Spec.MarshalYAML()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(schema))

	// Output:
	// openapi: 3.0.3
	// info:
	//   description: Put something here
	//   title: Things API
	//   version: 1.2.3
	// paths:
	//   /things/{id}:
	//     get:
	//       parameters:
	//       - in: query
	//         name: locale
	//         schema:
	//           pattern: ^[a-z]{2}-[A-Z]{2}$
	//           type: string
	//       - in: path
	//         name: id
	//         required: true
	//         schema:
	//           example: XXX-XXXXX
	//           type: string
	//       responses:
	//         "200":
	//           content:
	//             application/json:
	//               schema:
	//                 $ref: '#/components/schemas/Openapi3TestResp'
	//           description: OK
	//     put:
	//       parameters:
	//       - in: query
	//         name: locale
	//         schema:
	//           pattern: ^[a-z]{2}-[A-Z]{2}$
	//           type: string
	//       - in: path
	//         name: id
	//         required: true
	//         schema:
	//           example: XXX-XXXXX
	//           type: string
	//       requestBody:
	//         content:
	//           application/json:
	//             schema:
	//               $ref: '#/components/schemas/Openapi3TestReq'
	//       responses:
	//         "200":
	//           content:
	//             application/json:
	//               schema:
	//                 $ref: '#/components/schemas/Openapi3TestResp'
	//           description: OK
	//         "409":
	//           content:
	//             application/json:
	//               schema:
	//                 items:
	//                   $ref: '#/components/schemas/Openapi3TestResp'
	//                 type: array
	//           description: Conflict
	// components:
	//   schemas:
	//     Openapi3TestReq:
	//       properties:
	//         amount:
	//           minimum: 0
	//           type: integer
	//         items:
	//           items:
	//             properties:
	//               count:
	//                 minimum: 0
	//                 type: integer
	//               name:
	//                 type: string
	//             type: object
	//           type: array
	//         string:
	//           type: string
	//       type: object
	//     Openapi3TestResp:
	//       properties:
	//         amount:
	//           minimum: 0
	//           type: integer
	//         id:
	//           example: XXX-XXXXX
	//           type: string
	//         items:
	//           items:
	//             properties:
	//               count:
	//                 minimum: 0
	//                 type: integer
	//               name:
	//                 type: string
	//             type: object
	//           type: array
	//         updated_at:
	//           format: date-time
	//           type: string
	//       type: object
}

func ExampleReflector_AddOperation_queryObject() {
	reflector := openapi3.NewReflector()
	reflector.Spec.Info.
		WithTitle("Things API").
		WithVersion("1.2.3").
		WithDescription("Put something here")

	type jsonFilter struct {
		Foo    string `json:"foo"`
		Bar    int    `json:"bar"`
		Deeper struct {
			Val string `json:"val"`
		} `json:"deeper"`
	}

	type deepObjectFilter struct {
		Baz    bool    `query:"baz"`
		Quux   float64 `query:"quux"`
		Deeper struct {
			Val string `query:"val"`
		} `query:"deeper"`
	}

	type req struct {
		ID     string `path:"id" example:"XXX-XXXXX"`
		Locale string `query:"locale" pattern:"^[a-z]{2}-[A-Z]{2}$"`
		// Object values can be serialized in JSON (with json field tags in the value struct).
		JSONFilter jsonFilter `query:"json_filter"`
		// Or as deepObject (with same field tag as parent, .e.g query).
		DeepObjectFilter deepObjectFilter `query:"deep_object_filter"`
	}

	getOp, _ := reflector.NewOperationContext(http.MethodGet, "/things/{id}")

	getOp.AddReqStructure(new(req))
	_ = reflector.AddOperation(getOp)

	schema, err := reflector.Spec.MarshalYAML()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(schema))

	// Output:
	// openapi: 3.0.3
	// info:
	//   description: Put something here
	//   title: Things API
	//   version: 1.2.3
	// paths:
	//   /things/{id}:
	//     get:
	//       parameters:
	//       - in: query
	//         name: locale
	//         schema:
	//           pattern: ^[a-z]{2}-[A-Z]{2}$
	//           type: string
	//       - content:
	//           application/json:
	//             schema:
	//               $ref: '#/components/schemas/Openapi3TestJsonFilter'
	//         in: query
	//         name: json_filter
	//       - explode: true
	//         in: query
	//         name: deep_object_filter
	//         schema:
	//           $ref: '#/components/schemas/Openapi3TestDeepObjectFilter'
	//         style: deepObject
	//       - in: path
	//         name: id
	//         required: true
	//         schema:
	//           example: XXX-XXXXX
	//           type: string
	//       responses:
	//         "204":
	//           description: No Content
	// components:
	//   schemas:
	//     Openapi3TestDeepObjectFilter:
	//       properties:
	//         baz:
	//           type: boolean
	//         deeper:
	//           properties:
	//             val:
	//               type: string
	//           type: object
	//         quux:
	//           format: double
	//           type: number
	//       type: object
	//     Openapi3TestJsonFilter:
	//       properties:
	//         bar:
	//           type: integer
	//         deeper:
	//           properties:
	//             val:
	//               type: string
	//           type: object
	//         foo:
	//           type: string
	//       type: object
}
