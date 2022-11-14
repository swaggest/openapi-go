package openapi31

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/swaggest/jsonschema-go"
	"github.com/swaggest/refl"
)

// Reflector builds OpenAPI Schema with reflected structures.
type Reflector struct {
	jsonschema.Reflector
	Spec *Spec
}

// ResolveJSONSchemaRef builds JSON Schema from OpenAPI Component Schema reference.
//
// Can be used in jsonschema.Schema IsTrivial().
func (r Reflector) ResolveJSONSchemaRef(ref string) (s jsonschema.SchemaOrBool, found bool) {
	if r.Spec == nil || r.Spec.Components == nil || r.Spec.Components.Schemas == nil ||
		!strings.HasPrefix(ref, "#/components/schemas/") {
		return s, false
	}

	ref = strings.TrimPrefix(ref, "#/components/schemas/")
	os, found := r.Spec.Components.Schemas[ref]

	if found {
		if err := s.FromSimpleMap(os); err != nil {
			panic(err)
		}
	}

	return s, found
}

// joinErrors joins non-nil errors.
func joinErrors(errs ...error) error {
	join := ""

	for _, err := range errs {
		if err != nil {
			join += ", " + err.Error()
		}
	}

	if join != "" {
		return errors.New(join[2:])
	}

	return nil
}

// SpecEns ensures returned Spec is not nil.
func (r *Reflector) SpecEns() *Spec {
	if r.Spec == nil {
		r.Spec = &Spec{Openapi: "3.1.0"}
	}

	return r.Spec
}

// OperationContext describes operation.
type OperationContext struct {
	Operation  *Operation
	Input      interface{}
	HTTPMethod string

	ReqQueryMapping    map[string]string
	ReqPathMapping     map[string]string
	ReqCookieMapping   map[string]string
	ReqHeaderMapping   map[string]string
	ReqFormDataMapping map[string]string

	Output            interface{}
	HTTPStatus        int
	RespContentType   string
	RespHeaderMapping map[string]string

	ProcessingResponse bool
	ProcessingIn       string
}

// SetupRequest sets up operation parameters.
func (r *Reflector) SetupRequest(oc OperationContext) error {
	return joinErrors(
		r.parseParametersIn(oc, ParameterInQuery, oc.ReqQueryMapping),
		r.parseParametersIn(oc, ParameterInPath, oc.ReqPathMapping),
		r.parseParametersIn(oc, ParameterInCookie, oc.ReqCookieMapping),
		r.parseParametersIn(oc, ParameterInHeader, oc.ReqHeaderMapping),
		r.parseRequestBody(oc, tagJSON, mimeJSON, oc.HTTPMethod, nil),
		r.parseRequestBody(oc, tagFormData, mimeFormUrlencoded, oc.HTTPMethod, oc.ReqFormDataMapping),
	)
}

// SetRequest sets up operation parameters.
func (r *Reflector) SetRequest(o *Operation, input interface{}, httpMethod string) error {
	return r.SetupRequest(OperationContext{
		Operation:  o,
		Input:      input,
		HTTPMethod: httpMethod,
	})
}

const (
	tagJSON            = "json"
	tagFormData        = "formData"
	mimeJSON           = "application/json"
	mimeFormUrlencoded = "application/x-www-form-urlencoded"
	mimeMultipart      = "multipart/form-data"
)

// RequestBodyEnforcer enables request body for GET and HEAD methods.
//
// Should be implemented on input structure, function body can be empty.
// Forcing request body is not recommended and should only be used for backwards compatibility.
type RequestBodyEnforcer interface {
	ForceRequestBody()
}

// RequestJSONBodyEnforcer enables JSON request body for structures with `formData` tags.
//
// Should be implemented on input structure, function body can be empty.
type RequestJSONBodyEnforcer interface {
	ForceJSONRequestBody()
}

