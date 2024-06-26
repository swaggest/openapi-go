// Code generated by github.com/swaggest/json-cli v1.8.6, DO NOT EDIT.

package openapi3

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/swaggest/assertjson"
)

func TestSpec_MarshalJSON_roundtrip(t *testing.T) {
	var (
		jsonValue = []byte(`{"openapi":"cbbfff","info":{"title":"cdef","description":"aadce","termsOfService":"dc","contact":{"name":"daecfe","url":"cddbd","email":"accdf"},"license":{"name":"fd","url":"ad"},"version":"ead"},"externalDocs":{"description":"adbfc","url":"aebd"},"servers":[{"url":"ddbfe","description":"bfcce","variables":{"fdfed":{"enum":["cdba"],"default":"bedceb","description":"dadf"}}}],"security":[{"bffd":["eae"]}],"tags":[{"name":"db","description":"bfdbc","externalDocs":{"description":"ec","url":"eb"}}],"paths":{},"components":{"schemas":{},"responses":{},"parameters":{},"examples":{},"requestBodies":{},"headers":{},"securitySchemes":{},"links":{},"callbacks":{}}}`)
		v         Spec
	)

	require.NoError(t, json.Unmarshal(jsonValue, &v))

	marshaled, err := json.Marshal(v)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(marshaled, &v))
	assertjson.Equal(t, jsonValue, marshaled)
}

func TestInfo_MarshalJSON_roundtrip(t *testing.T) {
	var (
		jsonValue = []byte(`{"title":"fffbce","description":"dca","termsOfService":"ecaddd","contact":{"name":"bcf","url":"fa","email":"abdfad"},"license":{"name":"aeab","url":"bafd"},"version":"eaceae"}`)
		v         Info
	)

	require.NoError(t, json.Unmarshal(jsonValue, &v))

	marshaled, err := json.Marshal(v)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(marshaled, &v))
	assertjson.Equal(t, jsonValue, marshaled)
}

func TestContact_MarshalJSON_roundtrip(t *testing.T) {
	var (
		jsonValue = []byte(`{"name":"cc","url":"eadf","email":"ecdaea"}`)
		v         Contact
	)

	require.NoError(t, json.Unmarshal(jsonValue, &v))

	marshaled, err := json.Marshal(v)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(marshaled, &v))
	assertjson.Equal(t, jsonValue, marshaled)
}

func TestLicense_MarshalJSON_roundtrip(t *testing.T) {
	var (
		jsonValue = []byte(`{"name":"bcacc","url":"fcadff"}`)
		v         License
	)

	require.NoError(t, json.Unmarshal(jsonValue, &v))

	marshaled, err := json.Marshal(v)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(marshaled, &v))
	assertjson.Equal(t, jsonValue, marshaled)
}

func TestExternalDocumentation_MarshalJSON_roundtrip(t *testing.T) {
	var (
		jsonValue = []byte(`{"description":"eca","url":"bdb"}`)
		v         ExternalDocumentation
	)

	require.NoError(t, json.Unmarshal(jsonValue, &v))

	marshaled, err := json.Marshal(v)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(marshaled, &v))
	assertjson.Equal(t, jsonValue, marshaled)
}

func TestServer_MarshalJSON_roundtrip(t *testing.T) {
	var (
		jsonValue = []byte(`{"url":"dfe","description":"cfdcf","variables":{"baf":{"enum":["dda"],"default":"acdcfe","description":"de"}}}`)
		v         Server
	)

	require.NoError(t, json.Unmarshal(jsonValue, &v))

	marshaled, err := json.Marshal(v)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(marshaled, &v))
	assertjson.Equal(t, jsonValue, marshaled)
}

func TestServerVariable_MarshalJSON_roundtrip(t *testing.T) {
	var (
		jsonValue = []byte(`{"enum":["befdf"],"default":"dcba","description":"bffbb"}`)
		v         ServerVariable
	)

	require.NoError(t, json.Unmarshal(jsonValue, &v))

	marshaled, err := json.Marshal(v)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(marshaled, &v))
	assertjson.Equal(t, jsonValue, marshaled)
}

func TestTag_MarshalJSON_roundtrip(t *testing.T) {
	var (
		jsonValue = []byte(`{"name":"cdf","description":"caf","externalDocs":{"description":"eeacab","url":"bd"}}`)
		v         Tag
	)

	require.NoError(t, json.Unmarshal(jsonValue, &v))

	marshaled, err := json.Marshal(v)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(marshaled, &v))
	assertjson.Equal(t, jsonValue, marshaled)
}

