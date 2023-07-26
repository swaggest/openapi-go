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
	Description  string
	fieldMapping map[In]map[string]string
}

// SetFieldMapping sets custom field mapping.
func (c *ContentUnit) SetFieldMapping(in In, fieldToParamName map[string]string) {
	if len(fieldToParamName) == 0 {
		return
	}

	if c.fieldMapping == nil {
		c.fieldMapping = make(map[In]map[string]string)
	}

	c.fieldMapping[in] = fieldToParamName
}

// FieldMapping returns custom field mapping.
func (c ContentUnit) FieldMapping(in In) map[string]string {
	return c.fieldMapping[in]
}

// OperationContext defines operation and processing state.
type OperationContext interface {
	OperationInfo
	OperationState

	Method() string
	PathPattern() string

	Request() []ContentUnit
	Response() []ContentUnit

	AddReqStructure(i interface{}, options ...ContentOption)
	AddRespStructure(o interface{}, options ...ContentOption)
}

// OperationInfo extends OperationContext with general information.
type OperationInfo interface {
	SetTags(tags ...string)
	SetIsDeprecated(isDeprecated bool)
	SetSummary(summary string)
	SetDescription(description string)
	SetID(operationID string)
}

// OperationState extends OperationContext with processing state information.
type OperationState interface {
	IsProcessingResponse() bool
	ProcessingIn() In

	SetIsProcessingResponse(bool)
	SetProcessingIn(in In)
}

type ocCtxKey struct{}

// WithOperationCtx is a jsonschema.ReflectContext option.
func WithOperationCtx(oc OperationContext, isProcessingResponse bool, in In) func(rc *jsonschema.ReflectContext) {
	return func(rc *jsonschema.ReflectContext) {
		oc.SetIsProcessingResponse(isProcessingResponse)
		oc.SetProcessingIn(in)

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
