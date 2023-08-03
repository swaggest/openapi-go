package openapi_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/swaggest/openapi-go"
)

func TestContentUnit_options(t *testing.T) {
	cu := openapi.ContentUnit{}
	openapi.WithContentType("text/csv")(&cu)
	openapi.WithHTTPStatus(http.StatusConflict)(&cu)

	assert.Equal(t, "text/csv", cu.ContentType)
	assert.Equal(t, http.StatusConflict, cu.HTTPStatus)
}
