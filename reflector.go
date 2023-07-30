package openapi

import "github.com/swaggest/jsonschema-go"

// Reflector defines OpenAPI reflector behavior.
type Reflector interface {
	JSONSchemaWalker

	NewOperationContext(method, pathPattern string) (OperationContext, error)
	AddOperation(oc OperationContext) error

	SpecSchema() SpecSchema
	JSONSchemaReflector() *jsonschema.Reflector
}

// JSONSchemaCallback is a user function called by JSONSchemaWalker.
type JSONSchemaCallback func(in In, paramName string, schema *jsonschema.SchemaOrBool, required bool) error

// JSONSchemaWalker can extract JSON schemas (for example, for validation purposes) from a ContentUnit.
type JSONSchemaWalker interface {
	ResolveJSONSchemaRef(ref string) (s jsonschema.SchemaOrBool, found bool)
	WalkRequestJSONSchemas(method string, cu ContentUnit, cb JSONSchemaCallback, done func(oc OperationContext)) error
	WalkResponseJSONSchemas(cu ContentUnit, cb JSONSchemaCallback, done func(oc OperationContext)) error
}