func TestPathItem_MarshalJSON_roundtrip(t *testing.T) {
	var (
		jsonValue = []byte(`{"$ref":"db","summary":"ebae","description":"fcbcf","servers":[{"url":"ce","description":"bb","variables":{"eabafb":{"enum":["ab"],"default":"daac","description":"cba"}}}],"parameters":[{"$ref":"#/components/parameters/Foo"}]}`)
		v         PathItem
	)

	require.NoError(t, json.Unmarshal(jsonValue, &v))

	marshaled, err := json.Marshal(v)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(marshaled, &v))
	assertjson.Equal(t, jsonValue, marshaled)
}

func TestParameterReference_MarshalJSON_roundtrip(t *testing.T) {
	var (
		jsonValue = []byte(`{"$ref":"#/components/parameters/Foo"}`)
		v         ParameterReference
	)

	require.NoError(t, json.Unmarshal(jsonValue, &v))

	marshaled, err := json.Marshal(v)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(marshaled, &v))
	assertjson.Equal(t, jsonValue, marshaled)
}

func TestParameter_MarshalJSON_roundtrip(t *testing.T) {
	var (
		jsonValue = []byte(`{"name":"id","in":"path","required":true,"schema":{"type":"string"}}`)
		v         Parameter
	)

	require.NoError(t, json.Unmarshal(jsonValue, &v))

	marshaled, err := json.Marshal(v)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(marshaled, &v))
	assertjson.Equal(t, jsonValue, marshaled)
}

func TestSchema_MarshalJSON_roundtrip(t *testing.T) {
	var (
		jsonValue = []byte(`{"title":"bed","multipleOf":9700.2,"maximum":501.131,"exclusiveMaximum":true,"minimum":8468.441,"exclusiveMinimum":true,"maxLength":7410,"minLength":2398,"pattern":"cc","maxItems":6200,"minItems":9818,"uniqueItems":true,"maxProperties":8290,"minProperties":8930,"required":["bfee"],"enum":[null],"type":"boolean","not":{"$ref":"#/components/schemas/Foo"},"allOf":[{"$ref":"#/components/schemas/Foo"}],"oneOf":[{"$ref":"#/components/schemas/Foo"}],"anyOf":[{"$ref":"#/components/schemas/Foo"}],"items":{"$ref":"#/components/schemas/Foo"},"properties":{"eaeaae":{"type":"string"}},"additionalProperties":true,"description":"ac","format":"abdb","nullable":true,"discriminator":{"propertyName":"cccf","mapping":{"bed":"cec"}},"readOnly":true,"writeOnly":true,"externalDocs":{"description":"efecaf","url":"fdeeb"},"deprecated":true,"xml":{"name":"afcfaf","namespace":"ec","prefix":"eaeeaf","attribute":true,"wrapped":true}}`)
		v         Schema
	)

	require.NoError(t, json.Unmarshal(jsonValue, &v))

	marshaled, err := json.Marshal(v)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(marshaled, &v))
	assertjson.Equal(t, jsonValue, marshaled)
}

func TestSchemaReference_MarshalJSON_roundtrip(t *testing.T) {
	var (
		jsonValue = []byte(`{"$ref":"#/components/schemas/Foo"}`)
		v         SchemaReference
	)

	require.NoError(t, json.Unmarshal(jsonValue, &v))

	marshaled, err := json.Marshal(v)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(marshaled, &v))
	assertjson.Equal(t, jsonValue, marshaled)
}

func TestSchemaOrRef_MarshalJSON_roundtrip(t *testing.T) {
	var (
		jsonValue = []byte(`{"type":"string"}`)
		v         SchemaOrRef
	)

	require.NoError(t, json.Unmarshal(jsonValue, &v))

	marshaled, err := json.Marshal(v)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(marshaled, &v))
	assertjson.Equal(t, jsonValue, marshaled)
}

func TestSchemaAdditionalProperties_MarshalJSON_roundtrip(t *testing.T) {
	var (
		jsonValue = []byte(`{"$ref":"#/components/schemas/Foo"}`)
		v         SchemaAdditionalProperties
	)

	require.NoError(t, json.Unmarshal(jsonValue, &v))

	marshaled, err := json.Marshal(v)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(marshaled, &v))
	assertjson.Equal(t, jsonValue, marshaled)
}

func TestDiscriminator_MarshalJSON_roundtrip(t *testing.T) {
	var (
		jsonValue = []byte(`{"propertyName":"fdcbaf","mapping":{"abcdaf":"ac"}}`)
		v         Discriminator
	)

	require.NoError(t, json.Unmarshal(jsonValue, &v))

	marshaled, err := json.Marshal(v)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(marshaled, &v))
	assertjson.Equal(t, jsonValue, marshaled)
}