func (r *Reflector) parseRequestBody(
	oc OperationContext, tag, mime string, httpMethod string, mapping map[string]string,
) error {
	o := oc.Operation
	input := oc.Input

	httpMethod = strings.ToUpper(httpMethod)
	_, forceRequestBody := input.(RequestBodyEnforcer)
	_, forceJSONRequestBody := input.(RequestJSONBodyEnforcer)

	// GET, HEAD, DELETE and TRACE requests should not have body.
	switch httpMethod {
	case http.MethodGet, http.MethodHead, http.MethodDelete, http.MethodTrace:
		if !forceRequestBody {
			return nil
		}
	}

	hasTaggedFields := refl.HasTaggedFields(input, tag)

	// Form data can not have map or array as body.
	if !hasTaggedFields && len(mapping) == 0 && tag != tagJSON {
		return nil
	}

	// If `formData` is defined on a request body `json` is ignored.
	if tag == tagJSON && refl.HasTaggedFields(input, tagFormData) && !forceJSONRequestBody {
		return nil
	}

	// JSON can be a map or array without field tags.
	if !hasTaggedFields && len(mapping) == 0 && !refl.IsSliceOrMap(input) && refl.FindEmbeddedSliceOrMap(input) == nil {
		return nil
	}

	hasFileUpload := false
	definitionPrefix := ""

	if tag != tagJSON {
		definitionPrefix += strings.Title(tag)
	}

	schema, err := r.Reflect(input,
		r.withOperation(oc, false, "body"),
		jsonschema.DefinitionsPrefix("#/components/schemas/"+definitionPrefix),
		jsonschema.RootRef,
		jsonschema.PropertyNameMapping(mapping),
		jsonschema.PropertyNameTag(tag),
		jsonschema.InterceptType(func(v reflect.Value, s *jsonschema.Schema) (bool, error) {
			vv := v.Interface()

			found := false
			if _, ok := vv.(*multipart.File); ok {
				found = true
			}

			if _, ok := vv.(*multipart.FileHeader); ok {
				found = true
			}

			if found {
				s.AddType(jsonschema.String)
				s.WithFormat("binary")

				hasFileUpload = true

				return true, nil
			}

			return false, nil
		}),
	)
	if err != nil {
		return err
	}

	s, err := schema.ToSchemaOrBool().ToSimpleMap()
	if err != nil {
		return fmt.Errorf("schema to map: %w", err)
	}

	delete(s, "definitions")

	mt := MediaType{
		Schema: s,
	}

	for name, def := range schema.Definitions {
		s, err := def.ToSimpleMap()
		if err != nil {
			return fmt.Errorf("definition %s to map: %w", name, err)
		}

		r.SpecEns().ComponentsEns().WithSchemasItem(definitionPrefix+name, s)
	}

	if mime == mimeFormUrlencoded && hasFileUpload {
		mime = mimeMultipart
	}

	o.RequestBodyEns().RequestBodyEns().WithContentItem(mime, mt)

	return nil
}

const (
	// xForbidUnknown is a prefix of a vendor extension to indicate forbidden unknown parameters.
	// It should be used together with ParameterIn as a suffix.
	xForbidUnknown = "x-forbid-unknown-"
)

