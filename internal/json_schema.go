package internal

import (
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
)

var defNameSanitizer = regexp.MustCompile(`[^a-zA-Z0-9.\-_]+`)

func sanitizeDefName(rc *jsonschema.ReflectContext) {
	jsonschema.InterceptDefName(func(t reflect.Type, defaultDefName string) string {
		return defNameSanitizer.ReplaceAllString(defaultDefName, "")
	})(rc)
}

// ReflectRequestBody reflects JSON schema of request body.
func ReflectRequestBody(
	r *jsonschema.Reflector,
	reflOption func(rc *jsonschema.ReflectContext),
	cu openapi.ContentUnit,
	httpMethod string,
	mapping map[string]string,
	tag string,
	additionalTags ...string,
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

	// Form data can not have map or array as body.
	if !hasTaggedFields && len(mapping) == 0 && tag != tagJSON {
		return nil, false, nil
	}

	// If `formData` is defined on a request body `json` is ignored.
	if tag == tagJSON &&
		(refl.HasTaggedFields(input, tagFormData) || refl.HasTaggedFields(input, tagForm)) &&
		!forceJSONRequestBody {
		return nil, false, nil
	}

	// JSON can be a map or array without field tags.
	if !hasTaggedFields && len(mapping) == 0 && !refl.IsSliceOrMap(input) && refl.FindEmbeddedSliceOrMap(input) == nil {
		return nil, false, nil
	}

	definitionPrefix := ""

	if tag != tagJSON {
		definitionPrefix += strings.Title(tag)
	}

	if reflOption == nil {
		reflOption = func(rc *jsonschema.ReflectContext) {}
	}

	sch, err := r.Reflect(input,
		reflOption,
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
		jsonschema.InterceptSchema(func(params jsonschema.InterceptSchemaParams) (stop bool, err error) {
			vv := params.Value.Interface()

			found := false
			if _, ok := vv.(*multipart.File); ok {
				found = true
			}

			if _, ok := vv.(*multipart.FileHeader); ok {
				found = true
			}

			if found {
				params.Schema.AddType(jsonschema.String)
				params.Schema.WithFormat("binary")

				hasFileUpload = true

				return true, nil
			}

			return false, nil
		}),
	)
	if err != nil {
		return nil, false, err
	}

	return &sch, hasFileUpload, nil
}