func TestXML_MarshalJSON_roundtrip(t *testing.T) {
	var (
		jsonValue = []byte(`{"name":"dcefd","namespace":"dfecaf","prefix":"ffaf","attribute":true,"wrapped":true}`)
		v         XML
	)

	require.NoError(t, json.Unmarshal(jsonValue, &v))

	marshaled, err := json.Marshal(v)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(marshaled, &v))
	assertjson.Equal(t, jsonValue, marshaled)
}

func TestMediaType_MarshalJSON_roundtrip(t *testing.T) {
	var (
		jsonValue = []byte(`{"schema":{"$ref":"#/components/schemas/Foo"}}`)
		v         MediaType
	)

	require.NoError(t, json.Unmarshal(jsonValue, &v))

	marshaled, err := json.Marshal(v)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(marshaled, &v))
	assertjson.Equal(t, jsonValue, marshaled)
}

func TestExampleReference_MarshalJSON_roundtrip(t *testing.T) {
	var (
		jsonValue = []byte(`{"$ref":"#/components/examples/Foo"}`)
		v         ExampleReference
	)

	require.NoError(t, json.Unmarshal(jsonValue, &v))

	marshaled, err := json.Marshal(v)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(marshaled, &v))
	assertjson.Equal(t, jsonValue, marshaled)
}

func TestExample_MarshalJSON_roundtrip(t *testing.T) {
	var (
		jsonValue = []byte(`{"summary":"An Example","description":"This is an example.","value":"foo"}`)
		v         Example
	)

	require.NoError(t, json.Unmarshal(jsonValue, &v))

	marshaled, err := json.Marshal(v)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(marshaled, &v))
	assertjson.Equal(t, jsonValue, marshaled)
}

func TestExampleOrRef_MarshalJSON_roundtrip(t *testing.T) {
	var (
		jsonValue = []byte(`{"$ref":"#/components/examples/Foo"}`)
		v         ExampleOrRef
	)

	require.NoError(t, json.Unmarshal(jsonValue, &v))

	marshaled, err := json.Marshal(v)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(marshaled, &v))
	assertjson.Equal(t, jsonValue, marshaled)
}

func TestEncoding_MarshalJSON_roundtrip(t *testing.T) {
	var (
		jsonValue = []byte(`{"contentType":"cd","headers":{"bdeb":{"style":"simple","schema":{"type":"string","description":"Sample header response."}}},"style":"form","explode":true,"allowReserved":true}`)
		v         Encoding
	)

	require.NoError(t, json.Unmarshal(jsonValue, &v))

	marshaled, err := json.Marshal(v)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(marshaled, &v))
	assertjson.Equal(t, jsonValue, marshaled)
}

func TestHeader_MarshalJSON_roundtrip(t *testing.T) {
	var (
		jsonValue = []byte(`{"style":"simple","schema":{"type":"string","description":"Sample header response."}}`)
		v         Header
	)

	require.NoError(t, json.Unmarshal(jsonValue, &v))

	marshaled, err := json.Marshal(v)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(marshaled, &v))
	assertjson.Equal(t, jsonValue, marshaled)
}

func TestHasSchema_MarshalJSON_roundtrip(t *testing.T) {
	var (
		jsonValue = []byte(`{"schema":{"type":"string"}}`)
		v         HasSchema
	)

	require.NoError(t, json.Unmarshal(jsonValue, &v))

	marshaled, err := json.Marshal(v)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(marshaled, &v))
	assertjson.Equal(t, jsonValue, marshaled)
}

func TestHasContent_MarshalJSON_roundtrip(t *testing.T) {
	var (
		jsonValue = []byte(`{"content":{"application/json":{"schema":{"$ref":"#/components/schemas/Foo"}}}}`)
		v         HasContent
	)

	require.NoError(t, json.Unmarshal(jsonValue, &v))

	marshaled, err := json.Marshal(v)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(marshaled, &v))
	assertjson.Equal(t, jsonValue, marshaled)
}

func TestSchemaXORContent_MarshalJSON_roundtrip(t *testing.T) {
	var (
		jsonValue = []byte(`{"content":{"application/json":{"schema":{"$ref":"#/components/schemas/Foo"}}}}`)
		v         SchemaXORContent
	)

	require.NoError(t, json.Unmarshal(jsonValue, &v))

	marshaled, err := json.Marshal(v)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(marshaled, &v))
	assertjson.Equal(t, jsonValue, marshaled)
}

func TestSchemaXORContentNot_MarshalJSON_roundtrip(t *testing.T) {
	var (
		jsonValue = []byte(`{"schema":{"type":"string"},"content":{"application/json":{"schema":{"$ref":"#/components/schemas/Foo"}}}}`)
		v         SchemaXORContentNot
	)

	require.NoError(t, json.Unmarshal(jsonValue, &v))

	marshaled, err := json.Marshal(v)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(marshaled, &v))
	assertjson.Equal(t, jsonValue, marshaled)
}

