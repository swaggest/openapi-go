package openapi3_test

import (
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/swaggest/assertjson"
	"github.com/swaggest/jsonschema-go"
	"github.com/swaggest/openapi-go"
	"github.com/swaggest/openapi-go/openapi3"
)

type WeirdResp interface {
	Boo()
}

type UUID [16]byte

type EmbeddedHeader struct {
	HeaderField string `header:"X-Header-Field" description:"Sample header response."`
}

type Resp struct {
	EmbeddedHeader
	Field1 int    `json:"field1"`
	Field2 string `json:"field2"`
	Info   struct {
		Foo string  `json:"foo" default:"baz" required:"true" pattern:"\\d+"`
		Bar float64 `json:"bar" description:"This is Bar."`
	} `json:"info"`
	Parent               *Resp                  `json:"parent,omitempty"`
	Map                  map[string]int64       `json:"map,omitempty"`
	MapOfAnything        map[string]interface{} `json:"mapOfAnything,omitempty"`
	ArrayOfAnything      []interface{}          `json:"arrayOfAnything,omitempty"`
	Whatever             interface{}            `json:"whatever"`
	NullableWhatever     *interface{}           `json:"nullableWhatever,omitempty"`
	RecursiveArray       []WeirdResp            `json:"recursiveArray,omitempty"`
	RecursiveStructArray []Resp                 `json:"recursiveStructArray,omitempty"`
	UUID                 UUID                   `json:"uuid"`
}

func (r *Resp) Description() string {
	return "This is a sample response."
}

func (r *Resp) Title() string {
	return "Sample Response"
}

var _ jsonschema.Preparer = &Resp{}

func (r *Resp) PrepareJSONSchema(s *jsonschema.Schema) error {
	s.WithExtraPropertiesItem("x-foo", "bar")

	return nil
}

type Req struct {
	InQuery1       int                   `query:"in_query1" required:"true" description:"Query parameter."`
	InQuery2       int                   `query:"in_query2" required:"true" description:"Query parameter."`
	InQuery3       int                   `query:"in_query3" required:"true" description:"Query parameter."`
	InPath         int                   `path:"in_path"`
	InCookie       string                `cookie:"in_cookie" deprecated:"true"`
	InHeader       float64               `header:"in_header"`
	InBody1        int                   `json:"in_body1"`
	InBody2        string                `json:"in_body2"`
	InForm1        string                `formData:"in_form1"`
	InForm2        string                `formData:"in_form2"`
	File1          multipart.File        `formData:"upload1"`
	File2          *multipart.FileHeader `formData:"upload2"`
	UUID           UUID                  `header:"uuid"`
	ArrayCSV       []string              `query:"array_csv" explode:"false"`
	ArraySwg2CSV   []string              `query:"array_swg2_csv" collectionFormat:"csv"`
	ArraySwg2SSV   []string              `query:"array_swg2_ssv" collectionFormat:"ssv"`
	ArraySwg2Pipes []string              `query:"array_swg2_pipes" collectionFormat:"pipes"`
}

func (r Req) ForceJSONRequestBody() {}

type GetReq struct {
	InQuery1 int     `query:"in_query1" required:"true" description:"Query parameter." json:"q1"`
	InQuery3 int     `query:"in_query3" required:"true" description:"Query parameter." json:"q3"`
	InPath   int     `path:"in_path" json:"p"`
	InCookie string  `cookie:"in_cookie" deprecated:"true" json:"c"`
	InHeader float64 `header:"in_header" json:"h"`
}

const (
	apiName    = "SampleAPI"
	apiVersion = "1.2.3"
)

func TestReflector_AddOperation_request_array(t *testing.T) {
	reflector := openapi3.Reflector{}

	s := reflector.SpecEns()
	s.Info.Title = apiName
	s.Info.Version = apiVersion

	oc, err := reflector.NewOperationContext(http.MethodPost, "/somewhere")
	require.NoError(t, err)

	oc.AddReqStructure(new([]GetReq))

	require.NoError(t, reflector.AddOperation(oc))

	b, err := assertjson.MarshalIndentCompact(s, "", " ", 120)
	require.NoError(t, err)

	require.NoError(t, os.WriteFile("testdata/openapi_req_array_last_run.json", b, 0o600))

	expected, err := os.ReadFile("testdata/openapi_req_array.json")
	require.NoError(t, err)

	assertjson.Equal(t, expected, b)
}

func TestReflector_AddOperation_uploadInterface(t *testing.T) {
	type req struct {
		File1 multipart.File `formData:"upload1"`
	}

	reflector := openapi3.Reflector{}
	oc, err := reflector.NewOperationContext(http.MethodPost, "/somewhere")
	require.NoError(t, err)

	oc.AddReqStructure(new(req))

	require.NoError(t, reflector.AddOperation(oc))

	assertjson.EqMarshal(t, `{
	  "openapi":"3.0.3","info":{"title":"","version":""},
	  "paths":{
		"/somewhere":{
		  "post":{
			"requestBody":{
			  "content":{
				"multipart/form-data":{"schema":{"$ref":"#/components/schemas/FormDataOpenapi3TestReq"}}
			  }
			},
			"responses":{"204":{"description":"No Content"}}
		  }
		}
	  },
	  "components":{
		"schemas":{
		  "FormDataOpenapi3TestReq":{
			"type":"object",
			"properties":{"upload1":{"$ref":"#/components/schemas/MultipartFile"}}
		  },
		  "MultipartFile":{"type":"string","format":"binary"}
		}
	  }
	}`, reflector.Spec)
}

