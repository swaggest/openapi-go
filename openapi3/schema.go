package openapi3

import (
	"strings"

	"github.com/swaggest/jsonschema-go"
)

func (s *SchemaOrRef) FromJSONSchema(schema jsonschema.SchemaOrBool) {
	if schema.TypeBoolean != nil {
		s.fromBool(*schema.TypeBoolean)
		return
	}

	js := schema.TypeObject
	if js.Ref != nil {
		s.SchemaReference = &SchemaReference{Ref: *js.Ref}
		return
	}

	if s.Schema == nil {
		s.Schema = &Schema{}
	}

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
			if *js.Type.SimpleTypes == jsonschema.Null {
				os.WithNullable(true)
			} else {
				os.WithType(SchemaType(*js.Type.SimpleTypes))
			}
		} else if len(js.Type.SliceOfSimpleTypeValues) > 0 {
			for _, t := range js.Type.SliceOfSimpleTypeValues {
				if t == jsonschema.Null {
					os.WithNullable(true)
				} else {
					os.WithType(SchemaType(t))
				}
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
