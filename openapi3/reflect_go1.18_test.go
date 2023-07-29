//go:build go1.18
// +build go1.18

package openapi3_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/swaggest/assertjson"
	"github.com/swaggest/openapi-go"
	"github.com/swaggest/openapi-go/openapi3"
)

func Test_Foo(t *testing.T) {
	reflector := openapi3.Reflector{}
	reflector.Spec = &openapi3.Spec{Openapi: "3.0.3"}
	reflector.Spec.Info.
		WithTitle("Things API").
		WithVersion("1.2.3").
		WithDescription("Put something here")

	type req[T any] struct {
		ID     string `path:"id" example:"XXX-XXXXX"`
		Locale string `query:"locale" pattern:"^[a-z]{2}-[A-Z]{2}$"`
		Title  string `json:"string"`
		Amount uint   `json:"amount"`
		Items  []struct {
			Count uint   `json:"count"`
			Name  string `json:"name"`
		} `json:"items"`
	}

	type resp[T any] struct {
		ID     string `json:"id" example:"XXX-XXXXX"`
		Amount uint   `json:"amount"`
		Items  []struct {
			Count uint   `json:"count"`
			Name  string `json:"name"`
		} `json:"items"`
		UpdatedAt time.Time `json:"updated_at"`
	}

	putOC, err := reflector.NewOperationContext(http.MethodPut, "/things/{id}")
	require.NoError(t, err)
	putOC.AddReqStructure(new(req[time.Time]))
	putOC.AddRespStructure(new(resp[time.Time]), func(cu *openapi.ContentUnit) { cu.HTTPStatus = http.StatusOK })
	putOC.AddRespStructure(new([]resp[time.Time]), func(cu *openapi.ContentUnit) { cu.HTTPStatus = http.StatusConflict })
	require.NoError(t, reflector.AddOperation(putOC))

	getOC, err := reflector.NewOperationContext(http.MethodGet, "/things/{id}")
	require.NoError(t, err)
	getOC.AddReqStructure(new(req[time.Time]))
	getOC.AddRespStructure(new(resp[time.Time]), func(cu *openapi.ContentUnit) { cu.HTTPStatus = http.StatusOK })
	require.NoError(t, reflector.AddOperation(getOC))

	assertjson.EqMarshal(t, `{
	  "openapi":"3.0.3",
	  "info":{"title":"Things API","description":"Put something here","version":"1.2.3"},
	  "paths":{
		"/things/{id}":{
		  "get":{
			"parameters":[
			  {
				"name":"locale","in":"query",
				"schema":{"pattern":"^[a-z]{2}-[A-Z]{2}$","type":"string"}
			  },
			  {
				"name":"id","in":"path","required":true,
				"schema":{"type":"string","example":"XXX-XXXXX"}
			  }
			],
			"responses":{
			  "200":{
				"description":"OK",
				"content":{
				  "application/json":{
					"schema":{"$ref":"#/components/schemas/Openapi3TestRespTimeTime"}
				  }
				}
			  }
			}
		  },
		  "put":{
			"parameters":[
			  {
				"name":"locale","in":"query",
				"schema":{"pattern":"^[a-z]{2}-[A-Z]{2}$","type":"string"}
			  },
			  {
				"name":"id","in":"path","required":true,
				"schema":{"type":"string","example":"XXX-XXXXX"}
			  }
			],
			"requestBody":{
			  "content":{
				"application/json":{"schema":{"$ref":"#/components/schemas/Openapi3TestReqTimeTime"}}
			  }
			},
			"responses":{
			  "200":{
				"description":"OK",
				"content":{
				  "application/json":{
					"schema":{"$ref":"#/components/schemas/Openapi3TestRespTimeTime"}
				  }
				}
			  },
			  "409":{
				"description":"Conflict",
				"content":{
				  "application/json":{
					"schema":{
					  "type":"array",
					  "items":{"$ref":"#/components/schemas/Openapi3TestRespTimeTime"}
					}
				  }
				}
			  }
			}
		  }
		}
	  },
	  "components":{
		"schemas":{
		  "Openapi3TestReqTimeTime":{
			"type":"object",
			"properties":{
			  "amount":{"minimum":0,"type":"integer"},
			  "items":{
				"type":"array",
				"items":{
				  "type":"object",
				  "properties":{
					"count":{"minimum":0,"type":"integer"},"name":{"type":"string"}
				  }
				},
				"nullable":true
			  },
			  "string":{"type":"string"}
			}
		  },
		  "Openapi3TestRespTimeTime":{
			"type":"object",
			"properties":{
			  "amount":{"minimum":0,"type":"integer"},
			  "id":{"type":"string","example":"XXX-XXXXX"},
			  "items":{
				"type":"array",
				"items":{
				  "type":"object",
				  "properties":{
					"count":{"minimum":0,"type":"integer"},"name":{"type":"string"}
				  }
				},
				"nullable":true
			  },
			  "updated_at":{"type":"string","format":"date-time"}
			}
		  }
		}
	  }
	}`, reflector.Spec)
}