func TestReflector_AddOperation_request(t *testing.T) {
	reflector := openapi3.Reflector{}

	s := reflector.SpecSchema()
	s.SetTitle(apiName)
	s.SetVersion(apiVersion)
	s.SetDescription("This a sample API description.")

	oc, err := reflector.NewOperationContext(http.MethodGet, "/somewhere/{in_path}")
	require.NoError(t, err)
	oc.AddReqStructure(new(GetReq))
	oc.AddReqStructure(nil, func(cu *openapi.ContentUnit) {
		cu.ContentType = "text/csv"
		cu.Description = "Request body in CSV format."
	})

	require.NoError(t, reflector.AddOperation(oc))

	b, err := assertjson.MarshalIndentCompact(s, "", " ", 120)
	require.NoError(t, err)

	require.NoError(t, os.WriteFile("testdata/openapi_req_last_run.json", b, 0o600))

	expected, err := os.ReadFile("testdata/openapi_req.json")
	require.NoError(t, err)

	assertjson.Equal(t, expected, b)
}

func TestReflector_AddOperation_JSON_response(t *testing.T) {
	reflector := openapi3.Reflector{}

	s := reflector.SpecEns()
	s.Info.Title = apiName
	s.Info.Version = apiVersion

	reflector.AddTypeMapping(new(WeirdResp), new(Resp))

	oc, err := reflector.NewOperationContext(http.MethodPost, "/somewhere/{in_path}")
	require.NoError(t, err)

	oc.AddReqStructure(new(Req))
	oc.AddRespStructure(new(WeirdResp), func(cu *openapi.ContentUnit) {
		cu.HTTPStatus = http.StatusOK
	})
	oc.AddRespStructure(new([]WeirdResp), func(cu *openapi.ContentUnit) {
		cu.HTTPStatus = http.StatusConflict
	})
	oc.AddRespStructure("", func(cu *openapi.ContentUnit) {
		cu.HTTPStatus = http.StatusConflict
		cu.ContentType = "text/html"
	})

	pathItem := openapi3.PathItem{}
	pathItem.
		WithSummary("Path Summary").
		WithDescription("Path Description")
	s.Paths.WithMapOfPathItemValuesItem("/somewhere/{in_path}", pathItem)

	require.NoError(t, reflector.AddOperation(oc))

	require.NoError(t, s.SetupOperation(http.MethodPost, "/somewhere/{in_path}", func(op *openapi3.Operation) error {
		js := op.RequestBody.RequestBody.Content["multipart/form-data"].Schema.ToJSONSchema(s)
		expected, err := os.ReadFile("testdata/req_schema.json")
		require.NoError(t, err)
		assertjson.EqualMarshal(t, expected, js)

		return nil
	}))

	oc, err = reflector.NewOperationContext(http.MethodGet, "/somewhere/{in_path}")
	require.NoError(t, err)

	oc.AddReqStructure(new(GetReq))
	oc.AddRespStructure(new(Resp))

	require.NoError(t, reflector.AddOperation(oc))

	require.NoError(t, s.SetupOperation(http.MethodGet, "/somewhere/{in_path}", func(op *openapi3.Operation) error {
		js := op.Responses.MapOfResponseOrRefValues[strconv.Itoa(http.StatusOK)].Response.Content["application/json"].
			Schema.ToJSONSchema(s)
		jsb, err := assertjson.MarshalIndentCompact(js, "", " ", 120)
		require.NoError(t, err)

		require.NoError(t, os.WriteFile("testdata/resp_schema_last_run.json", jsb, 0o600))

		expected, err := os.ReadFile("testdata/resp_schema.json")
		require.NoError(t, err)
		assertjson.EqualMarshal(t, expected, js)

		js = op.Responses.MapOfResponseOrRefValues[strconv.Itoa(http.StatusOK)].Response.Headers["X-Header-Field"].Header.
			Schema.ToJSONSchema(s)
		assertjson.EqMarshal(t, `{"type": "string", "description": "Sample header response."}`, js)

		require.NoError(t, err)
		assertjson.EqMarshal(t, `{"type": "integer", "description": "Query parameter."}`,
			op.Parameters[0].Parameter.Schema.ToJSONSchema(s))

		return nil
	}))

	b, err := assertjson.MarshalIndentCompact(s, "", " ", 120)
	require.NoError(t, err)

	require.NoError(t, os.WriteFile("testdata/openapi_last_run.json", b, 0o600))

	expected, err := os.ReadFile("testdata/openapi.json")
	require.NoError(t, err)

	assertjson.EqualMarshal(t, expected, s)
}

type Identity struct {
	ID string `path:"id"`
}

type Data []string

type PathParamAndBody struct {
	Identity
	Data
}

func TestReflector_AddOperation_pathParamAndBody(t *testing.T) {
	reflector := openapi3.Reflector{}

	oc, err := reflector.NewOperationContext(http.MethodPost, "/somewhere/{id}")
	require.NoError(t, err)

	oc.AddReqStructure(new(PathParamAndBody))

	require.NoError(t, reflector.AddOperation(oc))

	s := reflector.SpecEns()
	s.Info.Title = apiName
	s.Info.Version = apiVersion

	assertjson.EqMarshal(t, `{
	 "openapi":"3.0.3","info":{"title":"SampleAPI","version":"1.2.3"},
	 "paths":{
	  "/somewhere/{id}":{
	   "post":{
		"parameters":[{"name":"id","in":"path","required":true,"schema":{"type":"string"}}],
		"requestBody":{
		 "content":{"application/json":{"schema":{"$ref":"#/components/schemas/Openapi3TestPathParamAndBody"}}}
		},
		"responses":{"204":{"description":"No Content"}}
	   }
	  }
	 },
	 "components":{
	  "schemas":{"Openapi3TestPathParamAndBody":{"type":"array","items":{"type":"string"},"nullable":true}}
	 }
	}`, s)
}

