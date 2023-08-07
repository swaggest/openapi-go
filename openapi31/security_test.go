package openapi31_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/swaggest/assertjson"
	"github.com/swaggest/openapi-go"
	"github.com/swaggest/openapi-go/openapi31"
)

func TestSpec_SetHTTPBasicSecurity(t *testing.T) {
	reflector := openapi31.Reflector{}
	securityName := "admin"

	// Declare security scheme.
	reflector.SpecEns().SetHTTPBasicSecurity(securityName, "Admin Access")

	oc, err := reflector.NewOperationContext(http.MethodGet, "/secure")
	require.NoError(t, err)
	oc.AddRespStructure(struct {
		Secret string `json:"secret"`
	}{})

	// Add security requirement to operation.
	oc.AddSecurity(securityName)

	// Describe unauthorized response.
	oc.AddRespStructure(struct {
		Error string `json:"error"`
	}{}, func(cu *openapi.ContentUnit) {
		cu.HTTPStatus = http.StatusUnauthorized
	})

	// Add operation to schema.
	require.NoError(t, reflector.AddOperation(oc))

	assertjson.EqMarshal(t, `{
	  "openapi":"3.1.0","info":{"title":"","version":""},
	  "paths":{
		"/secure":{
		  "get":{
			"responses":{
			  "200":{
				"description":"OK",
				"content":{
				  "application/json":{
					"schema":{"properties":{"secret":{"type":"string"}},"type":"object"}
				  }
				}
			  },
			  "401":{
				"description":"Unauthorized",
				"content":{
				  "application/json":{
					"schema":{"properties":{"error":{"type":"string"}},"type":"object"}
				  }
				}
			  }
			},
			"security":[{"admin":[]}]
		  }
		}
	  },
	  "components":{
		"securitySchemes":{"admin":{"description":"Admin Access","type":"http","scheme":"basic"}}
	  }
	}`, reflector.SpecSchema())
}

func TestSpec_SetAPIKeySecurity(t *testing.T) {
	reflector := openapi31.Reflector{}
	securityName := "admin"

	// Declare security scheme.
	reflector.SpecEns().SetAPIKeySecurity("User", "sessid", "cookie", "Session cookie.")

	oc, err := reflector.NewOperationContext(http.MethodGet, "/secure")
	require.NoError(t, err)
	oc.AddRespStructure(struct {
		Secret string `json:"secret"`
	}{})

	// Add security requirement to operation.
	oc.AddSecurity(securityName)

	// Describe unauthorized response.
	oc.AddRespStructure(struct {
		Error string `json:"error"`
	}{}, func(cu *openapi.ContentUnit) {
		cu.HTTPStatus = http.StatusUnauthorized
	})

	// Add operation to schema.
	require.NoError(t, reflector.AddOperation(oc))

	assertjson.EqMarshal(t, `{
	  "openapi":"3.1.0","info":{"title":"","version":""},
	  "paths":{
		"/secure":{
		  "get":{
			"responses":{
			  "200":{
				"description":"OK",
				"content":{
				  "application/json":{
					"schema":{"properties":{"secret":{"type":"string"}},"type":"object"}
				  }
				}
			  },
			  "401":{
				"description":"Unauthorized",
				"content":{
				  "application/json":{
					"schema":{"properties":{"error":{"type":"string"}},"type":"object"}
				  }
				}
			  }
			},
			"security":[{"admin":[]}]
		  }
		}
	  },
	  "components":{
		"securitySchemes":{
		  "User":{
			"description":"Session cookie.","type":"apiKey","name":"sessid",
			"in":"cookie"
		  }
		}
	  }
	}`, reflector.SpecSchema())
}
