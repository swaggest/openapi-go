package openapi31_test

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
	"github.com/swaggest/openapi-go/openapi31"
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
	reflector := openapi31.Reflector{}

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

	reflector := openapi31.Reflector{}
	oc, err := reflector.NewOperationContext(http.MethodPost, "/somewhere")
	require.NoError(t, err)

	oc.AddReqStructure(new(req))

	require.NoError(t, reflector.AddOperation(oc))

	assertjson.EqMarshal(t, `{
	  "openapi":"3.1.0","info":{"title":"","version":""},
	  "paths":{
		"/somewhere":{
		  "post":{
			"requestBody":{
			  "content":{
				"multipart/form-data":{
				  "schema":{"$ref":"#/components/schemas/FormDataOpenapi31TestReq"}
				}
			  }
			},
			"responses":{"204":{"description":"No Content"}}
		  }
		}
	  },
	  "components":{
		"schemas":{
		  "FormDataOpenapi31TestReq":{
			"properties":{"upload1":{"$ref":"#/components/schemas/MultipartFile"}},
			"type":"object"
		  },
		  "MultipartFile":{"format":"binary","type":"string","contentMediaType": "application/octet-stream"}
		}
	  }
	}`, reflector.Spec)
}

func TestReflector_AddOperation_request(t *testing.T) {
	reflector := openapi31.Reflector{}

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
	reflector := openapi31.Reflector{}

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

	pathItem := openapi31.PathItem{}
	pathItem.
		WithSummary("Path Summary").
		WithDescription("Path Description")
	s.Paths.WithMapOfPathItemValuesItem("/somewhere/{in_path}", pathItem)

	require.NoError(t, reflector.AddOperation(oc))

	require.NoError(t, s.SetupOperation(http.MethodPost, "/somewhere/{in_path}", func(op *openapi31.Operation) error {
		js := openapi31.ToJSONSchema(op.RequestBody.RequestBody.Content["multipart/form-data"].Schema, s)
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

	require.NoError(t, s.SetupOperation(http.MethodGet, "/somewhere/{in_path}", func(op *openapi31.Operation) error {
		js := openapi31.ToJSONSchema(op.Responses.MapOfResponseOrReferenceValues[strconv.Itoa(http.StatusOK)].Response.Content["application/json"].
			Schema, s)
		jsb, err := assertjson.MarshalIndentCompact(js, "", " ", 120)
		require.NoError(t, err)

		require.NoError(t, os.WriteFile("testdata/resp_schema_last_run.json", jsb, 0o600))

		expected, err := os.ReadFile("testdata/resp_schema.json")
		require.NoError(t, err)
		assertjson.EqualMarshal(t, expected, js)

		js = openapi31.ToJSONSchema(op.Responses.MapOfResponseOrReferenceValues[strconv.Itoa(http.StatusOK)].Response.Headers["X-Header-Field"].Header.
			Schema, s)
		assertjson.EqMarshal(t, `{"type": "string", "description": "Sample header response."}`, js)

		require.NoError(t, err)
		assertjson.EqMarshal(t, `{"type": "integer", "description": "Query parameter."}`,
			op.Parameters[0].Parameter.Schema)

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
	reflector := openapi31.Reflector{}

	oc, err := reflector.NewOperationContext(http.MethodPost, "/somewhere/{id}")
	require.NoError(t, err)

	oc.AddReqStructure(new(PathParamAndBody))

	require.NoError(t, reflector.AddOperation(oc))

	s := reflector.SpecEns()
	s.Info.Title = apiName
	s.Info.Version = apiVersion

	assertjson.EqMarshal(t, `{
	  "openapi":"3.1.0","info":{"title":"SampleAPI","version":"1.2.3"},
	  "paths":{
		"/somewhere/{id}":{
		  "post":{
			"parameters":[{"name":"id","in":"path","required":true,"schema":{"type":"string"}}],
			"requestBody":{
			  "content":{
				"application/json":{
				  "schema":{"$ref":"#/components/schemas/Openapi31TestPathParamAndBody"}
				}
			  }
			},
			"responses":{"204":{"description":"No Content"}}
		  }
		}
	  },
	  "components":{
		"schemas":{
		  "Openapi31TestPathParamAndBody":{"items":{"type":"string"},"type":["null","array"]}
		}
	  }
	}`, s)
}

type WithReqBody PathParamAndBody

func (*WithReqBody) ForceRequestBody() {}

func TestReflector_AddOperation_RequestBodyEnforcer(t *testing.T) {
	reflector := openapi31.Reflector{}
	oc, err := reflector.NewOperationContext(http.MethodGet, "/somewhere/{id}")
	require.NoError(t, err)

	oc.AddReqStructure(new(WithReqBody))

	s := reflector.SpecEns()
	s.Info.Title = apiName
	s.Info.Version = apiVersion

	require.NoError(t, reflector.AddOperation(oc))

	assertjson.EqMarshal(t, `{
	  "openapi":"3.1.0","info":{"title":"SampleAPI","version":"1.2.3"},
	  "paths":{
		"/somewhere/{id}":{
		  "get":{
			"parameters":[{"name":"id","in":"path","required":true,"schema":{"type":"string"}}],
			"requestBody":{
			  "content":{
				"application/json":{
				  "schema":{"$ref":"#/components/schemas/Openapi31TestWithReqBody"}
				}
			  }
			},
			"responses":{"204":{"description":"No Content"}}
		  }
		}
	  },
	  "components":{
		"schemas":{
		  "Openapi31TestWithReqBody":{"items":{"type":"string"},"type":["null","array"]}
		}
	  }
	}`, s)
}

func TestReflector_AddOperation_response(t *testing.T) {
	reflector := openapi31.Reflector{}
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
	 "openapi":"3.1.0","info":{"title":"SampleAPI","version":"1.2.3"},
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
	reflector := openapi31.Reflector{}
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
	  "openapi":"3.1.0","info":{"title":"SampleAPI","version":"1.2.3"},
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
					"properties":{
					  "upload6":{"$ref":"#/components/schemas/MultipartFile"},
					  "value3":{"type":"number"}
					},
					"type":"object"
				  }
				}
			  }
			},
			"responses":{"204":{"description":"No Content"}}
		  }
		}
	  },
	  "components":{"schemas":{"MultipartFile":{"format":"binary","type":"string", "contentMediaType": "application/octet-stream"}}}
	}`, s)
}

func TestReflector_AddOperation_request_queryObject(t *testing.T) {
	reflector := openapi31.Reflector{}

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
	  "openapi":"3.1.0","info":{"title":"SampleAPI","version":"1.2.3"},
	  "paths":{
		"/somewhere":{
		  "post":{
			"parameters":[
			  {
				"name":"in_query","in":"query",
				"schema":{
				  "additionalProperties":{"type":"number"},"type":["object","null"]
				},
				"style":"deepObject","explode":true
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
	reflector := openapi31.Reflector{}
	oc, err := reflector.NewOperationContext(http.MethodPost, "/somewhere")
	require.NoError(t, err)

	s := reflector.SpecEns()
	s.Info.Title = apiName
	s.Info.Version = apiVersion

	oc.AddReqStructure(new(namedType))
	require.NoError(t, reflector.AddOperation(oc))

	assertjson.EqMarshal(t, `{
	  "openapi":"3.1.0","info":{"title":"SampleAPI","version":"1.2.3"},
	  "paths":{
		"/somewhere":{
		  "post":{
			"parameters":[
			  {
				"name":"in_query","in":"query",
				"schema":{"$ref":"#/components/schemas/Openapi31TestLabels"},
				"style":"deepObject","explode":true
			  }
			],
			"responses":{"204":{"description":"No Content"}}
		  }
		}
	  },
	  "components":{
		"schemas":{
		  "Openapi31TestLabels":{"additionalProperties":{"type":"number"},"type":"object"}
		}
	  }
	}`, s)

	js, found := reflector.ResolveJSONSchemaRef("#/components/schemas/Openapi31TestLabels")
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

	r := openapi31.Reflector{}
	oc, err := r.NewOperationContext(http.MethodGet, "/")
	require.NoError(t, err)

	oc.AddReqStructure(new(req))

	require.NoError(t, r.AddOperation(oc))

	assertjson.EqMarshal(t, `{
	  "openapi":"3.1.0","info":{"title":"","version":""},
	  "paths":{
		"/":{
		  "get":{
			"parameters":[
			  {
				"name":"one","in":"query",
				"content":{
				  "application/json":{"schema":{"$ref":"#/components/schemas/Openapi31TestFilter"}}
				}
			  },
			  {
				"name":"two","in":"query",
				"content":{
				  "application/json":{"schema":{"$ref":"#/components/schemas/Openapi31TestFilter"}}
				}
			  },
			  {
				"name":"three","in":"query",
				"schema":{
				  "additionalProperties":{"type":"integer"},
				  "type":["object","null"]
				},
				"style":"deepObject","explode":true
			  },
			  {
				"name":"four","in":"query",
				"content":{
				  "application/json":{
					"schema":{
					  "additionalProperties":{"type":"integer"},
					  "type":["null","object"]
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
		  "Openapi31TestFilter":{
			"properties":{
			  "labels":{"items":{"type":"string"},"type":"array"},
			  "type":{"type":"string"}
			},
			"type":"object"
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

	r := openapi31.Reflector{}
	oc, err := r.NewOperationContext(http.MethodGet, "/{path}")
	require.NoError(t, err)

	oc.AddReqStructure(new(req))

	require.NoError(t, r.AddOperation(oc))

	assertjson.EqMarshal(t, `{
	  "openapi":"3.1.0","info":{"title":"","version":""},
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

	o3, ok := oc.(openapi31.OperationExposer)
	require.True(t, ok)

	assert.True(t, o3.Operation().UnknownParamIsForbidden(openapi31.ParameterInCookie))
	assert.False(t, o3.Operation().UnknownParamIsForbidden(openapi31.ParameterInHeader))
	assert.True(t, o3.Operation().UnknownParamIsForbidden(openapi31.ParameterInQuery))
	assert.False(t, o3.Operation().UnknownParamIsForbidden(openapi31.ParameterInPath))
}

func TestReflector_AddOperation_request_noBody(t *testing.T) {
	type req struct {
		ID int `json:"id" path:"id"`
	}

	r := openapi31.Reflector{}

	for _, method := range []string{http.MethodHead, http.MethodGet, http.MethodDelete, http.MethodTrace} {
		oc, err := r.NewOperationContext(method, "/{id}")
		require.NoError(t, err)

		oc.AddReqStructure(new(req))

		require.NoError(t, r.AddOperation(oc))

		require.NoError(t, r.SpecEns().SetupOperation(method, "/{id}", func(op *openapi31.Operation) error {
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

		require.NoError(t, r.SpecEns().SetupOperation(method, "/{id}", func(op *openapi31.Operation) error {
			assertjson.EqMarshal(t, `{
			  "parameters":[{"name":"id","in":"path","required":true,"schema":{"type":"integer"}}],
			  "requestBody":{
				"content":{
				  "application/json":{"schema":{"$ref":"#/components/schemas/Openapi31TestReq"}}
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

	r := openapi31.Reflector{}
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

	r := openapi31.Reflector{}
	oc, err := r.NewOperationContext(http.MethodPost, "/foo")
	require.NoError(t, err)

	oc.AddReqStructure(new(req))
	require.NoError(t, r.AddOperation(oc))

	assertjson.EqMarshal(t, `{
	  "openapi":"3.1.0","info":{"title":"","version":""},
	  "paths":{
		"/foo":{
		  "post":{
			"requestBody":{
			  "content":{
				"application/x-www-form-urlencoded":{"schema":{"$ref":"#/components/schemas/FormDataOpenapi31TestReq"}}
			  }
			},
			"responses":{"204":{"description":"No Content"}}
		  }
		}
	  },
	  "components":{
		"schemas":{
		  "FormDataOpenapi31TestReq":{"type":"object","properties":{"foo":{"type":"integer"}}}
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

	r := openapi31.Reflector{}
	oc, err := r.NewOperationContext(http.MethodPost, "/foo")
	require.NoError(t, err)

	oc.AddReqStructure(req{})

	require.NoError(t, r.AddOperation(oc))

	assertjson.EqMarshal(t, `{
	  "openapi":"3.1.0","info":{"title":"","version":""},
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
				"application/x-www-form-urlencoded":{"schema":{"$ref":"#/components/schemas/FormDataOpenapi31TestReq"}}
			  }
			},
			"responses":{"204":{"description":"No Content"}}
		  }
		}
	  },
	  "components":{
		"schemas":{
		  "FormDataOpenapi31TestReq":{
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

	r := openapi31.Reflector{}
	oc, err := r.NewOperationContext(http.MethodPost, "/foo")
	require.NoError(t, err)

	oc.AddReqStructure(req{})

	require.NoError(t, r.AddOperation(oc))

	assertjson.EqMarshal(t, `{
	  "openapi":"3.1.0","info":{"title":"","version":""},
	  "paths":{
		"/foo":{
		  "post":{
			"parameters":[
			  {"name":"bar","in":"query","schema":{"type":"integer"}},
			  {"name":"quux","in":"query","schema":{"type":"string"}}
			],
			"requestBody":{
			  "content":{
				"application/x-www-form-urlencoded":{"schema":{"$ref":"#/components/schemas/FormDataOpenapi31TestReq"}}
			  }
			},
			"responses":{"204":{"description":"No Content"}}
		  }
		}
	  },
	  "components":{
		"schemas":{
		  "FormDataOpenapi31TestReq":{
			"type":"object",
			"properties":{"bar":{"type":"integer"},"quux":{"type":"string"}}
		  }
		}
	  }
	}`, r.SpecEns())
}

func TestReflector_AddOperation_request_queryObject_deepObject(t *testing.T) {
	reflector := openapi31.Reflector{}

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
	  "openapi":"3.1.0","info":{"title":"","version":""},
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
					"schema":{"$ref":"#/components/schemas/Openapi31TestJsonFilter"}
				  }
				}
			  },
			  {
				"name":"deep_object_filter","in":"query",
				"schema":{"$ref":"#/components/schemas/Openapi31TestDeepObjectFilter"},
				"style":"deepObject","explode":true
			  },
			  {
				"name":"id","in":"path","required":true,
				"schema":{"examples":["XXX-XXXXX"],"type":"string"}
			  }
			],
			"responses":{"204":{"description":"No Content"}}
		  }
		}
	  },
	  "components":{
		"schemas":{
		  "Openapi31TestDeepObjectFilter":{
			"properties":{
			  "baz":{"type":"boolean"},
			  "deeper":{"properties":{"val":{"type":"string"}},"type":"object"},
			  "quux":{"type":"number"}
			},
			"type":"object"
		  },
		  "Openapi31TestJsonFilter":{
			"properties":{
			  "bar":{"type":"integer"},
			  "deeper":{"properties":{"val":{"type":"string"}},"type":"object"},
			  "foo":{"type":"string"}
			},
			"type":"object"
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
	r := openapi31.NewReflector()
	oc, err := r.NewOperationContext(http.MethodPost, "/foo")
	require.NoError(t, err)

	oc.AddReqStructure(textCSV{})
	oc.AddRespStructure(textCSV{})

	require.NoError(t, r.AddOperation(oc))

	assertjson.EqMarshal(t, `{
	  "openapi":"3.1.0","info":{"title":"","version":""},
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
	r := openapi31.NewReflector()
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
	  "openapi":"3.1.0","info":{"title":"","version":""},
	  "paths":{
		"/foo":{
		  "post":{
			"parameters":[{"name":"bar","in":"header","schema":{"type":"number"}}],
			"requestBody":{
			  "content":{
				"application/json":{"schema":{"$ref":"#/components/schemas/Openapi31TestReqJSON"}},
				"application/x-www-form-urlencoded":{
				  "schema":{"$ref":"#/components/schemas/FormDataOpenapi31TestReqForm"}
				}
			  }
			},
			"responses":{"204":{"description":"No Content"}}
		  }
		}
	  },
	  "components":{
		"schemas":{
		  "FormDataOpenapi31TestReqForm":{
			"properties":{
			  "foo":{"type":"integer"},"simple":{"type":"string"},
			  "special":{"$ref":"#/components/schemas/Openapi31TestSpecialString"}
			},
			"type":"object"
		  },
		  "Openapi31TestReqJSON":{
			"properties":{
			  "foo":{"type":"integer"},"simple":{"type":"string"},
			  "special":{"$ref":"#/components/schemas/Openapi31TestSpecialString"}
			},
			"type":"object"
		  },
		  "Openapi31TestSpecialString":{
			"description":"Very special.","examples":["xy5abcd4sq9s"],
			"maxLength":12,"minLength":12,"type":"string"
		  }
		}
	  }
	}`, r.Spec)
}

func Test_Repro2(t *testing.T) {
	oarefl := openapi31.NewReflector()
	oarefl.JSONSchemaReflector().DefaultOptions = append(oarefl.JSONSchemaReflector().DefaultOptions, jsonschema.ProcessWithoutTags)

	{
		var dummyIn struct {
			ID int
		}

		var dummyOut struct {
			Done bool
		}

		op, err := oarefl.NewOperationContext(http.MethodPost, "/postDelete")
		if err != nil {
			t.Fatal(err)
		}

		op.AddReqStructure(dummyIn, openapi.WithContentType("application/json"))
		op.AddRespStructure(dummyOut, openapi.WithHTTPStatus(200))

		if err = oarefl.AddOperation(op); err != nil {
			t.Fatal(err)
		}
	}

	assertjson.EqMarshal(t, `{
	  "openapi":"3.1.0","info":{"title":"","version":""},
	  "paths":{
		"/postDelete":{
		  "post":{
			"requestBody":{
			  "content":{
				"application/json":{
				  "schema":{"properties":{"ID":{"type":"integer"}},"type":"object"}
				}
			  }
			},
			"responses":{
			  "200":{
				"description":"OK",
				"content":{
				  "application/json":{
					"schema":{"properties":{"Done":{"type":"boolean"}},"type":"object"}
				  }
				}
			  }
			}
		  }
		}
	  }
	}`, oarefl.SpecSchema())
}