type WithReqBody PathParamAndBody

func (*WithReqBody) ForceRequestBody() {}

func TestReflector_AddOperation_RequestBodyEnforcer(t *testing.T) {
	reflector := openapi3.Reflector{}
	oc, err := reflector.NewOperationContext(http.MethodGet, "/somewhere/{id}")
	require.NoError(t, err)

	oc.AddReqStructure(new(WithReqBody))

	s := reflector.SpecEns()
	s.Info.Title = apiName
	s.Info.Version = apiVersion

	require.NoError(t, reflector.AddOperation(oc))

	assertjson.EqMarshal(t, `{
	 "openapi":"3.0.3","info":{"title":"SampleAPI","version":"1.2.3"},
	 "paths":{
	  "/somewhere/{id}":{
	   "get":{
		"parameters":[{"name":"id","in":"path","required":true,"schema":{"type":"string"}}],
		"requestBody":{"content":{"application/json":{"schema":{"$ref":"#/components/schemas/Openapi3TestWithReqBody"}}}},
		"responses":{"204":{"description":"No Content"}}
	   }
	  }
	 },
	 "components":{"schemas":{"Openapi3TestWithReqBody":{"type":"array","items":{"type":"string"},"nullable":true}}}
	}`, s)
}

func TestReflector_AddOperation_response(t *testing.T) {
	reflector := openapi3.Reflector{}
	oc, err := reflector.NewOperationContext(http.MethodGet, "/somewhere")
	require.NoError(t, err)

	s := reflector.SpecEns()
	s.Info.Title = apiName
	s.Info.Version = apiVersion

	oc.AddRespStructure(new(struct {
		Val1 int
		Val2 string
	}), func(cu *openapi.ContentUnit) {
		cu.ContentType = "text/csv; charset=utf-8"
		cu.HTTPStatus = http.StatusNoContent
		cu.SetFieldMapping(openapi.InHeader, map[string]string{
			"Val1": "X-Value-1",
			"Val2": "X-Value-2",
		})
	})

	require.NoError(t, reflector.AddOperation(oc))

	assertjson.EqMarshal(t, `{
	 "openapi":"3.0.3","info":{"title":"SampleAPI","version":"1.2.3"},
	 "paths":{
	  "/somewhere":{
	   "get":{
		"responses":{
		 "204":{
		  "description":"No Content",
		  "headers":{
		   "X-Value-1":{"style":"simple","schema":{"type":"integer"}},
		   "X-Value-2":{"style":"simple","schema":{"type":"string"}}
		  },
		  "content":{"text/csv":{"schema":{"type":"string"}}}
		 }
		}
	   }
	  }
	 }
	}`, s)
}

func TestReflector_AddOperation_setup_request(t *testing.T) {
	reflector := openapi3.Reflector{}
	oc, err := reflector.NewOperationContext(http.MethodPost, "/somewhere/{value-4}")
	require.NoError(t, err)

	s := reflector.SpecEns()
	s.Info.Title = apiName
	s.Info.Version = apiVersion

	oc.AddReqStructure(new(struct {
		Val1 int
		Val2 string
		Val3 float64
		Val4 bool
		Val5 string
		Val6 multipart.File
	}), func(cu *openapi.ContentUnit) {
		cu.SetFieldMapping(openapi.InHeader, map[string]string{
			"Val1": "X-Value-1",
		})
		cu.SetFieldMapping(openapi.InQuery, map[string]string{
			"Val2": "value_2",
		})
		cu.SetFieldMapping(openapi.InFormData, map[string]string{
			"Val3": "value3",
			"Val6": "upload6",
		})
		cu.SetFieldMapping(openapi.InPath, map[string]string{
			"Val4": "value-4",
		})
		cu.SetFieldMapping(openapi.InCookie, map[string]string{
			"Val5": "value_5",
		})
	})

	require.NoError(t, reflector.AddOperation(oc))

	assertjson.EqMarshal(t, `{
	  "openapi":"3.0.3","info":{"title":"SampleAPI","version":"1.2.3"},
	  "paths":{
		"/somewhere/{value-4}":{
		  "post":{
			"parameters":[
			  {"name":"value_2","in":"query","schema":{"type":"string"}},
			  {
				"name":"value-4","in":"path","required":true,
				"schema":{"type":"boolean"}
			  },
			  {"name":"value_5","in":"cookie","schema":{"type":"string"}},
			  {"name":"X-Value-1","in":"header","schema":{"type":"integer"}}
			],
			"requestBody":{
			  "content":{
				"multipart/form-data":{
				  "schema":{
					"type":"object",
					"properties":{
					  "upload6":{"$ref":"#/components/schemas/MultipartFile"},
					  "value3":{"type":"number"}
					}
				  }
				}
			  }
			},
			"responses":{"204":{"description":"No Content"}}
		  }
		}
	  },
	  "components":{
		"schemas":{"MultipartFile":{"type":"string","format":"binary"}}
	  }
	}`, s)
}

