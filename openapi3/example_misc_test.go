package openapi3_test

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/swaggest/assertjson"
	"github.com/swaggest/jsonschema-go"
	"github.com/swaggest/openapi-go/openapi3"
)

func ExampleReflector_options() {
	r := openapi3.Reflector{}

	// Reflector embeds jsonschema.Reflector and it is possible to configure optional behavior.
	r.Reflector.DefaultOptions = append(r.Reflector.DefaultOptions,
		jsonschema.InterceptNullability(func(params jsonschema.InterceptNullabilityParams) {
			// Removing nullability from non-pointer slices (regardless of omitempty).
			if params.Type.Kind() != reflect.Ptr && params.Schema.HasType(jsonschema.Null) && params.Schema.HasType(jsonschema.Array) {
				*params.Schema.Type = jsonschema.Array.Type()
			}
		}))

	type req struct {
		Foo []int `json:"foo"`
	}

	oc := openapi3.OperationContext{
		Operation: &openapi3.Operation{},
		Input:     new(req),
	}

	_ = r.SetupRequest(oc)
	_ = r.SpecEns().AddOperation(http.MethodGet, "/foo", *oc.Operation)

	j, _ := assertjson.MarshalIndentCompact(r.Spec, "", " ", 120)

	fmt.Println(string(j))

	// Output:
	// {
	//  "openapi":"3.0.3","info":{"title":"","version":""},
	//  "paths":{
	//   "/foo":{
	//    "get":{
	//     "requestBody":{"content":{"application/json":{"schema":{"$ref":"#/components/schemas/Openapi3TestReq"}}}},
	//     "responses":{"204":{"description":"No Content"}}
	//    }
	//   }
	//  },
	//  "components":{"schemas":{"Openapi3TestReq":{"type":"object","properties":{"foo":{"type":"array","items":{"type":"integer"}}}}}}
	// }
}
