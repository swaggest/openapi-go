package openapi3_test

import (
	"fmt"
	"log"
	"net/http"

	"github.com/swaggest/openapi-go/openapi3"
)

func ExampleReflector_SetJSONResponse_http_basic_auth() {
	reflector := openapi3.Reflector{}
	securityName := "admin"

	// Declare security scheme.
	reflector.SpecEns().ComponentsEns().SecuritySchemesEns().WithMapOfSecuritySchemeOrRefValuesItem(
		securityName,
		openapi3.SecuritySchemeOrRef{
			SecurityScheme: &openapi3.SecurityScheme{
				HTTPSecurityScheme: (&openapi3.HTTPSecurityScheme{}).WithScheme("basic").WithDescription("Admin Access"),
			},
		},
	)

	op := openapi3.Operation{}
	_ = reflector.SetJSONResponse(&op, struct {
		Secret string `json:"secret"`
	}{}, http.StatusOK)

	// Add security requirement to operation.
	op.Security = append(op.Security, map[string][]string{securityName: {}})

	// Describe unauthorized response.
	_ = reflector.SetJSONResponse(&op, struct {
		Error string `json:"error"`
	}{}, http.StatusUnauthorized)

	// Add operation to schema.
	_ = reflector.SpecEns().AddOperation(http.MethodGet, "/secure", op)

	schema, err := reflector.Spec.MarshalYAML()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(schema))

	// Output:
	// openapi: 3.0.3
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

func ExampleReflector_SetJSONResponse_api_key_auth() {
	reflector := openapi3.Reflector{}
	securityName := "api_key"

	// Declare security scheme.
	reflector.SpecEns().ComponentsEns().SecuritySchemesEns().WithMapOfSecuritySchemeOrRefValuesItem(
		securityName,
		openapi3.SecuritySchemeOrRef{
			SecurityScheme: &openapi3.SecurityScheme{
				APIKeySecurityScheme: (&openapi3.APIKeySecurityScheme{}).
					WithName("Authorization").
					WithIn("header").
					WithDescription("API Access"),
			},
		},
	)

	op := openapi3.Operation{}
	_ = reflector.SetJSONResponse(&op, struct {
		Secret string `json:"secret"`
	}{}, http.StatusOK)

	// Add security requirement to operation.
	op.Security = append(op.Security, map[string][]string{securityName: {}})

	// Describe unauthorized response.
	_ = reflector.SetJSONResponse(&op, struct {
		Error string `json:"error"`
	}{}, http.StatusUnauthorized)

	// Add operation to schema.
	_ = reflector.SpecEns().AddOperation(http.MethodGet, "/secure", op)

	schema, err := reflector.Spec.MarshalYAML()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(schema))

	// Output:
	// openapi: 3.0.3
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

func ExampleReflector_SetJSONResponse_http_bearer_token_auth() {
	reflector := openapi3.Reflector{}
	securityName := "bearer_token"

	// Declare security scheme.
	reflector.SpecEns().ComponentsEns().SecuritySchemesEns().WithMapOfSecuritySchemeOrRefValuesItem(
		securityName,
		openapi3.SecuritySchemeOrRef{
			SecurityScheme: &openapi3.SecurityScheme{
				HTTPSecurityScheme: (&openapi3.HTTPSecurityScheme{}).
					WithScheme("bearer").
					WithBearerFormat("JWT").
					WithDescription("Admin Access"),
			},
		},
	)

	op := openapi3.Operation{}
	_ = reflector.SetJSONResponse(&op, struct {
		Secret string `json:"secret"`
	}{}, http.StatusOK)

	// Add security requirement to operation.
	op.Security = append(op.Security, map[string][]string{securityName: {}})

	// Describe unauthorized response.
	_ = reflector.SetJSONResponse(&op, struct {
		Error string `json:"error"`
	}{}, http.StatusUnauthorized)

	// Add operation to schema.
	_ = reflector.SpecEns().AddOperation(http.MethodGet, "/secure", op)

	schema, err := reflector.Spec.MarshalYAML()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(schema))

	// Output:
	// openapi: 3.0.3
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
