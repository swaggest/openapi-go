package openapi3_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/swaggest/openapi-go/openapi3"
)

func TestSpec_SetOperation(t *testing.T) {
	s := openapi3.Spec{}
	op := openapi3.Operation{}

	op.WithParameters(
		openapi3.Parameter{In: openapi3.ParameterInPath, Name: "foo"}.ToParameterOrRef(),
	)

	assert.EqualError(t, s.AddOperation("bar", "/", op),
		"unexpected http method: bar")

	assert.EqualError(t, s.AddOperation(http.MethodGet, "/", op),
		"missing path parameter placeholder in url: foo")

	assert.EqualError(t, s.AddOperation(http.MethodGet, "/{bar}", op),
		"missing path parameter placeholder in url: foo, undefined path parameter: bar")

	assert.NoError(t, s.AddOperation(http.MethodGet, "/{foo}", op))

	assert.EqualError(t, s.AddOperation(http.MethodGet, "/{foo}", op),
		"operation with method and path already exists")

	op.WithParameters(
		openapi3.Parameter{In: openapi3.ParameterInPath, Name: "foo"}.ToParameterOrRef(),
		openapi3.Parameter{In: openapi3.ParameterInPath, Name: "foo"}.ToParameterOrRef(),
		openapi3.Parameter{In: openapi3.ParameterInQuery, Name: "bar"}.ToParameterOrRef(),
		openapi3.Parameter{In: openapi3.ParameterInQuery, Name: "bar"}.ToParameterOrRef(),
	)

	assert.EqualError(t, s.AddOperation(http.MethodGet, "/another/{foo}", op),
		"duplicate parameter in path: foo, duplicate parameter in query: bar")
}
