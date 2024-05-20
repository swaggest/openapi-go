package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"reflect"
	"regexp"
	"strings"

	"github.com/swaggest/jsonschema-go"
	"github.com/swaggest/openapi-go"
	"github.com/swaggest/refl"
)

const (
	tagJSON     = "json"
	tagFormData = "formData"
	tagForm     = "form"
	tagHeader   = "header"

	componentsSchemas = "#/components/schemas/"
)

var defNameSanitizer = regexp.MustCompile(`[^a-zA-Z0-9.\-_]+`)

func sanitizeDefName(rc *jsonschema.ReflectContext) {
	jsonschema.InterceptDefName(func(_ reflect.Type, defaultDefName string) string {
		return defNameSanitizer.ReplaceAllString(defaultDefName, "")
	})(rc)
}

// ReflectRequestBody reflects JSON schema of request body.
func ReflectRequestBody(
	is31 bool, // True if OpenAPI 3.1
	r *jsonschema.Reflector,
	cu openapi.ContentUnit,
	httpMethod string,
	mapping map[string]string,
	tag string,
	additionalTags []string,
	reflOptions ...func(rc *jsonschema.ReflectContext),
) (schema *jsonschema.Schema, hasFileUpload bool, err error) {
	input := cu.Structure

	httpMethod = strings.ToUpper(httpMethod)
	_, forceRequestBody := input.(openapi.RequestBodyEnforcer)
	_, forceJSONRequestBody := input.(openapi.RequestJSONBodyEnforcer)

	// GET, HEAD, DELETE and TRACE requests should not have body.
	switch httpMethod {
	case http.MethodGet, http.MethodHead, http.MethodDelete, http.MethodTrace:
		if !forceRequestBody {
			return nil, false, nil
		}
	}

	hasTaggedFields := refl.HasTaggedFields(input, tag)
	for _, t := range additionalTags {
		if hasTaggedFields {
			break
		}

		hasTaggedFields = refl.HasTaggedFields(input, t)
	}

	hasJSONSchemaStruct := false

	refl.WalkFieldsRecursively(reflect.ValueOf(input), func(v reflect.Value, _ reflect.StructField, _ []reflect.StructField) {
		if v.Type() == reflect.TypeOf(jsonschema.Struct{}) {
			hasJSONSchemaStruct = true
		}
	})

	// Form data can not have map or array as body.
	if !hasTaggedFields && len(mapping) == 0 && tag != tagJSON {
		return nil, false, nil
	}

	// If `formData` is defined on a request body `json` is ignored.
	if tag == tagJSON &&
		(refl.HasTaggedFields(input, tagFormData) || refl.HasTaggedFields(input, tagForm)) &&
		!forceJSONRequestBody && !hasJSONSchemaStruct {
		return nil, false, nil
	}

	// Checking for default options that allow tag-less JSON.
	isProcessWithoutTags := false

	_, err = r.Reflect("", func(rc *jsonschema.ReflectContext) {
		isProcessWithoutTags = rc.ProcessWithoutTags
	})
	if err != nil {
		return nil, false, fmt.Errorf("BUG: %w", err)
	}

	// JSON can be a map or array without field tags.
	if !hasTaggedFields && !hasJSONSchemaStruct && len(mapping) == 0 && !refl.IsSliceOrMap(input) &&
		refl.FindEmbeddedSliceOrMap(input) == nil && !isProcessWithoutTags {
		return nil, false, nil
	}

	definitionPrefix := ""

	if tag != tagJSON {
		definitionPrefix += strings.Title(tag)
	}

	reflOptions = append(reflOptions,
		jsonschema.InterceptDefName(func(t reflect.Type, defaultDefName string) string {
			if tag != tagJSON {
				v := reflect.New(t).Interface()

				if refl.HasTaggedFields(v, tag) {
					return definitionPrefix + defaultDefName
				}

				for _, at := range additionalTags {
					if refl.HasTaggedFields(v, at) {
						return definitionPrefix + defaultDefName
					}
				}
			}

			return defaultDefName
		}),
		jsonschema.RootRef,
		jsonschema.PropertyNameMapping(mapping),
		jsonschema.PropertyNameTag(tag, additionalTags...),
		sanitizeDefName,
		jsonschema.InterceptNullability(func(params jsonschema.InterceptNullabilityParams) {
			if params.NullAdded {
				if params.Schema.ReflectType == nil {
					return
				}

				vv := reflect.Zero(params.Schema.ReflectType).Interface()

				foundFiles := false
				if _, ok := vv.([]multipart.File); ok {
					foundFiles = true
				}

				if _, ok := vv.([]*multipart.FileHeader); ok {
					foundFiles = true
				}

				if foundFiles {
					params.Schema.RemoveType(jsonschema.Null)
				}
			}
		}),
		jsonschema.InterceptSchema(func(params jsonschema.InterceptSchemaParams) (stop bool, err error) {
			vv := params.Value.Interface()

			foundFile := false
			if _, ok := vv.(*multipart.File); ok {
				foundFile = true
			}

			if _, ok := vv.(*multipart.FileHeader); ok {
				foundFile = true
			}

			if foundFile {
				params.Schema.AddType(jsonschema.String)
				params.Schema.RemoveType(jsonschema.Null)
				params.Schema.WithFormat("binary")

				if is31 {
					params.Schema.WithExtraPropertiesItem("contentMediaType", "application/octet-stream")
				}

				hasFileUpload = true

				return true, nil
			}

			return false, nil
		}),
	)

	sch, err := r.Reflect(input, reflOptions...)
	if err != nil {
		return nil, false, err
	}

	return &sch, hasFileUpload, nil
}

