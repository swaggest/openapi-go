package openapi3_test

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/swaggest/openapi-go/openapi3"
)

func handleError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func ExampleReflector_SetJSONResponse() {
	reflector := openapi3.Reflector{}
	reflector.Spec = &openapi3.Spec{Openapi: "3.0.2"}
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

	// Output:
	// openapi: 3.0.2
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
