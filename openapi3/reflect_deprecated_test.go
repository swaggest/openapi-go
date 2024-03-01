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
	"github.com/swaggest/openapi-go/openapi3"
)

func TestReflector_SetRequest_array(t *testing.T) {
	reflector := openapi3.Reflector{}

	s := reflector.SpecEns()
	s.Info.Title = apiName
	s.Info.Version = apiVersion

	op := openapi3.Operation{}

	err := reflector.SetRequest(&op, new([]GetReq), http.MethodPost)
	require.NoError(t, err)

	require.NoError(t, s.AddOperation(http.MethodPost, "/somewhere", op))

	b, err := assertjson.MarshalIndentCompact(s, "", " ", 120)
	require.NoError(t, err)

	require.NoError(t, os.WriteFile("testdata/openapi_req_array_last_run.json", b, 0o600))

	expected, err := os.ReadFile("testdata/openapi_req_array.json")
	require.NoError(t, err)

	assertjson.Equal(t, expected, b)
}

func TestReflector_SetRequest_uploadInterface(t *testing.T) {
	type req struct {
		File1 multipart.File `formData:"upload1"`
	}

	reflector := openapi3.Reflector{}
	s := reflector.SpecEns()
	op := openapi3.Operation{}

	err := reflector.SetRequest(&op, new(req), http.MethodPost)
	require.NoError(t, err)

	require.NoError(t, s.AddOperation(http.MethodPost, "/somewhere", op))

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
	}`, s)
}

func TestReflector_SetRequest(t *testing.T) {
	reflector := openapi3.Reflector{}

	s := reflector.SpecEns()
	s.SetTitle(apiName)
	s.SetVersion(apiVersion)
	s.SetDescription("This a sample API description.")

	op := openapi3.Operation{}

	err := reflector.SetRequest(&op, new(GetReq), http.MethodGet)
	require.NoError(t, err)

	require.NoError(t, s.AddOperation(http.MethodGet, "/somewhere/{in_path}", op))

	b, err := assertjson.MarshalIndentCompact(s, "", " ", 120)
	require.NoError(t, err)

	require.NoError(t, os.WriteFile("testdata/openapi_req2_last_run.json", b, 0o600))

	expected, err := os.ReadFile("testdata/openapi_req2.json")
	require.NoError(t, err)

	assertjson.Equal(t, expected, b)
}

func TestReflector_SetJSONResponse(t *testing.T) {
	reflector := openapi3.Reflector{}

	s := reflector.SpecEns()
	s.Info.Title = apiName
	s.Info.Version = apiVersion

	reflector.AddTypeMapping(new(WeirdResp), new(Resp))

	op := openapi3.Operation{}

	require.NoError(t, reflector.SetRequest(&op, new(Req), http.MethodPost))
	require.NoError(t, reflector.SetJSONResponse(&op, new(WeirdResp), http.StatusOK))
	require.NoError(t, reflector.SetJSONResponse(&op, new([]WeirdResp), http.StatusConflict))
	require.NoError(t, reflector.SetStringResponse(&op, http.StatusConflict, "text/html"))

	pathItem := openapi3.PathItem{}
	pathItem.
		WithSummary("Path Summary").
		WithDescription("Path Description")
	s.Paths.WithMapOfPathItemValuesItem("/somewhere/{in_path}", pathItem)

	js := op.RequestBody.RequestBody.Content["multipart/form-data"].Schema.ToJSONSchema(s)
	expected, err := os.ReadFile("testdata/req_schema.json")
	require.NoError(t, err)
	assertjson.EqualMarshal(t, expected, js)

	require.NoError(t, s.AddOperation(http.MethodPost, "/somewhere/{in_path}", op))

	op = openapi3.Operation{}

	require.NoError(t, reflector.SetRequest(&op, new(GetReq), http.MethodGet))
	require.NoError(t, reflector.SetJSONResponse(&op, new(Resp), http.StatusOK))
	require.NoError(t, s.AddOperation(http.MethodGet, "/somewhere/{in_path}", op))

	js = op.Responses.MapOfResponseOrRefValues[strconv.Itoa(http.StatusOK)].Response.Content["application/json"].
		Schema.ToJSONSchema(s)
	jsb, err := assertjson.MarshalIndentCompact(js, "", " ", 120)
	require.NoError(t, err)

	require.NoError(t, os.WriteFile("testdata/resp_schema_last_run.json", jsb, 0o600))

	expected, err = os.ReadFile("testdata/resp_schema.json")
	require.NoError(t, err)
	assertjson.EqualMarshal(t, expected, js)

	js = op.Responses.MapOfResponseOrRefValues[strconv.Itoa(http.StatusOK)].Response.Headers["X-Header-Field"].Header.
		Schema.ToJSONSchema(s)
	assertjson.EqMarshal(t, `{"type": "string", "description": "Sample header response."}`, js)

	require.NoError(t, err)
	assertjson.EqMarshal(t, `{"type": "integer", "description": "Query parameter."}`,
		op.Parameters[0].Parameter.Schema.ToJSONSchema(s))

	b, err := assertjson.MarshalIndentCompact(s, "", " ", 120)
	require.NoError(t, err)

	require.NoError(t, os.WriteFile("testdata/openapi_last_run.json", b, 0o600))

	expected, err = os.ReadFile("testdata/openapi.json")
	require.NoError(t, err)

	assertjson.EqualMarshal(t, expected, s)
}

func TestReflector_SetRequest_pathParamAndBody(t *testing.T) {
	reflector := openapi3.Reflector{}
	op := openapi3.Operation{}

	err := reflector.SetRequest(&op, new(PathParamAndBody), http.MethodPost)
	require.NoError(t, err)

	s := reflector.SpecEns()
	s.Info.Title = apiName
	s.Info.Version = apiVersion
	require.NoError(t, s.AddOperation(http.MethodPost, "/somewhere/{id}", op))

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

func TestRequestBodyEnforcer(t *testing.T) {
	reflector := openapi3.Reflector{}
	op := openapi3.Operation{}

	err := reflector.SetRequest(&op, new(WithReqBody), http.MethodGet)
	require.NoError(t, err)

	s := reflector.SpecEns()
	s.Info.Title = apiName
	s.Info.Version = apiVersion
	require.NoError(t, s.AddOperation(http.MethodGet, "/somewhere/{id}", op))

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

func TestReflector_SetupResponse(t *testing.T) {
	reflector := openapi3.Reflector{}
	op := openapi3.Operation{}

	s := reflector.SpecEns()
	s.Info.Title = apiName
	s.Info.Version = apiVersion

	require.NoError(t, reflector.SetupResponse(openapi3.OperationContext{
		Operation:       &op,
		RespContentType: "text/csv; charset=utf-8",
		HTTPStatus:      http.StatusNoContent,
		Output: new(struct {
			Val1 int
			Val2 string
		}),
		RespHeaderMapping: map[string]string{
			"Val1": "X-Value-1",
			"Val2": "X-Value-2",
		},
	}))
	require.NoError(t, s.AddOperation(http.MethodGet, "/somewhere", op))

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

func TestReflector_SetupRequest(t *testing.T) {
	reflector := openapi3.Reflector{}
	op := openapi3.Operation{}

	s := reflector.SpecEns()
	s.Info.Title = apiName
	s.Info.Version = apiVersion

	require.NoError(t, reflector.SetupRequest(openapi3.OperationContext{
		Operation:  &op,
		HTTPMethod: http.MethodPost,
		Input: new(struct {
			Val1 int
			Val2 string
			Val3 float64
			Val4 bool
			Val5 string
			Val6 multipart.File
		}),
		ReqHeaderMapping: map[string]string{
			"Val1": "X-Value-1",
		},
		ReqQueryMapping: map[string]string{
			"Val2": "value_2",
		},
		ReqFormDataMapping: map[string]string{
			"Val3": "value3",
			"Val6": "upload6",
		},
		ReqPathMapping: map[string]string{
			"Val4": "value-4",
		},
		ReqCookieMapping: map[string]string{
			"Val5": "value_5",
		},
	}))
	require.NoError(t, s.AddOperation(http.MethodPost, "/somewhere/{value-4}", op))

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

func TestReflector_SetupRequest_queryObject(t *testing.T) {
	reflector := openapi3.Reflector{}
	op := openapi3.Operation{}

	s := reflector.SpecEns()
	s.Info.Title = apiName
	s.Info.Version = apiVersion

	require.NoError(t, reflector.SetupRequest(openapi3.OperationContext{
		Operation:  &op,
		HTTPMethod: http.MethodGet,
		Input: new(struct {
			InQuery map[int]float64 `query:"in_query"`
		}),
	}))
	require.NoError(t, s.AddOperation(http.MethodPost, "/somewhere", op))

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

func TestReflector_SetupRequest_queryNamedObject(t *testing.T) {
	reflector := openapi3.Reflector{}
	op := openapi3.Operation{}

	s := reflector.SpecEns()
	s.Info.Title = apiName
	s.Info.Version = apiVersion

	require.NoError(t, reflector.SetupRequest(openapi3.OperationContext{
		Operation:  &op,
		HTTPMethod: http.MethodGet,
		Input:      new(namedType),
	}))
	require.NoError(t, s.AddOperation(http.MethodPost, "/somewhere", op))

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

func TestReflector_SetupRequest_jsonQuery(t *testing.T) {
	type filter struct {
		Labels []string `json:"labels,omitempty"`
		Type   string   `json:"type"`
	}

	type req struct {
		One   filter         `query:"one"`
		Two   filter         `query:"two"`
		Three map[string]int `query:"three"`
	}

	r := openapi3.Reflector{}
	oc := openapi3.OperationContext{
		Operation: &openapi3.Operation{},
		Input:     new(req),
	}

	require.NoError(t, r.SetupRequest(oc))
	require.NoError(t, r.SpecEns().AddOperation(http.MethodGet, "/", *oc.Operation))

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

func TestReflector_SetupRequest_forbidParams(t *testing.T) {
	type req struct {
		Query  string `query:"query"`
		Header string `header:"header"`
		Cookie string `cookie:"cookie"`
		Path   string `path:"path"`

		_ struct{} `query:"_" cookie:"_" additionalProperties:"false"`
	}

	r := openapi3.Reflector{}
	oc := openapi3.OperationContext{
		Operation: &openapi3.Operation{},
		Input:     new(req),
	}

	require.NoError(t, r.SetupRequest(oc))
	require.NoError(t, r.SpecEns().AddOperation(http.MethodGet, "/{path}", *oc.Operation))

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

	assert.True(t, oc.Operation.UnknownParamIsForbidden(openapi3.ParameterInCookie))
	assert.False(t, oc.Operation.UnknownParamIsForbidden(openapi3.ParameterInHeader))
	assert.True(t, oc.Operation.UnknownParamIsForbidden(openapi3.ParameterInQuery))
	assert.False(t, oc.Operation.UnknownParamIsForbidden(openapi3.ParameterInPath))
}

func TestReflector_SetupRequest_noBody(t *testing.T) {
	type req struct {
		ID int `json:"id" path:"id"`
	}

	r := openapi3.Reflector{}

	for _, method := range []string{http.MethodHead, http.MethodGet, http.MethodDelete, http.MethodTrace} {
		oc := openapi3.OperationContext{
			Operation:  &openapi3.Operation{},
			HTTPMethod: method,
			Input:      new(req),
		}

		require.NoError(t, r.SetupRequest(oc))
		assertjson.EqMarshal(t, `{
		  "parameters":[{"name":"id","in":"path","required":true,"schema":{"type":"integer"}}],
		  "responses":{}
		}`, oc.Operation)
	}

	for _, method := range []string{http.MethodPost, http.MethodPatch, http.MethodPut} {
		oc := openapi3.OperationContext{
			Operation:  &openapi3.Operation{},
			HTTPMethod: method,
			Input:      new(req),
		}

		require.NoError(t, r.SetupRequest(oc))
		assertjson.EqMarshal(t, `{
		  "parameters":[{"name":"id","in":"path","required":true,"schema":{"type":"integer"}}],
		  "requestBody":{
			"content":{
			  "application/json":{"schema":{"$ref":"#/components/schemas/Openapi3TestReq"}}
			}
		  },
		  "responses":{}
		}`, oc.Operation)
	}
}

func TestOperationCtx(t *testing.T) {
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
	oc := openapi3.OperationContext{
		Operation: &openapi3.Operation{},
		Input:     new(req),
		Output:    new(resp),
	}

	visited := map[string]bool{}

	var currentRC *jsonschema.ReflectContext

	r.DefaultOptions = append(r.DefaultOptions,
		func(rc *jsonschema.ReflectContext) {
			currentRC = rc
		},
		jsonschema.InterceptSchema(func(_ jsonschema.InterceptSchemaParams) (stop bool, err error) {
			if occ, ok := openapi3.OperationCtx(currentRC); ok {
				if occ.ProcessingResponse {
					visited["resp:"+occ.ProcessingIn] = true
				} else {
					visited["req:"+occ.ProcessingIn] = true
				}
			}

			return false, nil
		}),
	)

	require.NoError(t, r.SetupRequest(oc))
	require.NoError(t, r.SetupResponse(oc))

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

func TestReflector_SetRequest_formData_with_json(t *testing.T) {
	// In presence of `formData` tags, `json` tags would be ignored as request body.
	type req struct {
		Foo int `formData:"foo" json:"foo"`
	}

	r := openapi3.Reflector{}
	oc := openapi3.OperationContext{
		Operation: &openapi3.Operation{},
		Input:     new(req),
	}

	require.NoError(t, r.SetupRequest(oc))
	require.NoError(t, r.SpecEns().AddOperation(http.MethodGet, "/foo", *oc.Operation))

	assertjson.EqMarshal(t, `{
	  "openapi":"3.0.3","info":{"title":"","version":""},
	  "paths":{
		"/foo":{
		  "get":{
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

func TestReflector_SetupRequest_form(t *testing.T) {
	// Fields with `form` will be present as both `query` parameter and `formData` property.
	type req struct {
		Foo  int    `formData:"foo"`
		Bar  int    `form:"bar"`
		Baz  string `query:"baz"`
		Quux string `query:"quux" form:"ignored"`
	}

	r := openapi3.Reflector{}
	oc := openapi3.OperationContext{
		Operation: &openapi3.Operation{},
		Input:     req{},
	}

	require.NoError(t, r.SetupRequest(oc))
	require.NoError(t, r.SpecEns().AddOperation(http.MethodPost, "/foo", *oc.Operation))

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

func TestReflector_SetupRequest_form_only(t *testing.T) {
	// Fields with `form` will be present as both `query` parameter and `formData` property.
	type req struct {
		Bar  int    `form:"bar"`
		Quux string `form:"quux"`
	}

	r := openapi3.Reflector{}
	oc := openapi3.OperationContext{
		Operation: &openapi3.Operation{},
		Input:     req{},
	}

	require.NoError(t, r.SetupRequest(oc))
	require.NoError(t, r.SpecEns().AddOperation(http.MethodPost, "/foo", *oc.Operation))

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

func TestReflector_SetRequest_queryObject(t *testing.T) {
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

	getOp := openapi3.Operation{}

	handleError(reflector.SetRequest(&getOp, new(req), http.MethodGet))
	handleError(reflector.Spec.AddOperation(http.MethodGet, "/things/{id}", getOp))

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
