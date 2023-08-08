package openapi31_test

import (
	"fmt"
	"log"
	"net/http"

	"github.com/swaggest/openapi-go"
	"github.com/swaggest/openapi-go/openapi31"
)

func ExampleSpec_SetHTTPBasicSecurity() {
	reflector := openapi31.Reflector{}
	securityName := "admin"

	// Declare security scheme.
	reflector.SpecEns().SetHTTPBasicSecurity(securityName, "Admin Access")

	oc, _ := reflector.NewOperationContext(http.MethodGet, "/secure")
	oc.AddRespStructure(struct {
		Secret string `json:"secret"`
	}{})

	// Add security requirement to operation.
	oc.AddSecurity(securityName)

	// Describe unauthorized response.
	oc.AddRespStructure(struct {
		Error string `json:"error"`
	}{}, func(cu *openapi.ContentUnit) {
		cu.HTTPStatus = http.StatusUnauthorized
	})

	// Add operation to schema.
	_ = reflector.AddOperation(oc)

	schema, err := reflector.Spec.MarshalYAML()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(schema))

	// Output:
	// openapi: 3.1.0
	// info:
	//   title: ""
	//   version: ""
	// paths:
	//   /secure:
	//     get:
	//       responses:
	//         "200":
	//           content:
	//             application/json:
	//               schema:
	//                 properties:
	//                   secret:
	//                     type: string
	//                 type: object
	//           description: OK
	//         "401":
	//           content:
	//             application/json:
	//               schema:
	//                 properties:
	//                   error:
	//                     type: string
	//                 type: object
	//           description: Unauthorized
	//       security:
	//       - admin: []
	// components:
	//   securitySchemes:
	//     admin:
	//       description: Admin Access
	//       scheme: basic
	//       type: http
}

func ExampleSpec_SetAPIKeySecurity() {
	reflector := openapi31.Reflector{}
	securityName := "api_key"

	// Declare security scheme.
	reflector.SpecEns().SetAPIKeySecurity(securityName, "Authorization",
		openapi.InHeader, "API Access")

	oc, _ := reflector.NewOperationContext(http.MethodGet, "/secure")
	oc.AddRespStructure(struct {
		Secret string `json:"secret"`
	}{})

	// Add security requirement to operation.
	oc.AddSecurity(securityName)

	// Describe unauthorized response.
	oc.AddRespStructure(struct {
		Error string `json:"error"`
	}{}, func(cu *openapi.ContentUnit) {
		cu.HTTPStatus = http.StatusUnauthorized
	})

	// Add operation to schema.
	_ = reflector.AddOperation(oc)

	schema, err := reflector.Spec.MarshalYAML()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(schema))

	// Output:
	// openapi: 3.1.0
	// info:
	//   title: ""
	//   version: ""
	// paths:
	//   /secure:
	//     get:
	//       responses:
	//         "200":
	//           content:
	//             application/json:
	//               schema:
	//                 properties:
	//                   secret:
	//                     type: string
	//                 type: object
	//           description: OK
	//         "401":
	//           content:
	//             application/json:
	//               schema:
	//                 properties:
	//                   error:
	//                     type: string
	//                 type: object
	//           description: Unauthorized
	//       security:
	//       - api_key: []
	// components:
	//   securitySchemes:
	//     api_key:
	//       description: API Access
	//       in: header
	//       name: Authorization
	//       type: apiKey
}

func ExampleSpec_SetHTTPBearerTokenSecurity() {
	reflector := openapi31.Reflector{}
	securityName := "bearer_token"

	// Declare security scheme.
	reflector.SpecEns().SetHTTPBearerTokenSecurity(securityName, "JWT", "Admin Access")

	oc, _ := reflector.NewOperationContext(http.MethodGet, "/secure")
	oc.AddRespStructure(struct {
		Secret string `json:"secret"`
	}{})

	// Add security requirement to operation.
	oc.AddSecurity(securityName)

	// Describe unauthorized response.
	oc.AddRespStructure(struct {
		Error string `json:"error"`
	}{}, func(cu *openapi.ContentUnit) {
		cu.HTTPStatus = http.StatusUnauthorized
	})

	// Add operation to schema.
	_ = reflector.AddOperation(oc)

	schema, err := reflector.Spec.MarshalYAML()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(schema))

	// Output:
	// openapi: 3.1.0
	// info:
	//   title: ""
	//   version: ""
	// paths:
	//   /secure:
	//     get:
	//       responses:
	//         "200":
	//           content:
	//             application/json:
	//               schema:
	//                 properties:
	//                   secret:
	//                     type: string
	//                 type: object
	//           description: OK
	//         "401":
	//           content:
	//             application/json:
	//               schema:
	//                 properties:
	//                   error:
	//                     type: string
	//                 type: object
	//           description: Unauthorized
	//       security:
	//       - bearer_token: []
	// components:
	//   securitySchemes:
	//     bearer_token:
	//       bearerFormat: JWT
	//       description: Admin Access
	//       scheme: bearer
	//       type: http
}