func TestPathParameter_MarshalJSON_roundtrip(t *testing.T) {
	var (
		jsonValue = []byte(`{"in":"path","style":"label","required":true}`)
		v         PathParameter
	)

	require.NoError(t, json.Unmarshal(jsonValue, &v))

	marshaled, err := json.Marshal(v)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(marshaled, &v))
	assertjson.Equal(t, jsonValue, marshaled)
}

func TestQueryParameter_MarshalJSON_roundtrip(t *testing.T) {
	var (
		jsonValue = []byte(`{"in":"query","style":"pipeDelimited"}`)
		v         QueryParameter
	)

	require.NoError(t, json.Unmarshal(jsonValue, &v))

	marshaled, err := json.Marshal(v)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(marshaled, &v))
	assertjson.Equal(t, jsonValue, marshaled)
}

func TestHeaderParameter_MarshalJSON_roundtrip(t *testing.T) {
	var (
		jsonValue = []byte(`{"in":"header","style":"simple"}`)
		v         HeaderParameter
	)

	require.NoError(t, json.Unmarshal(jsonValue, &v))

	marshaled, err := json.Marshal(v)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(marshaled, &v))
	assertjson.Equal(t, jsonValue, marshaled)
}

func TestCookieParameter_MarshalJSON_roundtrip(t *testing.T) {
	var (
		jsonValue = []byte(`{"in":"cookie","style":"form"}`)
		v         CookieParameter
	)

	require.NoError(t, json.Unmarshal(jsonValue, &v))

	marshaled, err := json.Marshal(v)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(marshaled, &v))
	assertjson.Equal(t, jsonValue, marshaled)
}

func TestParameterLocation_MarshalJSON_roundtrip(t *testing.T) {
	var (
		jsonValue = []byte(`{"in":"cookie","style":"form"}`)
		v         ParameterLocation
	)

	require.NoError(t, json.Unmarshal(jsonValue, &v))

	marshaled, err := json.Marshal(v)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(marshaled, &v))
	assertjson.Equal(t, jsonValue, marshaled)
}

func TestParameterOrRef_MarshalJSON_roundtrip(t *testing.T) {
	var (
		jsonValue = []byte(`{"name":"id","in":"path","required":true,"schema":{"type":"string"}}`)
		v         ParameterOrRef
	)

	require.NoError(t, json.Unmarshal(jsonValue, &v))

	marshaled, err := json.Marshal(v)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(marshaled, &v))
	assertjson.Equal(t, jsonValue, marshaled)
}

func TestOperation_MarshalJSON_roundtrip(t *testing.T) {
	var (
		jsonValue = []byte(`{"tags":["edc"],"summary":"aadabb","description":"fbadeb","externalDocs":{"description":"eccf","url":"eeacfa"},"operationId":"ffd","parameters":[{"$ref":"#/components/parameters/Foo"}],"requestBody":{"$ref":"#/components/requestBodies/Foo"},"responses":{"default":{"description":"This is a sample response.","content":{"application/json":{"schema":{"$ref":"#/components/schemas/Foo"}}}}},"callbacks":{"afa":{"$ref":"#/components/callbacks/Foo"}},"deprecated":true,"security":[{"cac":["ddd"]}],"servers":[{"url":"aceba","description":"bc","variables":{"dfa":{"enum":["bcacab"],"default":"ae","description":"feb"}}}]}`)
		v         Operation
	)

	require.NoError(t, json.Unmarshal(jsonValue, &v))

	marshaled, err := json.Marshal(v)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(marshaled, &v))
	assertjson.Equal(t, jsonValue, marshaled)
}

func TestRequestBodyReference_MarshalJSON_roundtrip(t *testing.T) {
	var (
		jsonValue = []byte(`{"$ref":"#/components/requestBodies/Foo"}`)
		v         RequestBodyReference
	)

	require.NoError(t, json.Unmarshal(jsonValue, &v))

	marshaled, err := json.Marshal(v)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(marshaled, &v))
	assertjson.Equal(t, jsonValue, marshaled)
}

func TestRequestBody_MarshalJSON_roundtrip(t *testing.T) {
	var (
		jsonValue = []byte(`{"content":{"multipart/form-data":{"schema":{"$ref":"#/components/schemas/Foo"}}}}`)
		v         RequestBody
	)

	require.NoError(t, json.Unmarshal(jsonValue, &v))

	marshaled, err := json.Marshal(v)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(marshaled, &v))
	assertjson.Equal(t, jsonValue, marshaled)
}

