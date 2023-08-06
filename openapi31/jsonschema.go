package openapi31

import "github.com/swaggest/jsonschema-go"

func isDeprecated(schema jsonschema.SchemaOrBool) *bool {
	if schema.TypeObject == nil {
		return nil
	}

	if d, ok := schema.TypeObject.ExtraProperties["deprecated"].(bool); ok {
		return &d
	}

	return nil
}
