package openapi31_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/swaggest/openapi-go/openapi31"
)

func TestSpec_UnmarshalYAML(t *testing.T) {
	bytes, err := os.ReadFile("testdata/albums_api.yaml")
	require.NoError(t, err)

	refl := openapi31.NewReflector()
	require.NoError(t, refl.Spec.UnmarshalYAML(bytes))
}

func TestSpec_UnmarshalYAML_refsInResponseHeaders(t *testing.T) {
	var s openapi31.Spec

	spec := `openapi: 3.1.0
info:
  description: description
  license:
    name: Apache-2.0
    url: https://www.apache.org/licenses/LICENSE-2.0.html
  title: title
  version: 2.0.0
servers:
  - url: /v2
paths:
  /user:
    put:
      summary: updates the user by id
      operationId: UpdateUser
      requestBody:
        content:
          application/json:
            schema:
              type: string
        description: Updated user object
        required: true
      responses:
        "404":
          description: User not found
          headers:
            Cache-Control:
              $ref: '#/components/headers/CacheControl'
            Authorisation:
              schema:
                type: string
            Custom:
              content:
                "text/plain":
                  schema:
                  type: string

components:
  headers:
    CacheControl:
      schema:
        type: string
`

	require.NoError(t, s.UnmarshalYAML([]byte(spec)))
}
