package openapi3

import (
	"errors"
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
		r.Spec = &Spec{Openapi: "3.0.2"}
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
	RespHeaderMapping map[string]string
}

// SetupRequest sets up operation parameters.
func (r *Reflector) SetupRequest(oc OperationContext) error {
	return joinErrors(
		r.parseParametersIn(oc.Operation, oc.Input, ParameterInQuery, oc.ReqQueryMapping),
		r.parseParametersIn(oc.Operation, oc.Input, ParameterInPath, oc.ReqPathMapping),
		r.parseParametersIn(oc.Operation, oc.Input, ParameterInCookie, oc.ReqCookieMapping),
		r.parseParametersIn(oc.Operation, oc.Input, ParameterInHeader, oc.ReqHeaderMapping),
		r.parseRequestBody(oc.Operation, oc.Input, tagJSON, mimeJSON, oc.HTTPMethod, nil),
		r.parseRequestBody(oc.Operation, oc.Input, tagFormData, mimeFormUrlencoded, oc.HTTPMethod, oc.ReqFormDataMapping),
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

var (
	typeOfMultipartFile       = reflect.TypeOf((*multipart.File)(nil)).Elem()
	typeOfMultipartFileHeader = reflect.TypeOf((*multipart.FileHeader)(nil)).Elem()
)

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

func (r *Reflector) parseRequestBody(
	o *Operation, input interface{}, tag, mime string, httpMethod string, mapping map[string]string,
) error {
	httpMethod = strings.ToUpper(httpMethod)
	_, forceRequestBody := input.(RequestBodyEnforcer)

	// GET and HEAD requests should not have body.
	if (httpMethod == http.MethodGet || httpMethod == http.MethodHead) && !forceRequestBody {
		return nil
	}

	hasTaggedFields := refl.HasTaggedFields(input, tag)

	// Form data can not have map or array as body.
	if !hasTaggedFields && len(mapping) == 0 && tag != tagJSON {
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

			if v.Type().Implements(typeOfMultipartFile) || v.Type() == typeOfMultipartFileHeader {
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

	schemaOrRef := SchemaOrRef{}

	schemaOrRef.FromJSONSchema(schema.ToSchemaOrBool())

	mt := MediaType{
		Schema: &schemaOrRef,
	}

	for name, def := range schema.Definitions {
		s := SchemaOrRef{}

		s.FromJSONSchema(def)

		r.SpecEns().ComponentsEns().SchemasEns().WithMapOfSchemaOrRefValuesItem(definitionPrefix+name, s)
	}

	if mime == mimeFormUrlencoded && hasFileUpload {
		mime = mimeMultipart
	}

	o.RequestBodyEns().RequestBodyEns().WithContentItem(mime, mt)

	return nil
}

func (r *Reflector) parseParametersIn(
	o *Operation, input interface{}, in ParameterIn, propertyMapping map[string]string,
) error {
	if refl.IsSliceOrMap(input) {
		return nil
	}

	_, err := r.Reflect(input,
		jsonschema.DefinitionsPrefix("#/components/schemas/"),
		jsonschema.CollectDefinitions(r.collectDefinition),
		jsonschema.PropertyNameMapping(propertyMapping),
		jsonschema.PropertyNameTag(string(in)),
		jsonschema.SkipEmbeddedMapsSlices,
		jsonschema.InterceptProperty(func(name string, field reflect.StructField, propertySchema *jsonschema.Schema) error {
			s := SchemaOrRef{}
			s.FromJSONSchema(propertySchema.ToSchemaOrBool())

			if s.Schema != nil && s.Schema.Nullable != nil {
				s.Schema.Nullable = nil
			}

			p := Parameter{
				Name:        name,
				In:          in,
				Description: propertySchema.Description,
				Schema:      &s,
				Content:     nil,
				Example:     nil,
				Examples:    nil,
				Location:    nil,
			}

			swg2CollectionFormat := ""
			refl.ReadStringTag(field.Tag, "collectionFormat", &swg2CollectionFormat)
			switch swg2CollectionFormat {
			case "csv":
				p.WithStyle("form").WithExplode(false)
			case "ssv":
				p.WithStyle("spaceDelimited").WithExplode(false)
			case "pipes":
				p.WithStyle("pipeDelimited").WithExplode(false)
			case "multi":
				p.WithStyle("form").WithExplode(true)
			}

			err := refl.PopulateFieldsFromTags(&p, field.Tag)
			if err != nil {
				return err
			}

			if in == ParameterInPath {
				p.WithRequired(true)
			}

			o.Parameters = append(o.Parameters, ParameterOrRef{Parameter: &p})

			return nil
		}),
	)
	if err != nil {
		return err
	}

	return nil
}

func (r *Reflector) collectDefinition(name string, schema jsonschema.Schema) {
	s := SchemaOrRef{}
	s.FromJSONSchema(schema.ToSchemaOrBool())

	r.SpecEns().ComponentsEns().SchemasEns().WithMapOfSchemaOrRefValuesItem(name, s)
}

func (r *Reflector) parseResponseHeader(resp *Response, output interface{}, mapping map[string]string) error {
	res := make(map[string]HeaderOrRef)

	schema, err := r.Reflect(output,
		jsonschema.InlineRefs,
		jsonschema.PropertyNameMapping(mapping),
		jsonschema.PropertyNameTag("header"),
		jsonschema.InterceptProperty(func(name string, field reflect.StructField, propertySchema *jsonschema.Schema) error {
			s := SchemaOrRef{}
			s.FromJSONSchema(propertySchema.ToSchemaOrBool())

			header := Header{
				Description:   propertySchema.Description,
				Deprecated:    s.Schema.Deprecated,
				Schema:        &s,
				Content:       nil,
				Example:       nil,
				Examples:      nil,
				MapOfAnything: nil,
			}

			err := refl.PopulateFieldsFromTags(&header, field.Tag)
			if err != nil {
				return err
			}

			res[name] = HeaderOrRef{
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

	if schema.Type == nil {
		return false, nil
	}

	if schema.HasType(jsonschema.Object) && schema.AdditionalProperties == nil && len(schema.Properties) == 0 {
		return false, nil
	}

	return true, nil
}

// SetupResponse sets up operation response.
func (r *Reflector) SetupResponse(oc OperationContext) error {
	resp := Response{}

	if oc.Output != nil {
		err := r.parseJSONResponse(&resp, oc.Output)
		if err != nil {
			return err
		}

		err = r.parseResponseHeader(&resp, oc.Output, oc.RespHeaderMapping)
		if err != nil {
			return err
		}
	}

	if resp.Description == "" {
		resp.Description = http.StatusText(oc.HTTPStatus)
	}

	oc.Operation.Responses.WithMapOfResponseOrRefValuesItem(strconv.Itoa(oc.HTTPStatus), ResponseOrRef{
		Response: &resp,
	})

	return nil
}

func (r *Reflector) parseJSONResponse(resp *Response, output interface{}) error {
	// Check if output structure exposes meaningful schema.
	if hasJSONBody, err := r.hasJSONBody(output); err == nil && !hasJSONBody {
		return nil
	}

	schema, err := r.Reflect(output,
		jsonschema.RootRef,
		jsonschema.DefinitionsPrefix("#/components/schemas/"),
		jsonschema.CollectDefinitions(r.collectDefinition),
	)
	if err != nil {
		return err
	}

	oaiSchema := SchemaOrRef{}
	oaiSchema.FromJSONSchema(schema.ToSchemaOrBool())

	if oaiSchema.Schema != nil {
		oaiSchema.Schema.Nullable = nil
	}

	if resp.Content == nil {
		resp.Content = map[string]MediaType{}
	}

	resp.Content[mimeJSON] = MediaType{
		Schema:        &oaiSchema,
		Example:       nil,
		Examples:      nil,
		Encoding:      nil,
		MapOfAnything: nil,
	}

	if schema.Description != nil && resp.Description == "" {
		resp.Description = *schema.Description
	}

	return nil
}