func TestReflector_AddOperation_request_queryObject(t *testing.T) {
	reflector := openapi3.Reflector{}

	oc, err := reflector.NewOperationContext(http.MethodPost, "/somewhere")
	require.NoError(t, err)

	s := reflector.SpecEns()
	s.Info.Title = apiName
	s.Info.Version = apiVersion

	oc.AddReqStructure(new(struct {
		InQuery map[int]float64 `query:"in_query"`
	}))

	require.NoError(t, reflector.AddOperation(oc))

	assertjson.EqMarshal(t, `{
	 "openapi":"3.0.3","info":{"title":"SampleAPI","version":"1.2.3"},
	 "paths":{
	  "/somewhere":{
	   "post":{
		"parameters":[
		 {
		  "name":"in_query","in":"query","style":"deepObject","explode":true,
		  "schema":{"type":"object","additionalProperties":{"type":"number"}}
		 }
		],
		"responses":{"204":{"description":"No Content"}}
	   }
	  }
	 }
	}`, s)
}

type namedType struct {
	InQuery labels `query:"in_query"`
}

type labels map[int]float64

func TestReflector_AddOperation_request_queryNamedObject(t *testing.T) {
	reflector := openapi3.Reflector{}
	oc, err := reflector.NewOperationContext(http.MethodPost, "/somewhere")
	require.NoError(t, err)

	s := reflector.SpecEns()
	s.Info.Title = apiName
	s.Info.Version = apiVersion

	oc.AddReqStructure(new(namedType))
	require.NoError(t, reflector.AddOperation(oc))

	assertjson.EqMarshal(t, `{
	  "openapi":"3.0.3","info":{"title":"SampleAPI","version":"1.2.3"},
	  "paths":{
		"/somewhere":{
		  "post":{
			"parameters":[
			  {
				"name":"in_query","in":"query","style":"deepObject","explode":true,
				"schema":{"$ref":"#/components/schemas/Openapi3TestLabels"}
			  }
			],
			"responses":{"204":{"description":"No Content"}}
		  }
		}
	  },
	  "components":{
		"schemas":{
		  "Openapi3TestLabels":{"type":"object","additionalProperties":{"type":"number"}}
		}
	  }
	}`, s)

	js, found := reflector.ResolveJSONSchemaRef("#/components/schemas/Openapi3TestLabels")
	assert.True(t, found)
	assertjson.EqMarshal(t, `{"type":"object","additionalProperties":{"type":"number"}}`, js)
}

func TestReflector_AddOperation_request_jsonQuery(t *testing.T) {
	type filter struct {
		Labels []string `json:"labels,omitempty"`
		Type   string   `json:"type"`
	}

	type req struct {
		One   filter         `query:"one"`
		Two   filter         `query:"two"`
		Three map[string]int `query:"three"`
		Four  map[string]int `query:"four" collectionFormat:"json"`
	}

	r := openapi3.Reflector{}
	oc, err := r.NewOperationContext(http.MethodGet, "/")
	require.NoError(t, err)

	oc.AddReqStructure(new(req))

	require.NoError(t, r.AddOperation(oc))

	assertjson.EqMarshal(t, `{
	  "openapi":"3.0.3","info":{"title":"","version":""},
	  "paths":{
		"/":{
		  "get":{
			"parameters":[
			  {
				"name":"one","in":"query",
				"content":{
				  "application/json":{"schema":{"$ref":"#/components/schemas/Openapi3TestFilter"}}
				}
			  },
			  {
				"name":"two","in":"query",
				"content":{
				  "application/json":{"schema":{"$ref":"#/components/schemas/Openapi3TestFilter"}}
				}
			  },
			  {
				"name":"three","in":"query","style":"deepObject","explode":true,
				"schema":{"type":"object","additionalProperties":{"type":"integer"}}
			  },
			  {
				"name":"four","in":"query",
				"content":{
				  "application/json":{
					"schema":{
					  "type":"object","additionalProperties":{"type":"integer"},
					  "nullable":true
					}
				  }
				}
			  }
			],
			"responses":{"204":{"description":"No Content"}}
		  }
		}
	  },
	  "components":{
		"schemas":{
		  "Openapi3TestFilter":{
			"type":"object",
			"properties":{
			  "labels":{"type":"array","items":{"type":"string"}},
			  "type":{"type":"string"}
			}
		  }
		}
	  }
	}`, r.SpecEns())
}

func TestReflector_AddOperation_request_forbidParams(t *testing.T) {
	type req struct {
		Query  string `query:"query"`
		Header string `header:"header"`
		Cookie string `cookie:"cookie"`
		Path   string `path:"path"`

		_ struct{} `query:"_" cookie:"_" additionalProperties:"false"`
	}

	r := openapi3.Reflector{}
	oc, err := r.NewOperationContext(http.MethodGet, "/{path}")
	require.NoError(t, err)

	oc.AddReqStructure(new(req))

	require.NoError(t, r.AddOperation(oc))

	assertjson.EqMarshal(t, `{
	  "openapi":"3.0.3","info":{"title":"","version":""},
	  "paths":{
		"/{path}":{
		  "get":{
			"parameters":[
			  {"name":"query","in":"query","schema":{"type":"string"}},
			  {
				"name":"path","in":"path","required":true,
				"schema":{"type":"string"}
			  },
			  {"name":"cookie","in":"cookie","schema":{"type":"string"}},
			  {"name":"header","in":"header","schema":{"type":"string"}}
			],
			"responses":{"204":{"description":"No Content"}},
			"x-forbid-unknown-cookie":true,"x-forbid-unknown-query":true
		  }
		}
	  }
	}`, r.SpecEns())

	o3, ok := oc.(openapi3.OperationExposer)
	require.True(t, ok)

	assert.True(t, o3.Operation().UnknownParamIsForbidden(openapi3.ParameterInCookie))
	assert.False(t, o3.Operation().UnknownParamIsForbidden(openapi3.ParameterInHeader))
	assert.True(t, o3.Operation().UnknownParamIsForbidden(openapi3.ParameterInQuery))
	assert.False(t, o3.Operation().UnknownParamIsForbidden(openapi3.ParameterInPath))
}

