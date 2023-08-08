package openapi31_test

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/swaggest/openapi-go"
	"github.com/swaggest/openapi-go/openapi31"
)

func handleError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func ExampleReflector_AddOperation() {
	reflector := openapi31.NewReflector()
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
	// openapi: 3.1.0
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
	//           examples:
	//           - XXX-XXXXX
	//           type: string
	//       responses:
	//         "200":
	//           content:
	//             application/json:
	//               schema:
	//                 $ref: '#/components/schemas/Openapi31TestResp'
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
	//           examples:
	//           - XXX-XXXXX
	//           type: string
	//       requestBody:
	//         content:
	//           application/json:
	//             schema:
	//               $ref: '#/components/schemas/Openapi31TestReq'
	//       responses:
	//         "200":
	//           content:
	//             application/json:
	//               schema:
	//                 $ref: '#/components/schemas/Openapi31TestResp'
	//           description: OK
	//         "409":
	//           content:
	//             application/json:
	//               schema:
	//                 items:
	//                   $ref: '#/components/schemas/Openapi31TestResp'
	//                 type:
	//                 - "null"
	//                 - array
	//           description: Conflict
	// components:
	//   schemas:
	//     Openapi31TestReq:
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
	//     Openapi31TestResp:
	//       properties:
	//         amount:
	//           minimum: 0
	//           type: integer
	//         id:
	//           examples:
	//           - XXX-XXXXX
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
	reflector := openapi31.NewReflector()
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
	// openapi: 3.1.0
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
	//               $ref: '#/components/schemas/Openapi31TestJsonFilter'
	//         in: query
	//         name: json_filter
	//       - explode: true
	//         in: query
	//         name: deep_object_filter
	//         schema:
	//           $ref: '#/components/schemas/Openapi31TestDeepObjectFilter'
	//         style: deepObject
	//       - in: path
	//         name: id
	//         required: true
	//         schema:
	//           examples:
	//           - XXX-XXXXX
	//           type: string
	//       responses:
	//         "204":
	//           description: No Content
	// components:
	//   schemas:
	//     Openapi31TestDeepObjectFilter:
	//       properties:
	//         baz:
	//           type: boolean
	//         deeper:
	//           properties:
	//             val:
	//               type: string
	//           type: object
	//         quux:
	//           type: number
	//       type: object
	//     Openapi31TestJsonFilter:
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
