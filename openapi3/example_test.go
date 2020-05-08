package openapi3_test

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/swaggest/openapi-go/openapi3"
)

func handleError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func ExampleReflector_SetJSONResponse() {
	reflector := openapi3.Reflector{}
	reflector.Spec = &openapi3.Spec{Openapi: "3.0.2"}
	reflector.Spec.Info.
		WithTitle("Things API").
		WithVersion("1.2.3").
		WithDescription("Put something here")

	type req struct {
		ID     string `path:"id" example:"XXX-XXXXX"`
		Locale string `query:"locale" pattern:"^[a-z]{2}-[A-Z]{2}$"`
		Title  string `json:"string"`
		Amount uint   `json:"amount"`
		Items  []struct {
			Count uint   `json:"count"`
			Name  string `json:"name"`
		} `json:"items"`
	}

	type resp struct {
		ID     string `json:"id" example:"XXX-XXXXX"`
		Amount uint   `json:"amount"`
		Items  []struct {
			Count uint   `json:"count"`
			Name  string `json:"name"`
		} `json:"items"`
		UpdatedAt time.Time `json:"updated_at"`
	}

	putOp := openapi3.Operation{}

	handleError(reflector.SetRequest(&putOp, new(req), http.MethodPut))
	handleError(reflector.SetJSONResponse(&putOp, new(resp), http.StatusOK))
	handleError(reflector.SetJSONResponse(&putOp, new([]resp), http.StatusConflict))
	handleError(reflector.Spec.AddOperation(http.MethodPut, "/things/{id}", putOp))

	getOp := openapi3.Operation{}

	handleError(reflector.SetRequest(&getOp, new(req), http.MethodGet))
	handleError(reflector.SetJSONResponse(&getOp, new(resp), http.StatusOK))
	handleError(reflector.Spec.AddOperation(http.MethodGet, "/things/{id}", getOp))

	schema, err := json.MarshalIndent(reflector.Spec, "", " ")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(schema))

	// Output:
	// {
	//  "openapi": "3.0.2",
	//  "info": {
	//   "title": "Things API",
	//   "description": "Put something here",
	//   "version": "1.2.3"
	//  },
	//  "paths": {
	//   "/things/{id}": {
	//    "get": {
	//     "parameters": [
	//      {
	//       "name": "locale",
	//       "in": "query",
	//       "schema": {
	//        "pattern": "^[a-z]{2}-[A-Z]{2}$",
	//        "type": "string"
	//       }
	//      },
	//      {
	//       "name": "id",
	//       "in": "path",
	//       "required": true,
	//       "schema": {
	//        "type": "string",
	//        "example": "XXX-XXXXX"
	//       }
	//      }
	//     ],
	//     "responses": {
	//      "200": {
	//       "description": "OK",
	//       "content": {
	//        "application/json": {
	//         "schema": {
	//          "$ref": "#/components/schemas/Openapi3TestResp"
	//         }
	//        }
	//       }
	//      }
	//     }
	//    },
	//    "put": {
	//     "parameters": [
	//      {
	//       "name": "locale",
	//       "in": "query",
	//       "schema": {
	//        "pattern": "^[a-z]{2}-[A-Z]{2}$",
	//        "type": "string"
	//       }
	//      },
	//      {
	//       "name": "id",
	//       "in": "path",
	//       "required": true,
	//       "schema": {
	//        "type": "string",
	//        "example": "XXX-XXXXX"
	//       }
	//      }
	//     ],
	//     "requestBody": {
	//      "content": {
	//       "application/json": {
	//        "schema": {
	//         "$ref": "#/components/schemas/Openapi3TestReq"
	//        }
	//       }
	//      }
	//     },
	//     "responses": {
	//      "200": {
	//       "description": "OK",
	//       "content": {
	//        "application/json": {
	//         "schema": {
	//          "$ref": "#/components/schemas/Openapi3TestResp"
	//         }
	//        }
	//       }
	//      },
	//      "409": {
	//       "description": "Conflict",
	//       "content": {
	//        "application/json": {
	//         "schema": {
	//          "type": "array",
	//          "items": {
	//           "$ref": "#/components/schemas/Openapi3TestResp"
	//          }
	//         }
	//        }
	//       }
	//      }
	//     }
	//    }
	//   }
	//  },
	//  "components": {
	//   "schemas": {
	//    "Openapi3TestReq": {
	//     "type": "object",
	//     "properties": {
	//      "amount": {
	//       "minimum": 0,
	//       "type": "integer"
	//      },
	//      "items": {
	//       "type": "array",
	//       "items": {
	//        "type": "object",
	//        "properties": {
	//         "count": {
	//          "minimum": 0,
	//          "type": "integer"
	//         },
	//         "name": {
	//          "type": "string"
	//         }
	//        }
	//       }
	//      },
	//      "string": {
	//       "type": "string"
	//      }
	//     }
	//    },
	//    "Openapi3TestResp": {
	//     "type": "object",
	//     "properties": {
	//      "amount": {
	//       "minimum": 0,
	//       "type": "integer"
	//      },
	//      "id": {
	//       "type": "string",
	//       "example": "XXX-XXXXX"
	//      },
	//      "items": {
	//       "type": "array",
	//       "items": {
	//        "type": "object",
	//        "properties": {
	//         "count": {
	//          "minimum": 0,
	//          "type": "integer"
	//         },
	//         "name": {
	//          "type": "string"
	//         }
	//        }
	//       }
	//      },
	//      "updated_at": {
	//       "type": "string",
	//       "format": "date-time"
	//      }
	//     }
	//    }
	//   }
	//  }
	// }
}
