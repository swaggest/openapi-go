package openapi31_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/swaggest/assertjson"
	"github.com/swaggest/openapi-go/openapi31"
)

func TestSpec_SetOperation(t *testing.T) {
	s := openapi31.Spec{}
	op := openapi31.Operation{}

	op.WithParameters(
		openapi31.Parameter{In: openapi31.ParameterInPath, Name: "foo"}.ToParameterOrRef(),
	)

	assert.EqualError(t, s.AddOperation("bar", "/", op),
		"unexpected http method: bar")

	assert.EqualError(t, s.AddOperation(http.MethodGet, "/", op),
		"missing path parameter placeholder in url: foo")

	assert.EqualError(t, s.AddOperation(http.MethodGet, "/{bar}", op),
		"missing path parameter placeholder in url: foo, undefined path parameter: bar")

	assert.NoError(t, s.AddOperation(http.MethodGet, "/{foo}", op))

	assert.EqualError(t, s.AddOperation(http.MethodGet, "/{foo}", op),
		"operation already exists: get /{foo}")

	op.WithParameters(
		openapi31.Parameter{In: openapi31.ParameterInPath, Name: "foo"}.ToParameterOrRef(),
		openapi31.Parameter{In: openapi31.ParameterInPath, Name: "foo"}.ToParameterOrRef(),
		openapi31.Parameter{In: openapi31.ParameterInQuery, Name: "bar"}.ToParameterOrRef(),
		openapi31.Parameter{In: openapi31.ParameterInQuery, Name: "bar"}.ToParameterOrRef(),
	)

	assert.EqualError(t, s.AddOperation(http.MethodGet, "/another/{foo}", op),
		"duplicate parameter in path: foo, duplicate parameter in query: bar")
}

func TestSpec_SetupOperation_pathRegex(t *testing.T) {
	s := openapi31.Spec{}

	for _, tc := range []struct {
		path   string
		params []string
	}{
		{`/{month}-{day}-{year}`, []string{"month", "day", "year"}},
		{`/{month}/{day}/{year}`, []string{"month", "day", "year"}},
		{`/{month:[\d]+}-{day:[\d]+}-{year:[\d]+}`, []string{"month", "day", "year"}},
		{`/{articleSlug:[a-z-]+}`, []string{"articleSlug"}},
		{"/articles/{rid:^[0-9]{5,6}}", []string{"rid"}},
		{"/articles/{rid:^[0-9]{5,6}}/{zid:^[0-9]{5,6}}", []string{"rid", "zid"}},
		{"/articles/{zid:^0[0-9]+}", []string{"zid"}},
		{"/articles/{name:^@[a-z]+}/posts", []string{"name"}},
		{"/articles/{op:^[0-9]+}/run", []string{"op"}},
		{"/users/{userID:[^/]+}", []string{"userID"}},
		{"/users/{userID:[^/]+}/books/{bookID:.+}", []string{"userID", "bookID"}},
	} {
		t.Run(tc.path, func(t *testing.T) {
			assert.NoError(t, s.SetupOperation(http.MethodGet, tc.path,
				func(operation *openapi31.Operation) error {
					var pp []openapi31.ParameterOrReference

					for _, p := range tc.params {
						pp = append(pp, openapi31.Parameter{In: openapi31.ParameterInPath, Name: p}.ToParameterOrRef())
					}

					operation.WithParameters(pp...)

					return nil
				},
			))
		})
	}
}

func TestSpec_SetupOperation_uncleanPath(t *testing.T) {
	s := openapi31.Spec{}
	f := func(operation *openapi31.Operation) error {
		operation.WithParameters(openapi31.Parameter{In: openapi31.ParameterInPath, Name: "userID"}.ToParameterOrRef())

		return nil
	}

	assert.NoError(t, s.SetupOperation(http.MethodGet, "/users/{userID:[^/]+}", f))
	assert.NoError(t, s.SetupOperation(http.MethodPost, "/users/{userID:[^/]+}", f))

	assertjson.EqualMarshal(t, []byte(`{
	  "openapi":"","info":{"title":"","version":""},
	  "paths":{
		"/users/{userID}":{
		  "get":{"parameters":[{"name":"userID","in":"path"}]},
		  "post":{"parameters":[{"name":"userID","in":"path"}]}
		}
	  }
	}`), s)
}