func TestRequestBodyOrRef_MarshalJSON_roundtrip(t *testing.T) {
	var (
		jsonValue = []byte(`{"content":{"multipart/form-data":{"schema":{"$ref":"#/components/schemas/Foo"}}}}`)
		v         RequestBodyOrRef
	)

	require.NoError(t, json.Unmarshal(jsonValue, &v))

	marshaled, err := json.Marshal(v)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(marshaled, &v))
	assertjson.Equal(t, jsonValue, marshaled)
}

func TestResponses_MarshalJSON_roundtrip(t *testing.T) {
	var (
		jsonValue = []byte(`{"default":{"description":"This is a sample response.","content":{"application/json":{"schema":{"$ref":"#/components/schemas/Foo"}}}}}`)
		v         Responses
	)

	require.NoError(t, json.Unmarshal(jsonValue, &v))

	marshaled, err := json.Marshal(v)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(marshaled, &v))
	assertjson.Equal(t, jsonValue, marshaled)
}

func TestResponseReference_MarshalJSON_roundtrip(t *testing.T) {
	var (
		jsonValue = []byte(`{"$ref":"#/components/responses/Foo"}`)
		v         ResponseReference
	)

	require.NoError(t, json.Unmarshal(jsonValue, &v))

	marshaled, err := json.Marshal(v)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(marshaled, &v))
	assertjson.Equal(t, jsonValue, marshaled)
}

func TestResponse_MarshalJSON_roundtrip(t *testing.T) {
	var (
		jsonValue = []byte(`{"description":"This is a sample response.","content":{"application/json":{"schema":{"$ref":"#/components/schemas/Foo"}}}}`)
		v         Response
	)

	require.NoError(t, json.Unmarshal(jsonValue, &v))

	marshaled, err := json.Marshal(v)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(marshaled, &v))
	assertjson.Equal(t, jsonValue, marshaled)
}

func TestHeaderReference_MarshalJSON_roundtrip(t *testing.T) {
	var (
		jsonValue = []byte(`{"$ref":"#/components/headers/Foo"}`)
		v         HeaderReference
	)

	require.NoError(t, json.Unmarshal(jsonValue, &v))

	marshaled, err := json.Marshal(v)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(marshaled, &v))
	assertjson.Equal(t, jsonValue, marshaled)
}

func TestHeaderOrRef_MarshalJSON_roundtrip(t *testing.T) {
	var (
		jsonValue = []byte(`{"style":"simple","schema":{"type":"string","description":"Sample header response."}}`)
		v         HeaderOrRef
	)

	require.NoError(t, json.Unmarshal(jsonValue, &v))

	marshaled, err := json.Marshal(v)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(marshaled, &v))
	assertjson.Equal(t, jsonValue, marshaled)
}

func TestLinkReference_MarshalJSON_roundtrip(t *testing.T) {
	var (
		jsonValue = []byte(`{"$ref":"#/components/links/Foo"}`)
		v         LinkReference
	)

	require.NoError(t, json.Unmarshal(jsonValue, &v))

	marshaled, err := json.Marshal(v)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(marshaled, &v))
	assertjson.Equal(t, jsonValue, marshaled)
}

func TestLink_MarshalJSON_roundtrip(t *testing.T) {
	var (
		jsonValue = []byte(`{"operationId":"foo"}`)
		v         Link
	)

	require.NoError(t, json.Unmarshal(jsonValue, &v))

	marshaled, err := json.Marshal(v)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(marshaled, &v))
	assertjson.Equal(t, jsonValue, marshaled)
}

func TestLinkNot_MarshalJSON_roundtrip(t *testing.T) {
	var (
		jsonValue = []byte(`{"operationId":"foo","operationRef":"bar"}`)
		v         LinkNot
	)

	require.NoError(t, json.Unmarshal(jsonValue, &v))

	marshaled, err := json.Marshal(v)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(marshaled, &v))
	assertjson.Equal(t, jsonValue, marshaled)
}

func TestLinkOrRef_MarshalJSON_roundtrip(t *testing.T) {
	var (
		jsonValue = []byte(`{"$ref":"#/components/links/Foo"}`)
		v         LinkOrRef
	)

	require.NoError(t, json.Unmarshal(jsonValue, &v))

	marshaled, err := json.Marshal(v)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(marshaled, &v))
	assertjson.Equal(t, jsonValue, marshaled)
}

func TestResponseOrRef_MarshalJSON_roundtrip(t *testing.T) {
	var (
		jsonValue = []byte(`{"$ref":"#/components/responses/Foo"}`)
		v         ResponseOrRef
	)

	require.NoError(t, json.Unmarshal(jsonValue, &v))

	marshaled, err := json.Marshal(v)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(marshaled, &v))
	assertjson.Equal(t, jsonValue, marshaled)
}