func TestReflector_AddOperation_request_noBody(t *testing.T) {
	type req struct {
		ID int `json:"id" path:"id"`
	}

	r := openapi3.Reflector{}

	for _, method := range []string{http.MethodHead, http.MethodGet, http.MethodDelete, http.MethodTrace} {
		oc, err := r.NewOperationContext(method, "/{id}")
		require.NoError(t, err)

		oc.AddReqStructure(new(req))

		require.NoError(t, r.AddOperation(oc))

		require.NoError(t, r.SpecEns().SetupOperation(method, "/{id}", func(op *openapi3.Operation) error {
			assertjson.EqMarshal(t, `{
			  "parameters":[{"name":"id","in":"path","required":true,"schema":{"type":"integer"}}],
			  "responses":{"204":{"description":"No Content"}}
			}`, op)

			return nil
		}))
	}

	for _, method := range []string{http.MethodPost, http.MethodPatch, http.MethodPut} {
		oc, err := r.NewOperationContext(method, "/{id}")
		require.NoError(t, err)

		oc.AddReqStructure(new(req))

		require.NoError(t, r.AddOperation(oc))

		require.NoError(t, r.SpecEns().SetupOperation(method, "/{id}", func(op *openapi3.Operation) error {
			assertjson.EqMarshal(t, `{
			  "parameters":[{"name":"id","in":"path","required":true,"schema":{"type":"integer"}}],
			  "requestBody":{
				"content":{
				  "application/json":{"schema":{"$ref":"#/components/schemas/Openapi3TestReq"}}
				}
			  },
			  "responses":{"204":{"description":"No Content"}}
			}`, op)

			return nil
		}))
	}
}

func TestReflector_AddOperation_OperationCtx(t *testing.T) {
	type req struct {
		Query  string `query:"query"`
		Header string `header:"header"`
		Cookie string `cookie:"cookie"`
		Path   string `path:"path"`
		Body   string `json:"body"`
	}

	type resp struct {
		Header string `header:"header"`
		Body   string `json:"body"`
	}

	r := openapi3.Reflector{}
	oc, err := r.NewOperationContext(http.MethodPost, "/{path}")
	require.NoError(t, err)

	oc.AddReqStructure(new(req))
	oc.AddRespStructure(new(resp))

	visited := map[string]bool{}

	var currentRC *jsonschema.ReflectContext

	r.DefaultOptions = append(r.DefaultOptions,
		func(rc *jsonschema.ReflectContext) {
			currentRC = rc
		},
		jsonschema.InterceptSchema(func(_ jsonschema.InterceptSchemaParams) (stop bool, err error) {
			if occ, ok := openapi.OperationCtx(currentRC); ok {
				if occ.IsProcessingResponse() {
					visited["resp:"+string(occ.ProcessingIn())] = true
				} else {
					visited["req:"+string(occ.ProcessingIn())] = true
				}
			}

			return false, nil
		}),
	)

	require.NoError(t, r.AddOperation(oc))

	assert.Equal(t, map[string]bool{
		"req:body":    true,
		"req:cookie":  true,
		"req:header":  true,
		"req:path":    true,
		"req:query":   true,
		"resp:body":   true,
		"resp:header": true,
	}, visited)
}

func TestReflector_AddOperation_request_formData_with_json(t *testing.T) {
	// In presence of `formData` tags, `json` tags would be ignored as request body.
	type req struct {
		Foo int `formData:"foo" json:"foo"`
	}

	r := openapi3.Reflector{}
	oc, err := r.NewOperationContext(http.MethodPost, "/foo")
	require.NoError(t, err)

	oc.AddReqStructure(new(req))
	require.NoError(t, r.AddOperation(oc))

	assertjson.EqMarshal(t, `{
	  "openapi":"3.0.3","info":{"title":"","version":""},
	  "paths":{
		"/foo":{
		  "post":{
			"requestBody":{
			  "content":{
				"application/x-www-form-urlencoded":{"schema":{"$ref":"#/components/schemas/FormDataOpenapi3TestReq"}}
			  }
			},
			"responses":{"204":{"description":"No Content"}}
		  }
		}
	  },
	  "components":{
		"schemas":{
		  "FormDataOpenapi3TestReq":{"type":"object","properties":{"foo":{"type":"integer"}}}
		}
	  }
	}`, r.SpecEns())
}

