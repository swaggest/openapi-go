package openapi31

import (
	"strings"

	"github.com/swaggest/jsonschema-go"
)

func isDeprecated(schema jsonschema.SchemaOrBool) *bool {
	if schema.TypeObject == nil {
		return nil
	}

	if d, ok := schema.TypeObject.ExtraProperties["deprecated"].(bool); ok {
		return &d
	}

	return nil
}

type toJSONSchemaContext struct {
	refsProcessed map[string]jsonschema.SchemaOrBool
	refsCount     map[string]int
	spec          *Spec
}

// ToJSONSchema converts OpenAPI Schema to JSON Schema.
//
// Local references are resolved against `#/components/schemas` in spec.
func ToJSONSchema(s map[string]interface{}, spec *Spec) jsonschema.SchemaOrBool {
	js := jsonschema.SchemaOrBool{}

	if err := js.FromSimpleMap(s); err != nil {
		panic(err.Error())
	}

	ctx := toJSONSchemaContext{
		refsProcessed: map[string]jsonschema.SchemaOrBool{},
		refsCount:     map[string]int{},
		spec:          spec,
	}

	findReferences(js, ctx)

	// Inline root reference without recursions.
	if js.TypeObjectEns().Ref != nil {
		dstName := strings.TrimPrefix(*js.TypeObjectEns().Ref, componentsSchemas)
		if ctx.refsCount[dstName] == 1 {
			js = ctx.refsProcessed[dstName]
			delete(ctx.refsProcessed, dstName)
		}
	}

	if len(ctx.refsProcessed) > 0 {
		js.TypeObjectEns().WithExtraPropertiesItem(
			"components",
			map[string]interface{}{"schemas": ctx.refsProcessed},
		)
	}

	return js
}

func findReferences(js jsonschema.SchemaOrBool, ctx toJSONSchemaContext) {
	if js.TypeBoolean != nil {
		return
	}

	jso := js.TypeObjectEns()

	if jso.Ref != nil { //nolint:nestif
		if strings.HasPrefix(*jso.Ref, componentsSchemas) {
			dstName := strings.TrimPrefix(*jso.Ref, componentsSchemas)

			if _, alreadyProcessed := ctx.refsProcessed[dstName]; !alreadyProcessed {
				ctx.refsProcessed[dstName] = jsonschema.SchemaOrBool{}

				dst := ctx.spec.Components.Schemas[dstName]

				js := jsonschema.SchemaOrBool{}
				if err := js.FromSimpleMap(dst); err != nil {
					panic("BUG: " + err.Error())
				}

				ctx.refsProcessed[dstName] = js
				findReferences(js, ctx)
			}

			ctx.refsCount[dstName]++
		}

		return
	}

	if jso.Not != nil {
		findReferences(*jso.Not, ctx)
	}

	for _, allOf := range jso.AllOf {
		findReferences(allOf, ctx)
	}

	for _, oneOf := range jso.OneOf {
		findReferences(oneOf, ctx)
	}

	for _, anyOf := range jso.AnyOf {
		findReferences(anyOf, ctx)
	}

	if jso.Items != nil {
		if jso.Items.SchemaOrBool != nil {
			findReferences(*jso.Items.SchemaOrBool, ctx)
		}

		for _, item := range jso.Items.SchemaArray {
			findReferences(item, ctx)
		}
	}

	for _, propSchema := range jso.Properties {
		findReferences(propSchema, ctx)
	}

	if jso.AdditionalProperties != nil {
		findReferences(*jso.AdditionalProperties, ctx)
	}
}
