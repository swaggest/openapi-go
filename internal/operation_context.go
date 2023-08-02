// Package internal keeps reusable internal code.
package internal

import "github.com/swaggest/openapi-go"

// NewOperationContext creates OperationContext.
func NewOperationContext(method, pathPattern string) *OperationContext {
	return &OperationContext{
		method:      method,
		pathPattern: pathPattern,
	}
}

// OperationContext implements openapi.OperationContext.
type OperationContext struct {
	method      string
	pathPattern string
	req         []openapi.ContentUnit
	resp        []openapi.ContentUnit

	isProcessingResponse bool
	processingIn         openapi.In
}

// Method returns HTTP method of an operation.
func (o *OperationContext) Method() string {
	return o.method
}

// PathPattern returns operation HTTP URL path pattern.
func (o *OperationContext) PathPattern() string {
	return o.pathPattern
}

// Request returns list of operation request content schemas.
func (o *OperationContext) Request() []openapi.ContentUnit {
	return o.req
}

// Response returns list of operation response content schemas.
func (o *OperationContext) Response() []openapi.ContentUnit {
	return o.resp
}

// SetIsProcessingResponse sets current processing state.
func (o *OperationContext) SetIsProcessingResponse(is bool) {
	o.isProcessingResponse = is
}

// IsProcessingResponse indicates if response is being processed.
func (o *OperationContext) IsProcessingResponse() bool {
	return o.isProcessingResponse
}

// SetProcessingIn sets current content location being processed.
func (o *OperationContext) SetProcessingIn(in openapi.In) {
	o.processingIn = in
}

// ProcessingIn return which content location is being processed now.
func (o *OperationContext) ProcessingIn() openapi.In {
	return o.processingIn
}

// SetMethod sets HTTP method of an operation.
func (o *OperationContext) SetMethod(method string) {
	o.method = method
}

// SetPathPattern sets URL path pattern of an operation.
func (o *OperationContext) SetPathPattern(pattern string) {
	o.pathPattern = pattern
}

// AddReqStructure adds request content schema.
func (o *OperationContext) AddReqStructure(s interface{}, options ...openapi.ContentOption) {
	c := openapi.ContentUnit{}

	if cp, ok := s.(openapi.ContentUnitPreparer); ok {
		cp.SetupContentUnit(&c)
	} else {
		c.Structure = s
	}

	for _, o := range options {
		o(&c)
	}

	o.req = append(o.req, c)
}

// AddRespStructure adds response content schema.
func (o *OperationContext) AddRespStructure(s interface{}, options ...openapi.ContentOption) {
	c := openapi.ContentUnit{}

	if cp, ok := s.(openapi.ContentUnitPreparer); ok {
		cp.SetupContentUnit(&c)
	} else {
		c.Structure = s
	}

	for _, o := range options {
		o(&c)
	}

	o.resp = append(o.resp, c)
}