func (r *Reflector) parseParametersIn(
	oc OperationContext, in ParameterIn, propertyMapping map[string]string,
) error {
	o := oc.Operation
	input := oc.Input

	if refl.IsSliceOrMap(input) {
		return nil
	}

	s, err := r.Reflect(input,
		r.withOperation(oc, false, string(in)),
		jsonschema.DefinitionsPrefix("#/components/schemas/"),
		jsonschema.CollectDefinitions(r.collectDefinition),
		jsonschema.PropertyNameMapping(propertyMapping),
		jsonschema.PropertyNameTag(string(in)),
		func(rc *jsonschema.ReflectContext) {
			rc.UnnamedFieldWithTag = true
		},
		jsonschema.SkipEmbeddedMapsSlices,
		jsonschema.InterceptProperty(func(name string, field reflect.StructField, propertySchema *jsonschema.Schema) error {
			s, err := propertySchema.ToSchemaOrBool().ToSimpleMap()
			if err != nil {
				return fmt.Errorf("property %s schema to map: %w", name, err)
			}

			p := Parameter{
				Name:        name,
				In:          in,
				Description: propertySchema.Description,
				Schema:      s,
				Content:     nil,
			}

			swg2CollectionFormat := ""
			refl.ReadStringTag(field.Tag, "collectionFormat", &swg2CollectionFormat)
			switch swg2CollectionFormat {
			case "csv":
				p.WithStyle(ParameterStyleForm).WithExplode(false)
			case "ssv":
				p.WithStyle(ParameterStyleSpaceDelimited).WithExplode(false)
			case "pipes":
				p.WithStyle(ParameterStylePipeDelimited).WithExplode(false)
			case "multi":
				p.WithStyle(ParameterStyleForm).WithExplode(true)
			}

			// Check if parameter is an JSON encoded object.
			property := reflect.New(field.Type).Interface()
			if refl.HasTaggedFields(property, tagJSON) {
				propertySchema, err := r.Reflect(property,
					r.withOperation(oc, false, string(in)),
					jsonschema.DefinitionsPrefix("#/components/schemas/"),
					jsonschema.CollectDefinitions(r.collectDefinition),
					jsonschema.RootRef,
				)
				if err != nil {
					return err
				}

				s, err := propertySchema.ToSchemaOrBool().ToSimpleMap()
				if err != nil {
					return fmt.Errorf("property %s to map: %w", name, err)
				}

				p.Schema = nil
				p.WithContentItem("application/json", MediaType{Schema: s})
			} else {
				ps, err := r.Reflect(reflect.New(field.Type).Interface(),
					r.withOperation(oc, false, string(in)),
					jsonschema.InlineRefs)
				if err != nil {
					return err
				}

				if ps.HasType(jsonschema.Object) {
					p.WithStyle(ParameterStyleDeepObject).WithExplode(true)
				}
			}

			if err := refl.PopulateFieldsFromTags(&p, field.Tag); err != nil {
				return err
			}

			if in == ParameterInPath {
				p.WithRequired(true)
			}

			alreadyExists := false
			for _, ep := range o.Parameters {
				if ep.Parameter != nil && ep.Parameter.In == p.In && ep.Parameter.Name == p.Name {
					alreadyExists = true

					break
				}
			}

			if alreadyExists {
				return fmt.Errorf("parameter %s in %s is already defined", p.Name, p.In)
			}

			o.Parameters = append(o.Parameters, ParameterOrReference{Parameter: &p})

			return nil
		}),
	)
	if err != nil {
		return err
	}

	if s.AdditionalProperties != nil &&
		s.AdditionalProperties.TypeBoolean != nil &&
		!*s.AdditionalProperties.TypeBoolean {
		o.WithMapOfAnythingItem(xForbidUnknown+string(in), true)
	}

	return nil
}

func (r *Reflector) collectDefinition(name string, schema jsonschema.Schema) {
	if _, exists := r.SpecEns().ComponentsEns().Schemas[name]; exists {
		return
	}

	s, err := schema.ToSchemaOrBool().ToSimpleMap()
	if err != nil {
		panic(err)
	}

	r.SpecEns().ComponentsEns().WithSchemasItem(name, s)
}

func (r *Reflector) parseResponseHeader(resp *Response, oc OperationContext) error {
	output := oc.Output
	mapping := oc.RespHeaderMapping

	res := make(map[string]HeaderOrReference)

	schema, err := r.Reflect(output,
		r.withOperation(oc, true, "header"),
		jsonschema.InlineRefs,
		jsonschema.PropertyNameMapping(mapping),
		jsonschema.PropertyNameTag("header"),
		jsonschema.InterceptProperty(func(name string, field reflect.StructField, propertySchema *jsonschema.Schema) error {
			s, err := propertySchema.ToSchemaOrBool().ToSimpleMap()
			if err != nil {
				return fmt.Errorf("property %s schema to map: %w", name, err)
			}

			header := Header{
				Description:   propertySchema.Description,
				Schema:        s,
				Content:       nil,
				Example:       nil,
				Examples:      nil,
				MapOfAnything: nil,
			}

			if deprecated, ok := propertySchema.ExtraProperties["deprecated"].(bool); ok {
				header.Deprecated = &deprecated
			}

			if err := refl.PopulateFieldsFromTags(&header, field.Tag); err != nil {
				return err
			}

			res[name] = HeaderOrReference{
				Header: &header,
			}

			return nil
		}),
	)
	if err != nil {
		return err
	}

	resp.Headers = res

	if schema.Description != nil && resp.Description == "" {
		resp.Description = *schema.Description
	}

	return nil
}

