package openapi31_test

import (
	"mime/multipart"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/swaggest/assertjson"
	"github.com/swaggest/openapi-go/openapi31"
)

func TestNewReflector_uploads(t *testing.T) {
	r := openapi31.NewReflector()

	oc, err := r.NewOperationContext(http.MethodPost, "/upload")
	require.NoError(t, err)

	type req struct {
		Upload1  multipart.File          `formData:"upload1"`
		Upload2  *multipart.FileHeader   `formData:"upload2"`
		Uploads3 []multipart.File        `formData:"uploads3"`
		Uploads4 []*multipart.FileHeader `formData:"uploads4"`
	}

	oc.AddReqStructure(req{})

	require.NoError(t, r.AddOperation(oc))

	schema, err := assertjson.MarshalIndentCompact(r.SpecSchema(), "", " ", 120)
	require.NoError(t, err)

	require.NoError(t, os.WriteFile("testdata/uploads_last_run.json", schema, 0o600))

	expected, err := os.ReadFile("testdata/uploads.json")
	require.NoError(t, err)

	assertjson.Equal(t, expected, schema)
}