func TestCallbackReference_MarshalJSON_roundtrip(t *testing.T) {
	var (
		jsonValue = []byte(`{"$ref":"#/components/callbacks/Foo"}`)
		v         CallbackReference
	)

	require.NoError(t, json.Unmarshal(jsonValue, &v))

	marshaled, err := json.Marshal(v)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(marshaled, &v))
	assertjson.Equal(t, jsonValue, marshaled)
}

func TestCallback_MarshalJSON_roundtrip(t *testing.T) {
	var (
		jsonValue = []byte(`{"fd":{"$ref":"acf","summary":"eeea","description":"ccedaf","servers":[{"url":"acaf","description":"cfd","variables":{"ddb":{"enum":["fa"],"default":"ccacec","description":"ece"}}}],"parameters":[{"$ref":"#/components/parameters/Foo"}]}}`)
		v         Callback
	)

	require.NoError(t, json.Unmarshal(jsonValue, &v))

	marshaled, err := json.Marshal(v)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(marshaled, &v))
	assertjson.Equal(t, jsonValue, marshaled)
}

func TestCallbackOrRef_MarshalJSON_roundtrip(t *testing.T) {
	var (
		jsonValue = []byte(`{"$ref":"#/components/callbacks/Foo"}`)
		v         CallbackOrRef
	)

	require.NoError(t, json.Unmarshal(jsonValue, &v))

	marshaled, err := json.Marshal(v)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(marshaled, &v))
	assertjson.Equal(t, jsonValue, marshaled)
}

func TestPaths_MarshalJSON_roundtrip(t *testing.T) {
	var (
		jsonValue = []byte(`{}`)
		v         Paths
	)

	require.NoError(t, json.Unmarshal(jsonValue, &v))

	marshaled, err := json.Marshal(v)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(marshaled, &v))
	assertjson.Equal(t, jsonValue, marshaled)
}

func TestComponents_MarshalJSON_roundtrip(t *testing.T) {
	var (
		jsonValue = []byte(`{"schemas":{},"responses":{},"parameters":{},"examples":{},"requestBodies":{},"headers":{},"securitySchemes":{},"links":{},"callbacks":{}}`)
		v         Components
	)

	require.NoError(t, json.Unmarshal(jsonValue, &v))

	marshaled, err := json.Marshal(v)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(marshaled, &v))
	assertjson.Equal(t, jsonValue, marshaled)
}

func TestComponentsSchemas_MarshalJSON_roundtrip(t *testing.T) {
	var (
		jsonValue = []byte(`{}`)
		v         ComponentsSchemas
	)

	require.NoError(t, json.Unmarshal(jsonValue, &v))

	marshaled, err := json.Marshal(v)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(marshaled, &v))
	assertjson.Equal(t, jsonValue, marshaled)
}

func TestComponentsResponses_MarshalJSON_roundtrip(t *testing.T) {
	var (
		jsonValue = []byte(`{}`)
		v         ComponentsResponses
	)

	require.NoError(t, json.Unmarshal(jsonValue, &v))

	marshaled, err := json.Marshal(v)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(marshaled, &v))
	assertjson.Equal(t, jsonValue, marshaled)
}

func TestComponentsParameters_MarshalJSON_roundtrip(t *testing.T) {
	var (
		jsonValue = []byte(`{}`)
		v         ComponentsParameters
	)

	require.NoError(t, json.Unmarshal(jsonValue, &v))

	marshaled, err := json.Marshal(v)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(marshaled, &v))
	assertjson.Equal(t, jsonValue, marshaled)
}

func TestComponentsExamples_MarshalJSON_roundtrip(t *testing.T) {
	var (
		jsonValue = []byte(`{}`)
		v         ComponentsExamples
	)

	require.NoError(t, json.Unmarshal(jsonValue, &v))

	marshaled, err := json.Marshal(v)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(marshaled, &v))
	assertjson.Equal(t, jsonValue, marshaled)
}

func TestComponentsRequestBodies_MarshalJSON_roundtrip(t *testing.T) {
	var (
		jsonValue = []byte(`{}`)
		v         ComponentsRequestBodies
	)

	require.NoError(t, json.Unmarshal(jsonValue, &v))

	marshaled, err := json.Marshal(v)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(marshaled, &v))
	assertjson.Equal(t, jsonValue, marshaled)
}

func TestComponentsHeaders_MarshalJSON_roundtrip(t *testing.T) {
	var (
		jsonValue = []byte(`{}`)
		v         ComponentsHeaders
	)

	require.NoError(t, json.Unmarshal(jsonValue, &v))

	marshaled, err := json.Marshal(v)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(marshaled, &v))
	assertjson.Equal(t, jsonValue, marshaled)
}