// SetStringResponse sets unstructured response.
func (r *Reflector) SetStringResponse(o *Operation, httpStatus int, contentType string) error {
	return r.SetupResponse(OperationContext{
		Operation:       o,
		HTTPStatus:      httpStatus,
		RespContentType: contentType,
	})
}

// SetJSONResponse sets up operation JSON response.
func (r *Reflector) SetJSONResponse(o *Operation, output interface{}, httpStatus int) error {
	return r.SetupResponse(OperationContext{
		Operation:  o,
		Output:     output,
		HTTPStatus: httpStatus,
	})
}

func (r *Reflector) hasJSONBody(output interface{}) (bool, error) {
	schema, err := r.Reflect(output)
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

// SetupResponse sets up operation response.
func (r *Reflector) SetupResponse(oc OperationContext) error {
	httpStatus := strconv.Itoa(oc.HTTPStatus)
	resp := oc.Operation.ResponsesEns().MapOfResponseOrReferenceValues[httpStatus].Response

	if resp == nil {
		resp = &Response{}
	}

	if oc.Output != nil {
		oc.RespContentType = strings.Split(oc.RespContentType, ";")[0]

		err := r.parseJSONResponse(resp, oc)
		if err != nil {
			return err
		}

		err = r.parseResponseHeader(resp, oc)
		if err != nil {
			return err
		}
	}

	if oc.RespContentType != "" {
		r.ensureResponseContentType(resp, oc.RespContentType)
	}

	if resp.Description == "" {
		resp.Description = http.StatusText(oc.HTTPStatus)
	}

	oc.Operation.Responses.WithMapOfResponseOrReferenceValuesItem(httpStatus, ResponseOrReference{
		Response: resp,
	})

	return nil
}

func (r *Reflector) ensureResponseContentType(resp *Response, contentType string) {
	if _, ok := resp.Content[contentType]; !ok {
		if resp.Content == nil {
			resp.Content = map[string]MediaType{}
		}

		resp.Content[contentType] = MediaType{
			Schema: map[string]interface{}{
				"type": "string",
			}}
	}
}

func (r *Reflector) parseJSONResponse(resp *Response, oc OperationContext) error {
	output := oc.Output
	contentType := oc.RespContentType

	// Check if output structure exposes meaningful schema.
	if hasJSONBody, err := r.hasJSONBody(output); err == nil && !hasJSONBody {
		return nil
	}

	schema, err := r.Reflect(output,
		r.withOperation(oc, true, "body"),
		jsonschema.RootRef,
		jsonschema.DefinitionsPrefix("#/components/schemas/"),
		jsonschema.CollectDefinitions(r.collectDefinition),
	)
	if err != nil {
		return err
	}

	s, err := schema.ToSchemaOrBool().ToSimpleMap()
	if err != nil {
		return fmt.Errorf("response schema to map: %w", err)
	}

	if resp.Content == nil {
		resp.Content = map[string]MediaType{}
	}

	if contentType == "" {
		contentType = mimeJSON
	}

	resp.Content[contentType] = MediaType{
		Schema:   s,
		Example:  nil,
		Examples: nil,
		Encoding: nil,
	}

	if schema.Description != nil && resp.Description == "" {
		resp.Description = *schema.Description
	}

	return nil
}

type ocCtxKey struct{}

func (r *Reflector) withOperation(oc OperationContext, processingResponse bool, in string) func(rc *jsonschema.ReflectContext) {
	return func(rc *jsonschema.ReflectContext) {
		oc.ProcessingResponse = processingResponse
		oc.ProcessingIn = in

		rc.Context = context.WithValue(rc.Context, ocCtxKey{}, oc)
	}
}

// OperationCtx retrieves operation context from reflect context.
func OperationCtx(rc *jsonschema.ReflectContext) (OperationContext, bool) {
	if oc, ok := rc.Value(ocCtxKey{}).(OperationContext); ok {
		return oc, true
	}

	return OperationContext{}, false
}
