package openapi3_test

import (
	"net/http"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/swaggest/assertjson"
	jsonschema "github.com/swaggest/jsonschema-go"
	"github.com/swaggest/openapi-go"
	"github.com/swaggest/openapi-go/openapi3"
)

func TestReflector_WalkRequestJSONSchemas(t *testing.T) {
	r := openapi3.NewReflector()

	type Embed struct {
		Query1 int    `query:"query1" minimum:"3"`
		Query2 string `query:"query2" minLength:"2"`
		Query3 bool   `query:"query3" description:"Trivial schema."`
	}

	type DeeplyEmbedded struct {
		Embed
	}

	type req struct {
		*DeeplyEmbedded

		Path1 int    `path:"path1" minimum:"3"`
		Path2 string `path:"path2" minLength:"2"`
		Path3 bool   `path:"path3" description:"Trivial schema."`

		Header1 int    `header:"header1" minimum:"3"`
		Header2 string `header:"header2" minLength:"2"`
		Header3 bool   `header:"header3" description:"Trivial schema."`

		Cookie1 int    `cookie:"cookie1" minimum:"3"`
		Cookie2 string `cookie:"cookie2" minLength:"2"`
		Cookie3 bool   `cookie:"cookie3" description:"Trivial schema."`

		Form1 int    `form:"form1" minimum:"3"`
		Form2 string `form:"form2" minLength:"2"`
		Form3 bool   `form:"form3" description:"Trivial schema."`
	}

	cu := openapi.ContentUnit{}
	cu.Structure = req{}

	schemas := map[string]*jsonschema.SchemaOrBool{}
	doneCalled := 0

	require.NoError(t, r.WalkRequestJSONSchemas(http.MethodPost, cu,
		func(in openapi.In, paramName string, schema *jsonschema.SchemaOrBool, required bool) error {
			schemas[string(in)+"-"+paramName+"-"+strconv.FormatBool(required)+"-"+
				strconv.FormatBool(schema.IsTrivial(r.ResolveJSONSchemaRef))] = schema

			return nil
		},
		func(_ openapi.OperationContext) {
			doneCalled++
		},
	))

	assert.Equal(t, 1, doneCalled)
	assertjson.EqMarshal(t, `{
	  "cookie-cookie1-false-false":{"minimum":3,"type":"integer"},
	  "cookie-cookie2-false-false":{"minLength":2,"type":"string"},
	  "cookie-cookie3-false-true":{"description":"Trivial schema.","type":"boolean"},
	  "formData-form1-false-false":{"minimum":3,"type":"integer"},
	  "formData-form2-false-false":{"minLength":2,"type":"string"},
	  "formData-form3-false-true":{"description":"Trivial schema.","type":"boolean"},
	  "header-header1-false-false":{"minimum":3,"type":"integer"},
	  "header-header2-false-false":{"minLength":2,"type":"string"},
	  "header-header3-false-true":{"description":"Trivial schema.","type":"boolean"},
	  "path-path1-true-false":{"minimum":3,"type":"integer"},
	  "path-path2-true-false":{"minLength":2,"type":"string"},
	  "path-path3-true-true":{"description":"Trivial schema.","type":"boolean"},
	  "query-form1-false-false":{"minimum":3,"type":"integer"},
	  "query-form2-false-false":{"minLength":2,"type":"string"},
	  "query-form3-false-true":{"description":"Trivial schema.","type":"boolean"},
	  "query-query1-false-false":{"minimum":3,"type":"integer"},
	  "query-query2-false-false":{"minLength":2,"type":"string"},
	  "query-query3-false-true":{"description":"Trivial schema.","type":"boolean"}
	}`, schemas)

	assertjson.EqMarshal(t, `{
	  "openapi":"3.0.3","info":{"title":"","version":""},"paths":{},
	  "components":{
		"schemas":{
		  "FormDataOpenapi3TestReq":{
			"type":"object",
			"properties":{
			  "form1":{"minimum":3,"type":"integer"},
			  "form2":{"minLength":2,"type":"string"},
			  "form3":{"type":"boolean","description":"Trivial schema."}
			}
		  }
		}
	  }
	}`, r.Spec)
}

