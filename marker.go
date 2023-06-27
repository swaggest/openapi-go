package openapi

// RequestBodyEnforcer enables request body for GET and HEAD methods.
//
// Should be implemented on input structure, function body can be empty.
// Forcing request body is not recommended and should only be used for backwards compatibility.
type RequestBodyEnforcer interface {
	ForceRequestBody()
}

// RequestJSONBodyEnforcer enables JSON request body for structures with `formData` tags.
//
// Should be implemented on input structure, function body can be empty.
type RequestJSONBodyEnforcer interface {
	ForceJSONRequestBody()
}