func TestSecuritySchemeReference_MarshalJSON_roundtrip(t *testing.T) {
	var (
		jsonValue = []byte(`{"$ref":"#/components/securitySchemes/Foo"}`)
		v         SecuritySchemeReference
	)

	require.NoError(t, json.Unmarshal(jsonValue, &v))

	marshaled, err := json.Marshal(v)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(marshaled, &v))
	assertjson.Equal(t, jsonValue, marshaled)
}

func TestAPIKeySecurityScheme_MarshalJSON_roundtrip(t *testing.T) {
	var (
		jsonValue = []byte(`{"type":"apiKey","name":"eadbf","in":"query","description":"fac"}`)
		v         APIKeySecurityScheme
	)

	require.NoError(t, json.Unmarshal(jsonValue, &v))

	marshaled, err := json.Marshal(v)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(marshaled, &v))
	assertjson.Equal(t, jsonValue, marshaled)
}

func TestHTTPSecurityScheme_MarshalJSON_roundtrip(t *testing.T) {
	var (
		jsonValue = []byte(`{"scheme":"bearer","bearerFormat":"cffaee","description":"dcddce","type":"http"}`)
		v         HTTPSecurityScheme
	)

	require.NoError(t, json.Unmarshal(jsonValue, &v))

	marshaled, err := json.Marshal(v)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(marshaled, &v))
	assertjson.Equal(t, jsonValue, marshaled)
}

func TestBearer_MarshalJSON_roundtrip(t *testing.T) {
	var (
		jsonValue = []byte(`{"scheme":"bearer"}`)
		v         Bearer
	)

	require.NoError(t, json.Unmarshal(jsonValue, &v))

	marshaled, err := json.Marshal(v)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(marshaled, &v))
	assertjson.Equal(t, jsonValue, marshaled)
}

func TestNonBearer_MarshalJSON_roundtrip(t *testing.T) {
	var (
		jsonValue = []byte(`{}`)
		v         NonBearer
	)

	require.NoError(t, json.Unmarshal(jsonValue, &v))

	marshaled, err := json.Marshal(v)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(marshaled, &v))
	assertjson.Equal(t, jsonValue, marshaled)
}

func TestOAuth2SecurityScheme_MarshalJSON_roundtrip(t *testing.T) {
	var (
		jsonValue = []byte(`{"type":"oauth2","flows":{"implicit":{"authorizationUrl":"eb","refreshUrl":"ccec","scopes":{"dbcc":"ecabba"}},"password":{"tokenUrl":"ccfc","refreshUrl":"dfce","scopes":{"dbbfd":"fabcce"}},"clientCredentials":{"tokenUrl":"caee","refreshUrl":"cbed","scopes":{"edbe":"bddcd"}},"authorizationCode":{"authorizationUrl":"eb","tokenUrl":"bbab","refreshUrl":"ddf","scopes":{"ebdb":"cbe"}}},"description":"becfee"}`)
		v         OAuth2SecurityScheme
	)

	require.NoError(t, json.Unmarshal(jsonValue, &v))

	marshaled, err := json.Marshal(v)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(marshaled, &v))
	assertjson.Equal(t, jsonValue, marshaled)
}

func TestOAuthFlows_MarshalJSON_roundtrip(t *testing.T) {
	var (
		jsonValue = []byte(`{"implicit":{"authorizationUrl":"aaab","refreshUrl":"fe","scopes":{"da":"ce"}},"password":{"tokenUrl":"bfcfa","refreshUrl":"af","scopes":{"bfe":"cf"}},"clientCredentials":{"tokenUrl":"accb","refreshUrl":"aa","scopes":{"dfba":"ab"}},"authorizationCode":{"authorizationUrl":"ac","tokenUrl":"faceb","refreshUrl":"cffd","scopes":{"bbeffa":"bfd"}}}`)
		v         OAuthFlows
	)

	require.NoError(t, json.Unmarshal(jsonValue, &v))

	marshaled, err := json.Marshal(v)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(marshaled, &v))
	assertjson.Equal(t, jsonValue, marshaled)
}

func TestImplicitOAuthFlow_MarshalJSON_roundtrip(t *testing.T) {
	var (
		jsonValue = []byte(`{"authorizationUrl":"dde","refreshUrl":"ecf","scopes":{"fbffc":"ffde"}}`)
		v         ImplicitOAuthFlow
	)

	require.NoError(t, json.Unmarshal(jsonValue, &v))

	marshaled, err := json.Marshal(v)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(marshaled, &v))
	assertjson.Equal(t, jsonValue, marshaled)
}