// ReflectJSONResponse reflects JSON schema of response.
func ReflectJSONResponse(
	r *jsonschema.Reflector,
	output interface{},
	reflOptions ...func(rc *jsonschema.ReflectContext),
) (schema *jsonschema.Schema, err error) {
	if output == nil {
		return nil, nil
	}

	// Check if output structure exposes meaningful schema.
	if hasJSONBody, err := hasJSONBody(r, output); err == nil && !hasJSONBody {
		return nil, nil
	}

	reflOptions = append(reflOptions,
		jsonschema.RootRef,
		sanitizeDefName,
	)

	sch, err := r.Reflect(output, reflOptions...)
	if err != nil {
		return nil, err
	}

	return &sch, nil
}

func hasJSONBody(r *jsonschema.Reflector, output interface{}) (bool, error) {
	schema, err := r.Reflect(output, sanitizeDefName)
	if err != nil {
		return false, err
	}

	// Remove non-constraining fields to prepare for marshaling.
	schema.Title = nil
	schema.Description = nil
	schema.Comment = nil
	schema.ExtraProperties = nil
	schema.ID = nil
	schema.Examples = nil

	j, err := json.Marshal(schema)
	if err != nil {
		return false, err
	}

	if !bytes.Equal([]byte("{}"), j) && !bytes.Equal([]byte(`{"type":"object"}`), j) {
		return true, nil
	}

	return false, nil
}

// ReflectResponseHeader reflects response headers from content unit.
func ReflectResponseHeader(
	r *jsonschema.Reflector,
	oc openapi.OperationContext,
	cu openapi.ContentUnit,
	interceptProp jsonschema.InterceptPropFunc,
) (jsonschema.Schema, error) {
	output := cu.Structure
	mapping := cu.FieldMapping(openapi.InHeader)

	if output == nil {
		return jsonschema.Schema{}, nil
	}

	return r.Reflect(output,
		func(rc *jsonschema.ReflectContext) {
			rc.ProcessWithoutTags = false
		},
		openapi.WithOperationCtx(oc, true, openapi.InHeader),
		jsonschema.InlineRefs,
		jsonschema.PropertyNameMapping(mapping),
		jsonschema.PropertyNameTag(tagHeader),
		sanitizeDefName,
		jsonschema.InterceptProp(interceptProp),
	)
}

// ReflectParametersIn reflects JSON schema of request parameters.
func ReflectParametersIn(
	r *jsonschema.Reflector,
	oc openapi.OperationContext,
	c openapi.ContentUnit,
	in openapi.In,
	collectDefinitions func(name string, schema jsonschema.Schema),
	interceptProp jsonschema.InterceptPropFunc,
	additionalTags ...string,
) (jsonschema.Schema, error) {
	input := c.Structure
	propertyMapping := c.FieldMapping(in)

	if refl.IsSliceOrMap(input) {
		return jsonschema.Schema{}, nil
	}

	return r.Reflect(input,
		func(rc *jsonschema.ReflectContext) {
			rc.ProcessWithoutTags = false
		},
		openapi.WithOperationCtx(oc, false, in),
		jsonschema.DefinitionsPrefix(componentsSchemas),
		jsonschema.CollectDefinitions(collectDefinitions),
		jsonschema.PropertyNameMapping(propertyMapping),
		jsonschema.PropertyNameTag(string(in), additionalTags...),
		func(rc *jsonschema.ReflectContext) {
			rc.UnnamedFieldWithTag = true
		},
		sanitizeDefName,
		jsonschema.SkipEmbeddedMapsSlices,
		jsonschema.InterceptProp(interceptProp),
	)
}
