package openapi3

import (
	"github.com/swaggest/jsonschema-go"
	"github.com/swaggest/openapi-go"
)

// SetupRequest sets up operation parameters.
//
// Deprecated: instrument openapi.OperationContext and use AddOperation.
func (r *Reflector) SetupRequest(c OperationContext) error {
	return r.setupRequest(c.Operation, toOpCtx(c))
}

// SetRequest sets up operation parameters.
//
// Deprecated: instrument openapi.OperationContext and use AddOperation.
func (r *Reflector) SetRequest(o *Operation, input interface{}, httpMethod string) error {
	return r.SetupRequest(OperationContext{
		Operation:  o,
		Input:      input,
		HTTPMethod: httpMethod,
	})
}

// RequestBodyEnforcer enables request body for GET and HEAD methods.
//
// Should be implemented on input structure, function body can be empty.
// Forcing request body is not recommended and should only be used for backwards compatibility.
//
// Deprecated: use openapi.RequestBodyEnforcer.
type RequestBodyEnforcer interface {
	ForceRequestBody()
}

// RequestJSONBodyEnforcer enables JSON request body for structures with `formData` tags.
//
// Should be implemented on input structure, function body can be empty.
//
// Deprecated: use openapi.RequestJSONBodyEnforcer.
type RequestJSONBodyEnforcer interface {
	ForceJSONRequestBody()
}

// SetStringResponse sets unstructured response.
//
// Deprecated: use AddOperation with openapi.OperationContext AddRespStructure.
func (r *Reflector) SetStringResponse(o *Operation, httpStatus int, contentType string) error {
	return r.SetupResponse(OperationContext{
		Operation:       o,
		HTTPStatus:      httpStatus,
		RespContentType: contentType,
	})
}

// SetJSONResponse sets up operation JSON response.
//
// Deprecated: use AddOperation with openapi.OperationContext AddRespStructure.
func (r *Reflector) SetJSONResponse(o *Operation, output interface{}, httpStatus int) error {
	return r.SetupResponse(OperationContext{
		Operation:  o,
		Output:     output,
		HTTPStatus: httpStatus,
	})
}

// SetupResponse sets up operation response.
//
// Deprecated: use AddOperation with openapi.OperationContext AddRespStructure.
func (r *Reflector) SetupResponse(oc OperationContext) error {
	return r.setupResponse(oc.Operation, toOpCtx(oc))
}

// OperationCtx retrieves operation context from reflect context.
//
// Deprecated: use openapi.OperationCtx.
func OperationCtx(rc *jsonschema.ReflectContext) (OperationContext, bool) {
	if oc, ok := openapi.OperationCtx(rc); ok {
		return fromOpCtx(oc), true
	}

	return OperationContext{}, false
}

// OperationContext describes operation.
//
// Deprecated: use Reflector.NewOperationContext, exported access might be revoked in the future.
type OperationContext struct {
	Operation  *Operation
	Input      interface{}
	HTTPMethod string

	ReqQueryMapping    map[string]string
	ReqPathMapping     map[string]string
	ReqCookieMapping   map[string]string
	ReqHeaderMapping   map[string]string
	ReqFormDataMapping map[string]string

	Output            interface{}
	HTTPStatus        int
	RespContentType   string
	RespHeaderMapping map[string]string

	ProcessingResponse bool
	ProcessingIn       string
}
