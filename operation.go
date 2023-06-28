package openapi

import (
	"context"

	"github.com/swaggest/jsonschema-go"
)

// In defines value location in HTTP content.
type In string

// In values enumeration.
const (
	InPath     = In("path")
	InQuery    = In("query")
	InHeader   = In("header")
	InCookie   = In("cookie")
	InFormData = In("formData")
	InBody     = In("body")
)

// ContentOption configures ContentUnit.
type ContentOption func(cu *ContentUnit)

// ContentUnit defines HTTP content.
type ContentUnit struct {
	Structure    interface{}
	ContentType  string
	Format       string
	HTTPStatus   int
	IsSchemaLess bool
	Description  string
	fieldMapping map[In]map[string]string
}

// FieldMapping returns custom field mapping.
func (c ContentUnit) FieldMapping(in In) map[string]string {
	return c.fieldMapping[in]
}

// FieldMapping is an option to set up custom field mapping (instead of field tags).
func FieldMapping(in In, fieldToParamName map[string]string) ContentOption {
	return func(c *ContentUnit) {
		if len(fieldToParamName) == 0 {
			return
		}

		if c.fieldMapping == nil {
			c.fieldMapping = make(map[In]map[string]string)
		}

		c.fieldMapping[in] = fieldToParamName
	}
}

// OperationContext defines operation and processing state.
type OperationContext interface {
	Method() string
	PathPattern() string

	Request() []ContentUnit
	Response() []ContentUnit

	AddReqStructure(i interface{}, options ...ContentOption)
	AddRespStructure(o interface{}, options ...ContentOption)

	IsProcessingResponse() bool
	ProcessingIn() In
}

type operationState interface {
	SetIsProcessingResponse(bool)
	SetProcessingIn(in In)
}

type ocCtxKey struct{}

// WithOperationCtx is a jsonschema.ReflectContext option.
func WithOperationCtx(oc OperationContext, isProcessingResponse bool, in In) func(rc *jsonschema.ReflectContext) {
	return func(rc *jsonschema.ReflectContext) {
		if os, ok := oc.(operationState); ok {
			os.SetIsProcessingResponse(isProcessingResponse)
			os.SetProcessingIn(in)
		}

		rc.Context = context.WithValue(rc.Context, ocCtxKey{}, oc)
	}
}

// OperationCtx retrieves operation context from reflect context.
func OperationCtx(rc *jsonschema.ReflectContext) (OperationContext, bool) {
	if oc, ok := rc.Value(ocCtxKey{}).(OperationContext); ok {
		return oc, true
	}

	return nil, false
}