func TestReflector_AddOperation_request_form(t *testing.T) {
	// Fields with `form` will be present as both `query` parameter and `formData` property.
	type req struct {
		Foo  int    `formData:"foo"`
		Bar  int    `form:"bar"`
		Baz  string `query:"baz"`
		Quux string `query:"quux" form:"ignored"`
	}

	r := openapi3.Reflector{}
	oc, err := r.NewOperationContext(http.MethodPost, "/foo")
	require.NoError(t, err)

	oc.AddReqStructure(req{})

	require.NoError(t, r.AddOperation(oc))

	assertjson.EqMarshal(t, `{
	  "openapi":"3.0.3","info":{"title":"","version":""},
	  "paths":{
		"/foo":{
		  "post":{
			"parameters":[
			  {"name":"bar","in":"query","schema":{"type":"integer"}},
			  {"name":"baz","in":"query","schema":{"type":"string"}},
			  {"name":"quux","in":"query","schema":{"type":"string"}}
			],
			"requestBody":{
			  "content":{
				"application/x-www-form-urlencoded":{"schema":{"$ref":"#/components/schemas/FormDataOpenapi3TestReq"}}
			  }
			},
			"responses":{"204":{"description":"No Content"}}
		  }
		}
	  },
	  "components":{
		"schemas":{
		  "FormDataOpenapi3TestReq":{
			"type":"object",
			"properties":{
			  "bar":{"type":"integer"},"foo":{"type":"integer"},
			  "ignored":{"type":"string"}
			}
		  }
		}
	  }
	}`, r.SpecEns())
}

func TestReflector_AddOperation_request_form_only(t *testing.T) {
	// Fields with `form` will be present as both `query` parameter and `formData` property.
	type req struct {
		Bar  int    `form:"bar"`
		Quux string `form:"quux"`
	}

	r := openapi3.Reflector{}
	oc, err := r.NewOperationContext(http.MethodPost, "/foo")
	require.NoError(t, err)

	oc.AddReqStructure(req{})

	require.NoError(t, r.AddOperation(oc))

	assertjson.EqMarshal(t, `{
	  "openapi":"3.0.3","info":{"title":"","version":""},
	  "paths":{
		"/foo":{
		  "post":{
			"parameters":[
			  {"name":"bar","in":"query","schema":{"type":"integer"}},
			  {"name":"quux","in":"query","schema":{"type":"string"}}
			],
			"requestBody":{
			  "content":{
				"application/x-www-form-urlencoded":{"schema":{"$ref":"#/components/schemas/FormDataOpenapi3TestReq"}}
			  }
			},
			"responses":{"204":{"description":"No Content"}}
		  }
		}
	  },
	  "components":{
		"schemas":{
		  "FormDataOpenapi3TestReq":{
			"type":"object",
			"properties":{"bar":{"type":"integer"},"quux":{"type":"string"}}
		  }
		}
	  }
	}`, r.SpecEns())
}

func TestReflector_AddOperation_request_queryObject_deepObject(t *testing.T) {
	reflector := openapi3.Reflector{}

	// JSON object is only enabled when at least one `json` tag is available on top-level property,
	// and there are no `in` tags, e.g. `query`.
	type jsonFilter struct {
		Foo    string `json:"foo"`
		Bar    int    `json:"bar"`
		Deeper struct {
			Val string `json:"val"`
		} `json:"deeper"`
	}

	// Deep object structure may have `json` tags, they are ignored in presence of
	// at least one top-level field with a matching `in` tag, e.g. `query`.
	type deepObjectFilter struct {
		Baz    bool    `json:"baz" query:"baz"`
		Quux   float64 `json:"quux" query:"quux"`
		Deeper struct {
			Val string `query:"val"`
		} `query:"deeper"`
	}

	type EmbeddedParams struct {
		Embedded string `query:"embedded"`
	}

	type DeeplyEmbedded struct {
		EmbeddedParams
	}

	type req struct {
		// Simple scalar parameters.
		ID     string `path:"id" example:"XXX-XXXXX"`
		Locale string `query:"locale" pattern:"^[a-z]{2}-[A-Z]{2}$"`
		// Embedded types are expanded into top-level scope.
		DeeplyEmbedded
		// Object values can be serialized in JSON (with json field tags in the value struct).
		JSONFilter jsonFilter `query:"json_filter"`
		// Or as deepObject (with same field tag as parent, .e.g query).
		DeepObjectFilter deepObjectFilter `query:"deep_object_filter"`
		// JSON body tags are ignored for GET request by default.
		Amount uint `json:"amount"`
	}

	oc, err := reflector.NewOperationContext(http.MethodGet, "/things/{id}")
	require.NoError(t, err)

	oc.AddReqStructure(new(req))

	require.NoError(t, reflector.AddOperation(oc))

	assertjson.EqMarshal(t, `{
	  "openapi":"3.0.3","info":{"title":"","version":""},
	  "paths":{
		"/things/{id}":{
		  "get":{
			"parameters":[
			  {
				"name":"locale","in":"query",
				"schema":{"pattern":"^[a-z]{2}-[A-Z]{2}$","type":"string"}
			  },
			  {"name":"embedded","in":"query","schema":{"type":"string"}},
			  {
				"name":"json_filter","in":"query",
				"content":{
				  "application/json":{
					"schema":{"$ref":"#/components/schemas/Openapi3TestJsonFilter"}
				  }
				}
			  },
			  {
				"name":"deep_object_filter","in":"query","style":"deepObject",
				"explode":true,
				"schema":{"$ref":"#/components/schemas/Openapi3TestDeepObjectFilter"}
			  },
			  {
				"name":"id","in":"path","required":true,
				"schema":{"type":"string","example":"XXX-XXXXX"}
			  }
			],
			"responses":{"204":{"description":"No Content"}}
		  }
		}
	  },
	  "components":{
		"schemas":{
		  "Openapi3TestDeepObjectFilter":{
			"type":"object",
			"properties":{
			  "baz":{"type":"boolean"},
			  "deeper":{"type":"object","properties":{"val":{"type":"string"}}},
			  "quux":{"type":"number"}
			}
		  },
		  "Openapi3TestJsonFilter":{
			"type":"object",
			"properties":{
			  "bar":{"type":"integer"},
			  "deeper":{"type":"object","properties":{"val":{"type":"string"}}},
			  "foo":{"type":"string"}
			}
		  }
		}
	  }
	}`, reflector.Spec)
}