func TestReflector_WalkRequestJSONSchemas_jsonBody(t *testing.T) {
	r := openapi3.NewReflector()

	type Embed struct {
		Query1 int    `query:"query1" minimum:"3"`
		Query2 string `query:"query2" minLength:"2"`
		Query3 bool   `query:"query3" description:"Trivial schema."`

		Foo int `json:"foo" minimum:"6"`
	}

	type DeeplyEmbedded struct {
		Embed
	}

	type req struct {
		*DeeplyEmbedded

		Path1 int    `path:"path1" minimum:"3"`
		Path2 string `path:"path2" minLength:"2"`
		Path3 bool   `path:"path3" description:"Trivial schema."`

		Header1 int    `header:"header1" minimum:"3"`
		Header2 string `header:"header2" minLength:"2"`
		Header3 bool   `header:"header3" description:"Trivial schema."`

		Cookie1 int    `cookie:"cookie1" minimum:"3"`
		Cookie2 string `cookie:"cookie2" minLength:"2"`
		Cookie3 bool   `cookie:"cookie3" description:"Trivial schema."`

		Bar []string `json:"bar" minItems:"15"`
	}

	cu := openapi.ContentUnit{}
	cu.Structure = req{}

	schemas := map[string]*jsonschema.SchemaOrBool{}
	doneCalled := 0

	require.NoError(t, r.WalkRequestJSONSchemas(http.MethodPost, cu,
		func(in openapi.In, paramName string, schema *jsonschema.SchemaOrBool, required bool) error {
			schemas[string(in)+"-"+paramName+"-"+strconv.FormatBool(required)+"-"+
				strconv.FormatBool(schema.IsTrivial(r.ResolveJSONSchemaRef))] = schema

			return nil
		},
		func(_ openapi.OperationContext) {
			doneCalled++
		},
	))

	assert.Equal(t, 1, doneCalled)
	assertjson.EqMarshal(t, `{
	  "body-body-false-false":{
		"properties":{
		  "bar":{"items":{"type":"string"},"minItems":15,"type":["array","null"]},
		  "foo":{"minimum":6,"type":"integer"}
		},
		"type":"object"
	  },
	  "cookie-cookie1-false-false":{"minimum":3,"type":"integer"},
	  "cookie-cookie2-false-false":{"minLength":2,"type":"string"},
	  "cookie-cookie3-false-true":{"description":"Trivial schema.","type":"boolean"},
	  "header-header1-false-false":{"minimum":3,"type":"integer"},
	  "header-header2-false-false":{"minLength":2,"type":"string"},
	  "header-header3-false-true":{"description":"Trivial schema.","type":"boolean"},
	  "path-path1-true-false":{"minimum":3,"type":"integer"},
	  "path-path2-true-false":{"minLength":2,"type":"string"},
	  "path-path3-true-true":{"description":"Trivial schema.","type":"boolean"},
	  "query-query1-false-false":{"minimum":3,"type":"integer"},
	  "query-query2-false-false":{"minLength":2,"type":"string"},
	  "query-query3-false-true":{"description":"Trivial schema.","type":"boolean"}
	}`, schemas)

	assertjson.EqMarshal(t, `{
	  "openapi":"3.0.3","info":{"title":"","version":""},"paths":{},
	  "components":{
		"schemas":{
		  "Openapi3TestReq":{
			"type":"object",
			"properties":{
			  "bar":{
				"minItems":15,"type":"array","items":{"type":"string"},
				"nullable":true
			  },
			  "foo":{"minimum":6,"type":"integer"}
			}
		  }
		}
	  }
	}`, r.Spec)
}

func TestReflector_WalkResponseJSONSchemas(t *testing.T) {
	r := openapi3.NewReflector()

	type Embed struct {
		Header1 int    `header:"header1" minimum:"3"`
		Header2 string `header:"header2" minLength:"2"`
		Header3 bool   `header:"header3" description:"Trivial schema."`

		Foo int `json:"foo" minimum:"6"`
	}

	type DeeplyEmbedded struct {
		Embed
	}

	type req struct {
		*DeeplyEmbedded

		Header4 int    `header:"header4" minimum:"3"`
		Header5 string `header:"header5" minLength:"2"`
		Header6 bool   `header:"header6" description:"Trivial schema."`

		Bar []string `json:"bar" minItems:"15"`
	}

	cu := openapi.ContentUnit{}
	cu.Structure = req{}

	schemas := map[string]*jsonschema.SchemaOrBool{}
	doneCalled := 0

	require.NoError(t, r.WalkResponseJSONSchemas(cu,
		func(in openapi.In, paramName string, schema *jsonschema.SchemaOrBool, required bool) error {
			schemas[string(in)+"-"+paramName+"-"+strconv.FormatBool(required)+"-"+
				strconv.FormatBool(schema.IsTrivial(r.ResolveJSONSchemaRef))] = schema

			return nil
		},
		func(_ openapi.OperationContext) {
			doneCalled++
		},
	))

	assert.Equal(t, 1, doneCalled)
	assertjson.EqMarshal(t, `{
	  "body-body-false-false":{
		"properties":{
		  "bar":{"items":{"type":"string"},"minItems":15,"type":["array","null"]},
		  "foo":{"minimum":6,"type":"integer"}
		},
		"type":"object"
	  },
	  "header-Header1-false-false":{"minimum":3,"type":"integer"},
	  "header-Header2-false-false":{"minLength":2,"type":"string"},
	  "header-Header3-false-true":{"description":"Trivial schema.","type":"boolean"},
	  "header-Header4-false-false":{"minimum":3,"type":"integer"},
	  "header-Header5-false-false":{"minLength":2,"type":"string"},
	  "header-Header6-false-true":{"description":"Trivial schema.","type":"boolean"}
	}`, schemas)

	assertjson.EqMarshal(t, `{
	  "openapi":"3.0.3","info":{"title":"","version":""},"paths":{},
	  "components":{
		"schemas":{
		  "Openapi3TestReq":{
			"type":"object",
			"properties":{
			  "bar":{
				"minItems":15,"type":"array","items":{"type":"string"},
				"nullable":true
			  },
			  "foo":{"minimum":6,"type":"integer"}
			}
		  }
		}
	  }
	}`, r.Spec)
}
