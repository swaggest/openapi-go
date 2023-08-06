package openapi31_test

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/swaggest/assertjson"
	"github.com/swaggest/jsonschema-go"
	"github.com/swaggest/openapi-go/openapi31"
)

func ExampleReflector_options() {
	r := openapi31.Reflector{}

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

	oc, _ := r.NewOperationContext(http.MethodPost, "/foo")
	oc.AddReqStructure(new(req))

	_ = r.AddOperation(oc)

	j, _ := assertjson.MarshalIndentCompact(r.Spec, "", " ", 120)

	fmt.Println(string(j))

	// Output:
	// {
	//  "openapi":"3.1.0","info":{"title":"","version":""},
	//  "paths":{
	//   "/foo":{
	//    "post":{
	//     "requestBody":{"content":{"application/json":{"schema":{"$ref":"#/components/schemas/Openapi31TestReq"}}}},
	//     "responses":{"204":{"description":"No Content"}}
	//    }
	//   }
	//  },
	//  "components":{"schemas":{"Openapi31TestReq":{"properties":{"foo":{"items":{"type":"integer"},"type":"array"}},"type":"object"}}}
	// }
}