type textCSV struct{}

func (t textCSV) SetupContentUnit(cu *openapi.ContentUnit) {
	cu.ContentType = "text/csv"
	cu.Description = "This is CSV."
	cu.IsDefault = true
}

func TestReflector_AddOperation_contentUnitPreparer(t *testing.T) {
	r := openapi3.NewReflector()
	oc, err := r.NewOperationContext(http.MethodPost, "/foo")
	require.NoError(t, err)

	oc.AddReqStructure(textCSV{})
	oc.AddRespStructure(textCSV{})

	require.NoError(t, r.AddOperation(oc))

	assertjson.EqMarshal(t, `{
	  "openapi":"3.0.3","info":{"title":"","version":""},
	  "paths":{
		"/foo":{
		  "post":{
			"requestBody":{
			  "description":"This is CSV.",
			  "content":{"text/csv":{"schema":{"type":"string"}}}
			},
			"responses":{
			  "default":{
				"description":"This is CSV.",
				"content":{"text/csv":{"schema":{"type":"string"}}}
			  }
			}
		  }
		}
	  }
	}`, r.SpecSchema())
}

func TestReflector_AddOperation_defName(t *testing.T) {
	r := openapi3.NewReflector()
	oc, err := r.NewOperationContext(http.MethodPost, "/foo")
	require.NoError(t, err)

	type (
		simpleString  string
		specialString string
	)

	specialStringDef := jsonschema.Schema{}
	specialStringDef.AddType(jsonschema.String)
	specialStringDef.WithExamples("xy5abcd4sq9s")
	specialStringDef.MinLength = 12
	specialStringDef.WithMaxLength(12)
	specialStringDef.WithDescription("Very special.")
	r.AddTypeMapping(specialString(""), specialStringDef)

	type reqForm struct {
		Simple  simpleString  `formData:"simple"`
		Special specialString `formData:"special"`
		Foo     int           `formData:"foo"`
		Bar     float64       `header:"bar"`
	}

	type reqJSON struct {
		Simple  simpleString  `json:"simple"`
		Special specialString `json:"special"`
		Foo     int           `json:"foo"`
	}

	oc.AddReqStructure(reqForm{})
	oc.AddReqStructure(reqJSON{})

	require.NoError(t, r.AddOperation(oc))

	assertjson.EqMarshal(t, `{
	  "openapi":"3.0.3","info":{"title":"","version":""},
	  "paths":{
		"/foo":{
		  "post":{
			"parameters":[{"name":"bar","in":"header","schema":{"type":"number"}}],
			"requestBody":{
			  "content":{
				"application/json":{"schema":{"$ref":"#/components/schemas/Openapi3TestReqJSON"}},
				"application/x-www-form-urlencoded":{
				  "schema":{"$ref":"#/components/schemas/FormDataOpenapi3TestReqForm"}
				}
			  }
			},
			"responses":{"204":{"description":"No Content"}}
		  }
		}
	  },
	  "components":{
		"schemas":{
		  "FormDataOpenapi3TestReqForm":{
			"type":"object",
			"properties":{
			  "foo":{"type":"integer"},"simple":{"type":"string"},
			  "special":{"$ref":"#/components/schemas/Openapi3TestSpecialString"}
			}
		  },
		  "Openapi3TestReqJSON":{
			"type":"object",
			"properties":{
			  "foo":{"type":"integer"},"simple":{"type":"string"},
			  "special":{"$ref":"#/components/schemas/Openapi3TestSpecialString"}
			}
		  },
		  "Openapi3TestSpecialString":{
			"maxLength":12,"minLength":12,"type":"string",
			"description":"Very special.","example":"xy5abcd4sq9s"
		  }
		}
	  }
	}`, r.Spec)
}

func TestReflector_AddOperation_jsonschemaStruct(t *testing.T) {
	r := openapi3.NewReflector()

	oc, err := r.NewOperationContext(http.MethodPost, "/foo/{id}")
	require.NoError(t, err)

	type Req struct {
		ID int `path:"id"`
		jsonschema.Struct
	}

	req := Req{}
	req.DefName = "FooStruct"
	req.Fields = append(req.Fields, jsonschema.Field{
		Name:  "Foo",
		Value: "abc",
		Tag:   `json:"foo" minLength:"3"`,
	})

	type Resp struct {
		ID int `json:"id"`
		jsonschema.Struct
		Nested jsonschema.Struct `json:"nested"`
	}

	resp := Resp{}
	resp.DefName = "BarStruct"
	resp.Fields = append(resp.Fields, jsonschema.Field{
		Name:  "Bar",
		Value: "cba",
		Tag:   `json:"bar" maxLength:"3"`,
	})
	resp.Nested.DefName = "BazStruct"
	resp.Nested.Fields = append(resp.Nested.Fields, jsonschema.Field{
		Name:  "Baz",
		Value: "def",
		Tag:   `json:"baz" maxLength:"5"`,
	})

	oc.AddReqStructure(req)
	oc.AddRespStructure(resp)

	require.NoError(t, r.AddOperation(oc))

	assertjson.EqMarshal(t, `{
	  "openapi":"3.0.3","info":{"title":"","version":""},
	  "paths":{
		"/foo/{id}":{
		  "post":{
			"parameters":[
			  {
				"name":"id","in":"path","required":true,"schema":{"type":"integer"}
			  }
			],
			"requestBody":{
			  "content":{
				"application/json":{"schema":{"$ref":"#/components/schemas/FooStruct"}}
			  }
			},
			"responses":{
			  "200":{
				"description":"OK",
				"content":{
				  "application/json":{"schema":{"$ref":"#/components/schemas/BarStruct"}}
				}
			  }
			}
		  }
		}
	  },
	  "components":{
		"schemas":{
		  "BarStruct":{
			"properties":{
			  "bar":{"maxLength":3,"type":"string"},"id":{"type":"integer"},
			  "nested":{"$ref":"#/components/schemas/BazStruct"}
			},
			"type":"object"
		  },
		  "BazStruct":{"properties":{"baz":{"maxLength":5,"type":"string"}},"type":"object"},
		  "FooStruct":{"properties":{"foo":{"minLength":3,"type":"string"}},"type":"object"}
		}
	  }
	}`, r.SpecSchema())
}

