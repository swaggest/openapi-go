package openapi_test

import (
	"net/http"
	"testing"

	"github.com/swaggest/openapi-go"
	"github.com/swaggest/openapi-go/openapi31"
)

// ErrorResponse represents a common error response structure
type ErrorResponseTest struct {
	Message string `json:"message" description:"Error message"`
	Code    int    `json:"code" description:"Error code"`
}

// User represents a user object
type UserTest struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func TestResponseReuse(t *testing.T) {
	// Create reflector with ReuseResponses enabled
	reflector := openapi31.NewReflector()
	reflector.ReuseResponses = true

	reflector.SpecEns()
	reflector.Spec.Info.WithTitle("API with Reusable Responses")
	reflector.Spec.Info.WithVersion("1.0.0")
	reflector.Spec.Info.WithDescription("Demonstrates response reuse in components")

	// Operation 1: Get User
	type getUserInput struct {
		ID int `path:"id"`
	}
	getUserOp, _ := reflector.NewOperationContext(http.MethodGet, "/users/{id}")
	getUserOp.SetDescription("Get user by ID")
	getUserOp.SetID("getUser")
	getUserOp.AddReqStructure(getUserInput{})

	// Add success response
	getUserOp.AddRespStructure(UserTest{}, openapi.WithHTTPStatus(http.StatusOK))

	// Add error responses - these will be collected in components/responses
	getUserOp.AddRespStructure(ErrorResponseTest{},
		openapi.WithHTTPStatus(http.StatusBadRequest))
	getUserOp.AddRespStructure(ErrorResponseTest{},
		openapi.WithHTTPStatus(http.StatusNotFound))
	getUserOp.AddRespStructure(ErrorResponseTest{},
		openapi.WithHTTPStatus(http.StatusInternalServerError))

	if err := reflector.AddOperation(getUserOp); err != nil {
		t.Fatal(err)
	}

	// Operation 2: Update User - will reuse the same error responses
	type updateUserInput struct {
		ID int `path:"id"`
	}
	updateUserOp, _ := reflector.NewOperationContext(http.MethodPut, "/users/{id}")
	updateUserOp.SetDescription("Update user by ID")
	updateUserOp.SetID("updateUser")
	updateUserOp.AddReqStructure(updateUserInput{})

	// Add success response
	updateUserOp.AddRespStructure(UserTest{}, openapi.WithHTTPStatus(http.StatusOK))

	// Add the same error responses - they will reference the components
	updateUserOp.AddRespStructure(ErrorResponseTest{},
		openapi.WithHTTPStatus(http.StatusBadRequest))
	updateUserOp.AddRespStructure(ErrorResponseTest{},
		openapi.WithHTTPStatus(http.StatusNotFound))
	updateUserOp.AddRespStructure(ErrorResponseTest{},
		openapi.WithHTTPStatus(http.StatusInternalServerError))

	if err := reflector.AddOperation(updateUserOp); err != nil {
		t.Fatal(err)
	}

	// Marshal to YAML to see the result
	schema, err := reflector.Spec.MarshalYAML()
	if err != nil {
		t.Fatal(err)
	}

	yamlStr := string(schema)

	// Verify that components/responses is populated
	if reflector.Spec.Components == nil || reflector.Spec.Components.Responses == nil {
		t.Fatal("Expected components/responses to be populated")
	}

	// Verify we have the error responses in components
	if len(reflector.Spec.Components.Responses) == 0 {
		t.Fatal("Expected error responses in components/responses")
	}

	t.Logf("Generated OpenAPI spec:\n%s", yamlStr)
	t.Logf("Number of component responses: %d", len(reflector.Spec.Components.Responses))
}

func TestResponseReuseDisabled(t *testing.T) {
	// Create reflector with ReuseResponses disabled (default)
	reflector := openapi31.NewReflector()
	// ReuseResponses is false by default

	reflector.SpecEns()
	reflector.Spec.Info.WithTitle("API without Response Reuse")
	reflector.Spec.Info.WithVersion("1.0.0")

	// Operation 1
	type getUserInput struct {
		ID int `path:"id"`
	}
	getUserOp, _ := reflector.NewOperationContext(http.MethodGet, "/users/{id}")
	getUserOp.SetDescription("Get user by ID")
	getUserOp.AddReqStructure(getUserInput{})
	getUserOp.AddRespStructure(UserTest{}, openapi.WithHTTPStatus(http.StatusOK))
	getUserOp.AddRespStructure(ErrorResponseTest{}, openapi.WithHTTPStatus(http.StatusBadRequest))

	if err := reflector.AddOperation(getUserOp); err != nil {
		t.Fatal(err)
	}

	// With ReuseResponses disabled, components/responses should be empty or nil
	if reflector.Spec.Components != nil && reflector.Spec.Components.Responses != nil && len(reflector.Spec.Components.Responses) > 0 {
		t.Fatal("Expected no responses in components/responses when ReuseResponses is disabled")
	}

	t.Log("Response reuse disabled - responses inlined as expected")
}
