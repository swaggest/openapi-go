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

// OperationPreparer controls operation parameters.
type OperationPreparer interface {
	SetMethod(method string)
	SetPathPattern(pattern string)
	AddReqStructure(i interface{}, options ...ContentOption)
	AddRespStructure(o interface{}, options ...ContentOption)
}

// OperationContext defines operation and processing state.
type OperationContext interface {
	Method() string
	PathPattern() string

	Request() []ContentUnit
	Response() []ContentUnit

	SetIsProcessingResponse(bool)
	IsProcessingResponse() bool

	SetProcessingIn(in In)
	ProcessingIn() In
}

// NewOperationContext creates OperationContext.
func NewOperationContext(method, pathPattern string) *OpCtx {
	return &OpCtx{
		method:      method,
		pathPattern: pathPattern,
	}
}

// OpCtx implements OperationContext and OperationPreparer.
type OpCtx struct {
	method      string
	pathPattern string
	req         []ContentUnit
	resp        []ContentUnit

	isProcessingResponse bool
	processingIn         In
}

// Method returns HTTP method of an operation.
func (o *OpCtx) Method() string {
	return o.method
}

// PathPattern returns operation HTTP URL path pattern.
func (o *OpCtx) PathPattern() string {
	return o.pathPattern
}

// Request returns list of operation request content schemas.
func (o *OpCtx) Request() []ContentUnit {
	return o.req
}

// Response returns list of operation response content schemas.
func (o *OpCtx) Response() []ContentUnit {
	return o.resp
}

// SetIsProcessingResponse sets current processing state.
func (o *OpCtx) SetIsProcessingResponse(is bool) {
	o.isProcessingResponse = is
}

// IsProcessingResponse indicates if response is being processed.
func (o *OpCtx) IsProcessingResponse() bool {
	return o.isProcessingResponse
}

// SetProcessingIn sets current content location being processed.
func (o *OpCtx) SetProcessingIn(in In) {
	o.processingIn = in
}

// ProcessingIn return which content location is being processed now.
func (o *OpCtx) ProcessingIn() In {
	return o.processingIn
}

// SetMethod sets HTTP method of an operation.
func (o *OpCtx) SetMethod(method string) {
	o.method = method
}

// SetPathPattern sets URL path pattern of an operation.
func (o *OpCtx) SetPathPattern(pattern string) {
	o.pathPattern = pattern
}

// AddReqStructure adds request content schema.
func (o *OpCtx) AddReqStructure(s interface{}, options ...ContentOption) {
	c := ContentUnit{}
	c.Structure = s

	for _, o := range options {
		o(&c)
	}

	o.req = append(o.req, c)
}

// AddRespStructure adds response content schema.
func (o *OpCtx) AddRespStructure(s interface{}, options ...ContentOption) {
	c := ContentUnit{}
	c.Structure = s

	for _, o := range options {
		o(&c)
	}

	o.resp = append(o.resp, c)
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