func TestNewReflector_examples(t *testing.T) {
	r := openapi3.NewReflector()

	op, err := r.NewOperationContext(http.MethodGet, "/")
	require.NoError(t, err)

	type O1 struct {
		F1   int    `json:"f1,omitempty"`
		Code string `json:"code"`
	}

	type O2 struct {
		F2   int    `json:"f2,omitempty"`
		Code string `json:"code"`
	}

	st := http.StatusCreated

	op.AddRespStructure(jsonschema.OneOf(O1{}, O2{}), func(cu *openapi.ContentUnit) {
		cu.HTTPStatus = st
	})

	if o3, ok := op.(openapi3.OperationExposer); ok {
		c := openapi3.MediaType{}
		c.Examples = map[string]openapi3.ExampleOrRef{
			"responseOne": {
				Example: (&openapi3.Example{}).WithSummary("First possible answer").WithValue(O1{Code: "1234567890123456789012"}),
			},
			"responseTwo": {
				Example: (&openapi3.Example{}).WithSummary("Other possible answer").WithValue(O2{Code: "0000000000000000000000"}),
			},
		}
		resp := openapi3.Response{}
		resp.WithContentItem("application/json", c)

		o3.Operation().Responses.WithMapOfResponseOrRefValuesItem(strconv.Itoa(st), openapi3.ResponseOrRef{
			Response: &resp,
		})
	}

	require.NoError(t, r.AddOperation(op))
	assertjson.EqMarshal(t, `{
	  "openapi":"3.0.3","info":{"title":"","version":""},
	  "paths":{
		"/":{
		  "get":{
			"responses":{
			  "201":{
				"description":"Created",
				"content":{
				  "application/json":{
					"schema":{
					  "oneOf":[
						{"$ref":"#/components/schemas/Openapi3TestO1"},
						{"$ref":"#/components/schemas/Openapi3TestO2"}
					  ]
					},
					"examples":{
					  "responseOne":{
						"summary":"First possible answer",
						"value":{"code":"1234567890123456789012"}
					  },
					  "responseTwo":{
						"summary":"Other possible answer",
						"value":{"code":"0000000000000000000000"}
					  }
					}
				  }
				}
			  }
			}
		  }
		}
	  },
	  "components":{
		"schemas":{
		  "Openapi3TestO1":{
			"type":"object",
			"properties":{"code":{"type":"string"},"f1":{"type":"integer"}}
		  },
		  "Openapi3TestO2":{
			"type":"object",
			"properties":{"code":{"type":"string"},"f2":{"type":"integer"}}
		  }
		}
	  }
	}`, r.SpecSchema())
}

func TestWithCustomize(t *testing.T) {
	r := openapi3.NewReflector()

	op, err := r.NewOperationContext(http.MethodPost, "/{document_id}/{client}")
	require.NoError(t, err)

	op.AddReqStructure(new(struct {
		DocumentID string `path:"document_id"`
		Client     string `path:"client"`
		Foo        int    `json:"foo"`
	}), openapi.WithCustomize(func(cor openapi.ContentOrReference) {
		_, ok := cor.(*openapi3.RequestBodyOrRef)
		assert.True(t, ok)

		cor.SetReference("../somewhere/components/requests/foo.yaml")
	}))

	op.AddRespStructure(
		nil, openapi.WithReference("../somewhere/components/responses/204.yaml"), openapi.WithHTTPStatus(204),
	)
	op.AddRespStructure(
		nil, openapi.WithCustomize(func(cor openapi.ContentOrReference) {
			_, ok := cor.(*openapi3.ResponseOrRef)
			assert.True(t, ok)

			cor.SetReference("../somewhere/components/responses/200.yaml")
		}), openapi.WithHTTPStatus(200),
	)

	require.NoError(t, r.AddOperation(op))

	assertjson.EqMarshal(t, `{
	  "openapi":"3.0.3","info":{"title":"","version":""},
	  "paths":{
		"/{document_id}/{client}":{
		  "post":{
			"parameters":[
			  {
				"name":"document_id","in":"path","required":true,
				"schema":{"type":"string"}
			  },
			  {
				"name":"client","in":"path","required":true,
				"schema":{"type":"string"}
			  }
			],
			"requestBody":{"$ref":"../somewhere/components/requests/foo.yaml"},
			"responses":{
			  "200":{"$ref":"../somewhere/components/responses/200.yaml"},
			  "204":{"$ref":"../somewhere/components/responses/204.yaml"}
			}
		  }
		}
	  }
	}`, r.SpecSchema())
}
