package openapi3

import (
	"strings"

	"github.com/swaggest/jsonschema-go"
)

type toJSONSchemaContext struct {
	refsProcessed map[string]jsonschema.SchemaOrBool
	refsCount     map[string]int
	spec          *Spec
}

// ToJSONSchema converts OpenAPI Schema to JSON Schema.
//
// Local references are resolved against `#/components/schemas` in spec.
func (s *SchemaOrRef) ToJSONSchema(spec *Spec) jsonschema.SchemaOrBool {
	ctx := toJSONSchemaContext{
		refsProcessed: map[string]jsonschema.SchemaOrBool{},
		refsCount:     map[string]int{},
		spec:          spec,
	}
	js := s.toJSONSchema(ctx)

	// Inline root reference without recursions.
	if s.SchemaReference != nil {
		dstName := strings.TrimPrefix(s.SchemaReference.Ref, componentsSchemas)
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

func (s *SchemaOrRef) toJSONSchema(ctx toJSONSchemaContext) jsonschema.SchemaOrBool {
	js := jsonschema.SchemaOrBool{}
	jso := js.TypeObjectEns()

	if s.SchemaReference != nil {
		jso.WithRef(s.SchemaReference.Ref)

		if strings.HasPrefix(s.SchemaReference.Ref, componentsSchemas) {
			dstName := strings.TrimPrefix(s.SchemaReference.Ref, componentsSchemas)

			if _, alreadyProcessed := ctx.refsProcessed[dstName]; !alreadyProcessed {
				ctx.refsProcessed[dstName] = jsonschema.SchemaOrBool{}

				dst := ctx.spec.Components.Schemas.MapOfSchemaOrRefValues[dstName]
				ctx.refsProcessed[dstName] = dst.toJSONSchema(ctx)
			}

			ctx.refsCount[dstName]++
		}

		return js
	}

	if s.Schema == nil {
		return js
	}

	ss := s.Schema

	jso.ReflectType = ss.ReflectType
	jso.Description = ss.Description
	jso.Title = ss.Title

	if ss.Type != nil {
		jso.AddType(jsonschema.SimpleType(*ss.Type))
	}

	if ss.Nullable != nil && *ss.Nullable {
		jso.AddType(jsonschema.Null)
	}

	jso.MultipleOf = ss.MultipleOf
	if ss.ExclusiveMaximum != nil && *ss.ExclusiveMaximum {
		jso.ExclusiveMaximum = ss.Maximum
	} else {
		jso.Maximum = ss.Maximum
	}

	if ss.ExclusiveMinimum != nil && *ss.ExclusiveMinimum {
		jso.ExclusiveMinimum = ss.Minimum
	} else {
		jso.Minimum = ss.Minimum
	}

	jso.MaxLength = ss.MaxLength

	if ss.MinLength != nil {
		jso.MinLength = *ss.MinLength
	}

	jso.Pattern = ss.Pattern
	jso.MaxItems = ss.MaxItems

	if ss.MinItems != nil {
		jso.MinItems = *ss.MinItems
	}

	jso.UniqueItems = ss.UniqueItems
	jso.MaxProperties = ss.MaxProperties

	if ss.MinProperties != nil {
		jso.MinProperties = *ss.MinProperties
	}

	jso.Required = ss.Required
	jso.Enum = ss.Enum

	if ss.Not != nil {
		jso.WithNot(ss.Not.toJSONSchema(ctx))
	}

	if len(ss.AllOf) != 0 {
		for _, allOf := range ss.AllOf {
			jso.AllOf = append(jso.AllOf, allOf.toJSONSchema(ctx))
		}
	}

	if len(ss.OneOf) != 0 {
		for _, oneOf := range ss.OneOf {
			jso.OneOf = append(jso.OneOf, oneOf.toJSONSchema(ctx))
		}
	}

	if len(ss.AnyOf) != 0 {
		for _, anyOf := range ss.AnyOf {
			jso.AnyOf = append(jso.AnyOf, anyOf.toJSONSchema(ctx))
		}
	}

	if ss.Items != nil {
		jso.ItemsEns().WithSchemaOrBool(ss.Items.toJSONSchema(ctx))
	}

	if ss.Properties != nil {
		for propName, propSchema := range ss.Properties {
			jso.WithPropertiesItem(propName, propSchema.toJSONSchema(ctx))
		}
	}

	if ss.AdditionalProperties != nil {
		if ss.AdditionalProperties.Bool != nil {
			jso.AdditionalProperties = &jsonschema.SchemaOrBool{
				TypeBoolean: ss.AdditionalProperties.Bool,
			}
		} else if ss.AdditionalProperties.SchemaOrRef != nil {
			jso.WithAdditionalProperties(ss.AdditionalProperties.SchemaOrRef.toJSONSchema(ctx))
		}
	}

	jso.Format = ss.Format
	jso.Default = ss.Default
	jso.ReadOnly = ss.ReadOnly

	if ss.Example != nil {
		jso.WithExamples(*ss.Example)
	}

	for k, v := range ss.MapOfAnything {
		jso.WithExtraPropertiesItem(k, v)
	}

	return js
}

// FromJSONSchema loads OpenAPI Schema from JSON Schema.
func (s *SchemaOrRef) FromJSONSchema(schema jsonschema.SchemaOrBool) {
	if schema.TypeBoolean != nil {
		s.fromBool(*schema.TypeBoolean)

		return
	}

	js := schema.TypeObject
	if js.Ref != nil {
		if deprecated, ok := js.ExtraProperties["deprecated"].(bool); ok && deprecated {
			s.Schema = (&Schema{}).WithAllOf(
				SchemaOrRef{
					Schema: (&Schema{}).WithDeprecated(true),
				},
				SchemaOrRef{
					SchemaReference: &SchemaReference{Ref: *js.Ref},
				},
			)

			return
		}

		s.SchemaReference = &SchemaReference{Ref: *js.Ref}

		return
	}

	if s.Schema == nil {
		s.Schema = &Schema{}
	}

	s.Schema.ReflectType = js.ReflectType
	os := s.Schema

	if js.Not != nil {
		os.Not = &SchemaOrRef{}
		os.Not.FromJSONSchema(*js.Not)
	}

	fromSchemaArray(&os.OneOf, js.OneOf)
	fromSchemaArray(&os.AnyOf, js.AnyOf)
	fromSchemaArray(&os.AllOf, js.AllOf)

	os.Title = js.Title
	os.Description = js.Description
	os.Required = js.Required
	os.Default = js.Default
	os.Enum = js.Enum

	if len(js.Examples) > 0 {
		os.Example = &js.Examples[0]
	}

	if deprecated, ok := js.ExtraProperties["deprecated"].(bool); ok {
		os.Deprecated = &deprecated
	}

	if js.Type != nil {
		if js.Type.SimpleTypes != nil {
			checkNullable(*js.Type.SimpleTypes, os)
		} else if len(js.Type.SliceOfSimpleTypeValues) > 0 {
			for _, t := range js.Type.SliceOfSimpleTypeValues {
				checkNullable(t, os)
			}
		}
	}

	if js.AdditionalProperties != nil {
		os.AdditionalProperties = &SchemaAdditionalProperties{}
		if js.AdditionalProperties.TypeBoolean != nil {
			os.AdditionalProperties.Bool = js.AdditionalProperties.TypeBoolean
		} else {
			ap := SchemaOrRef{}
			ap.FromJSONSchema(*js.AdditionalProperties)
			os.AdditionalProperties.SchemaOrRef = &ap
		}
	}

	if js.Items != nil && js.Items.SchemaOrBool != nil {
		os.Items = &SchemaOrRef{}
		os.Items.FromJSONSchema(*js.Items.SchemaOrBool)
	}

	if js.ExclusiveMaximum != nil {
		os.WithExclusiveMaximum(true)
		os.Maximum = js.Maximum
	} else if js.Maximum != nil {
		os.Maximum = js.Maximum
	}

	if js.ExclusiveMinimum != nil {
		os.WithExclusiveMinimum(true)
		os.Minimum = js.Minimum
	} else if js.Minimum != nil {
		os.Minimum = js.Minimum
	}

	os.Format = js.Format

	if js.MinItems != 0 {
		os.MinItems = &js.MinItems
	}

	os.MaxItems = js.MaxItems

	if js.MinLength != 0 {
		os.MinLength = &js.MinLength
	}

	os.MaxLength = js.MaxLength

	if js.MinProperties != 0 {
		os.MinProperties = &js.MinProperties
	}

	os.MaxProperties = js.MaxProperties

	os.MultipleOf = js.MultipleOf

	os.Pattern = js.Pattern

	if len(js.Properties) > 0 {
		os.Properties = make(map[string]SchemaOrRef, len(js.Properties))

		for name, jsp := range js.Properties {
			osp := SchemaOrRef{}
			osp.FromJSONSchema(jsp)
			os.Properties[name] = osp
		}
	}

	os.ReadOnly = js.ReadOnly
	os.UniqueItems = js.UniqueItems

	if writeOnly, ok := js.ExtraProperties["writeOnly"].(bool); ok {
		os.WriteOnly = &writeOnly
	}

	for name, val := range js.ExtraProperties {
		if strings.HasPrefix(name, "x-") {
			if os.MapOfAnything == nil {
				os.MapOfAnything = map[string]interface{}{
					name: val,
				}
			} else {
				os.MapOfAnything[name] = val
			}
		}
	}
}

func checkNullable(t jsonschema.SimpleType, os *Schema) {
	if t == jsonschema.Null {
		os.WithNullable(true)
	} else {
		os.WithType(SchemaType(t))
	}
}

func fromSchemaArray(os *[]SchemaOrRef, js []jsonschema.SchemaOrBool) {
	if len(js) == 0 {
		return
	}

	osa := make([]SchemaOrRef, len(js))

	for i, jso := range js {
		oso := SchemaOrRef{}
		oso.FromJSONSchema(jso)
		osa[i] = oso
	}

	*os = osa
}

func (s *SchemaOrRef) fromBool(val bool) {
	if s.Schema == nil {
		s.Schema = &Schema{}
	}

	if val {
		return
	}

	s.Schema.Not = &SchemaOrRef{
		Schema: &Schema{},
	}
}