func TestPasswordOAuthFlow_MarshalJSON_roundtrip(t *testing.T) {
	var (
		jsonValue = []byte(`{"tokenUrl":"df","refreshUrl":"aeedb","scopes":{"ccafeb":"eea"}}`)
		v         PasswordOAuthFlow
	)

	require.NoError(t, json.Unmarshal(jsonValue, &v))

	marshaled, err := json.Marshal(v)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(marshaled, &v))
	assertjson.Equal(t, jsonValue, marshaled)
}

func TestClientCredentialsFlow_MarshalJSON_roundtrip(t *testing.T) {
	var (
		jsonValue = []byte(`{"tokenUrl":"efaabe","refreshUrl":"cbe","scopes":{"caa":"bf"}}`)
		v         ClientCredentialsFlow
	)

	require.NoError(t, json.Unmarshal(jsonValue, &v))

	marshaled, err := json.Marshal(v)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(marshaled, &v))
	assertjson.Equal(t, jsonValue, marshaled)
}

func TestAuthorizationCodeOAuthFlow_MarshalJSON_roundtrip(t *testing.T) {
	var (
		jsonValue = []byte(`{"authorizationUrl":"aacdcd","tokenUrl":"bdc","refreshUrl":"ebaebb","scopes":{"bc":"daaf"}}`)
		v         AuthorizationCodeOAuthFlow
	)

	require.NoError(t, json.Unmarshal(jsonValue, &v))

	marshaled, err := json.Marshal(v)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(marshaled, &v))
	assertjson.Equal(t, jsonValue, marshaled)
}

func TestOpenIDConnectSecurityScheme_MarshalJSON_roundtrip(t *testing.T) {
	var (
		jsonValue = []byte(`{"type":"openIdConnect","openIdConnectUrl":"eb","description":"bfddfd"}`)
		v         OpenIDConnectSecurityScheme
	)

	require.NoError(t, json.Unmarshal(jsonValue, &v))

	marshaled, err := json.Marshal(v)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(marshaled, &v))
	assertjson.Equal(t, jsonValue, marshaled)
}

func TestSecurityScheme_MarshalJSON_roundtrip(t *testing.T) {
	var (
		jsonValue = []byte(`{"type":"http","scheme":"basic","description":"Admin access"}`)
		v         SecurityScheme
	)

	require.NoError(t, json.Unmarshal(jsonValue, &v))

	marshaled, err := json.Marshal(v)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(marshaled, &v))
	assertjson.Equal(t, jsonValue, marshaled)
}

func TestSecuritySchemeOrRef_MarshalJSON_roundtrip(t *testing.T) {
	var (
		jsonValue = []byte(`{"$ref":"#/components/securitySchemes/Foo"}`)
		v         SecuritySchemeOrRef
	)

	require.NoError(t, json.Unmarshal(jsonValue, &v))

	marshaled, err := json.Marshal(v)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(marshaled, &v))
	assertjson.Equal(t, jsonValue, marshaled)
}

func TestComponentsSecuritySchemes_MarshalJSON_roundtrip(t *testing.T) {
	var (
		jsonValue = []byte(`{}`)
		v         ComponentsSecuritySchemes
	)

	require.NoError(t, json.Unmarshal(jsonValue, &v))

	marshaled, err := json.Marshal(v)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(marshaled, &v))
	assertjson.Equal(t, jsonValue, marshaled)
}

func TestComponentsLinks_MarshalJSON_roundtrip(t *testing.T) {
	var (
		jsonValue = []byte(`{}`)
		v         ComponentsLinks
	)

	require.NoError(t, json.Unmarshal(jsonValue, &v))

	marshaled, err := json.Marshal(v)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(marshaled, &v))
	assertjson.Equal(t, jsonValue, marshaled)
}

func TestComponentsCallbacks_MarshalJSON_roundtrip(t *testing.T) {
	var (
		jsonValue = []byte(`{}`)
		v         ComponentsCallbacks
	)

	require.NoError(t, json.Unmarshal(jsonValue, &v))

	marshaled, err := json.Marshal(v)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(marshaled, &v))
	assertjson.Equal(t, jsonValue, marshaled)
}

func TestSpec_UnmarshalJSON_openai(t *testing.T) {
	var s Spec

	j, err := os.ReadFile("testdata/openai-openapi.json")
	require.NoError(t, err)

	// "oneOf constraint failed for SchemaOrRef with 0 valid results:
	//   map[Schema:oneOf constraint failed for SchemaOrRef with 0 valid results:
	//    map[Schema:additional properties not allowed in Schema:
	//     [$ref] SchemaReference:additional properties not allowed in SchemaReference:
	//      [description]] SchemaReference:required key missing: $ref]"
	require.Error(t, s.UnmarshalJSON(j))

	j, err = os.ReadFile("testdata/openai-openapi-fixed.json")
	require.NoError(t, err)

	require.NoError(t, s.UnmarshalJSON(j))
}
