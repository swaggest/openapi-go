package openapi3

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/swaggest/jsonschema-go"
	"github.com/swaggest/openapi-go"
	"github.com/swaggest/openapi-go/internal"
	"github.com/swaggest/refl"
)

// Reflector builds OpenAPI Schema with reflected structures.
type Reflector struct {
	jsonschema.Reflector
	Spec *Spec
}

// OperationAccessor is an extra behavior present in openapi.OperationContext.
type OperationAccessor interface {
	Operation() *Operation
}

// NewOperationContext initializes openapi.OperationContext to be prepared
// and added later with Reflector.SetOperation.
func (r *Reflector) NewOperationContext(method, pathPattern string) (openapi.OperationContext, error) {
	method, pathPattern, pathParams, err := internal.SanitizeMethodPath(method, pathPattern)
	if err != nil {
		return nil, err
	}

	pathItem := r.SpecEns().Paths.MapOfPathItemValues[pathPattern]
	operation, found := pathItem.MapOfOperationValues[method]

	if found {
		return nil, fmt.Errorf("operation already exists: %s %s", method, pathPattern)
	}

	oc := operationContext{
		OperationContext: internal.NewOperationContext(method, pathPattern),
		op:               &operation,
		pathParams:       pathParams,
	}

	return oc, nil
}

// ResolveJSONSchemaRef builds JSON Schema from OpenAPI Component Schema reference.
//
// Can be used in jsonschema.Schema IsTrivial().
func (r *Reflector) ResolveJSONSchemaRef(ref string) (s jsonschema.SchemaOrBool, found bool) {
	if r.Spec == nil || r.Spec.Components == nil || r.Spec.Components.Schemas == nil ||
		!strings.HasPrefix(ref, "#/components/schemas/") {
		return s, false
	}

	ref = strings.TrimPrefix(ref, "#/components/schemas/")
	os, found := r.Spec.Components.Schemas.MapOfSchemaOrRefValues[ref]

	if found {
		s = os.ToJSONSchema(r.Spec)
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
		r.Spec = &Spec{Openapi: "3.0.3"}
	}

	return r.Spec
}

// OperationContext describes operation.
//
// Deprecated: use Reflector.NewOperationContext.
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

type operationContext struct {
	*internal.OperationContext

	op *Operation

	pathParams map[string]bool
}

// Operation returns OpenAPI 3 operation for customization.
func (o operationContext) Operation() *Operation {
	return o.op
}

func toOpCtx(c OperationContext) operationContext {
	oc := internal.NewOperationContext(c.HTTPMethod, "")

	oc.AddReqStructure(c.Input,
		openapi.FieldMapping(openapi.InHeader, c.ReqHeaderMapping),
		openapi.FieldMapping(openapi.InQuery, c.ReqQueryMapping),
		openapi.FieldMapping(openapi.InCookie, c.ReqCookieMapping),
		openapi.FieldMapping(openapi.InPath, c.ReqPathMapping),
		openapi.FieldMapping(openapi.InFormData, c.ReqFormDataMapping),
	)

	oc.AddRespStructure(c.Output,
		openapi.FieldMapping(openapi.InHeader, c.RespHeaderMapping),
		func(cu *openapi.ContentUnit) {
			cu.ContentType = c.RespContentType
			cu.HTTPStatus = c.HTTPStatus
		},
	)

	oc.SetProcessingIn(openapi.In(c.ProcessingIn))
	oc.SetIsProcessingResponse(c.ProcessingResponse)

	return operationContext{
		OperationContext: oc,
		op:               c.Operation,
	}
}

func fromOpCtx(oc openapi.OperationContext) OperationContext {
	c := OperationContext{}

	c.HTTPMethod = oc.Method()

	if req := oc.Request(); len(req) > 0 {
		cu := req[0]

		c.Input = cu.Structure
		c.ReqQueryMapping = cu.FieldMapping(openapi.InQuery)
		c.ReqPathMapping = cu.FieldMapping(openapi.InPath)
		c.ReqHeaderMapping = cu.FieldMapping(openapi.InHeader)
		c.ReqFormDataMapping = cu.FieldMapping(openapi.InFormData)
		c.ReqCookieMapping = cu.FieldMapping(openapi.InCookie)
	}

	if resp := oc.Response(); len(resp) > 0 {
		cu := resp[0]

		c.Output = cu.Structure
		c.RespHeaderMapping = cu.FieldMapping(openapi.InHeader)
		c.RespContentType = cu.ContentType
		c.HTTPStatus = cu.HTTPStatus
	}

	c.ProcessingResponse = oc.IsProcessingResponse()
	c.ProcessingIn = string(oc.ProcessingIn())

	return c
}

// AddOperation configures operation request and response schema.
func (r *Reflector) AddOperation(oc openapi.OperationContext) error {
	c, ok := oc.(operationContext)
	if !ok {
		return fmt.Errorf("wrong operation context %T received, %T expected", oc, operationContext{})
	}

	if err := c.op.validatePathParams(c.pathParams); err != nil {
		return err
	}

	return r.SpecEns().AddOperation(oc.Method(), oc.PathPattern(), *c.op)
}

func (r *Reflector) setupRequest(o *Operation, oc openapi.OperationContext) error {
	for _, cu := range oc.Request() {
		switch cu.ContentType {
		case "":
			return joinErrors(
				r.parseParameters(o, oc, cu),
				r.parseRequestBody(o, oc, cu, mimeJSON, oc.Method(), nil, tagJSON),
				r.parseRequestBody(o, oc, cu, mimeFormUrlencoded, oc.Method(), cu.FieldMapping(openapi.InFormData), tagFormData, tagForm),
			)
		case mimeJSON:
			return joinErrors(
				r.parseParameters(o, oc, cu),
				r.parseRequestBody(o, oc, cu, mimeJSON, oc.Method(), nil, tagJSON),
			)
		case mimeFormUrlencoded, mimeMultipart:
			return joinErrors(
				r.parseParameters(o, oc, cu),
				r.parseRequestBody(o, oc, cu, mimeFormUrlencoded, oc.Method(), cu.FieldMapping(openapi.InFormData), tagFormData, tagForm),
			)
		default:
			r.stringRequestBody(o, cu.ContentType, cu.Format)
		}
	}

	return nil
}

// SetupRequest sets up operation parameters.
func (r *Reflector) SetupRequest(c OperationContext) error {
	return r.setupRequest(c.Operation, toOpCtx(c))
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
	tagForm            = "form"
	tagHeader          = "header"
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

func mediaType(format string) MediaType {
	schema := jsonschema.String.ToSchemaOrBool()
	if format != "" {
		schema.TypeObject.WithFormat(format)
	}

	schemaOrRef := SchemaOrRef{}

	schemaOrRef.FromJSONSchema(schema)

	mt := MediaType{
		Schema: &schemaOrRef,
	}

	return mt
}

func (r *Reflector) stringRequestBody(
	o *Operation,
	mime string,
	format string,
) {
	o.RequestBodyEns().RequestBodyEns().WithContentItem(mime, mediaType(format))
}

func (r *Reflector) parseRequestBody(
	o *Operation,
	oc openapi.OperationContext,
	cu openapi.ContentUnit,
	mime string,
	httpMethod string,
	mapping map[string]string,
	tag string,
	additionalTags ...string,
) error {
	input := cu.Structure

	httpMethod = strings.ToUpper(httpMethod)
	_, forceRequestBody := input.(openapi.RequestBodyEnforcer)
	_, forceJSONRequestBody := input.(openapi.RequestJSONBodyEnforcer)

	// GET, HEAD, DELETE and TRACE requests should not have body.
	switch httpMethod {
	case http.MethodGet, http.MethodHead, http.MethodDelete, http.MethodTrace:
		if !forceRequestBody {
			return nil
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
		openapi.WithOperationCtx(oc, false, "body"),
		jsonschema.DefinitionsPrefix("#/components/schemas/"+definitionPrefix),
		jsonschema.RootRef,
		jsonschema.PropertyNameMapping(mapping),
		jsonschema.PropertyNameTag(tag, additionalTags...),
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

const (
	// xForbidUnknown is a prefix of a vendor extension to indicate forbidden unknown parameters.
	// It should be used together with ParameterIn as a suffix.
	xForbidUnknown = "x-forbid-unknown-"
)

func (r *Reflector) parseParameters(o *Operation, oc openapi.OperationContext, cu openapi.ContentUnit) error {
	return joinErrors(r.parseParametersIn(o, oc, cu, openapi.InQuery, tagForm),
		r.parseParametersIn(o, oc, cu, openapi.InPath),
		r.parseParametersIn(o, oc, cu, openapi.InCookie),
		r.parseParametersIn(o, oc, cu, openapi.InHeader),
	)
}

func (r *Reflector) parseParametersIn(
	o *Operation,
	oc openapi.OperationContext,
	c openapi.ContentUnit,
	in openapi.In,
	additionalTags ...string,
) error {
	input := c.Structure
	propertyMapping := c.FieldMapping(in)

	if refl.IsSliceOrMap(input) {
		return nil
	}

	defNamePrefix := strings.Title(string(in))
	definitionsPrefix := "#/components/schemas/" + defNamePrefix

	s, err := r.Reflect(input,
		openapi.WithOperationCtx(oc, false, in),
		jsonschema.DefinitionsPrefix(definitionsPrefix),
		jsonschema.CollectDefinitions(r.collectDefinition(defNamePrefix)),
		jsonschema.PropertyNameMapping(propertyMapping),
		jsonschema.PropertyNameTag(string(in), additionalTags...),
		func(rc *jsonschema.ReflectContext) {
			rc.UnnamedFieldWithTag = true
		},
		jsonschema.SkipEmbeddedMapsSlices,
		jsonschema.InterceptProp(func(params jsonschema.InterceptPropParams) error {
			if !params.Processed || len(params.Path) > 1 {
				return nil
			}

			name := params.Name
			propertySchema := params.PropertySchema
			field := params.Field

			s := SchemaOrRef{}
			s.FromJSONSchema(propertySchema.ToSchemaOrBool())

			if s.Schema != nil && s.Schema.Nullable != nil && field.Type.Kind() != reflect.Ptr {
				s.Schema.Nullable = nil
			}

			p := Parameter{
				Name:        name,
				In:          ParameterIn(in),
				Description: propertySchema.Description,
				Schema:      &s,
				Content:     nil,
			}

			swg2CollectionFormat := ""
			refl.ReadStringTag(field.Tag, "collectionFormat", &swg2CollectionFormat)
			switch swg2CollectionFormat {
			case "csv":
				p.WithStyle(string(QueryParameterStyleForm)).WithExplode(false)
			case "ssv":
				p.WithStyle(string(QueryParameterStyleSpaceDelimited)).WithExplode(false)
			case "pipes":
				p.WithStyle(string(QueryParameterStylePipeDelimited)).WithExplode(false)
			case "multi":
				p.WithStyle(string(QueryParameterStyleForm)).WithExplode(true)
			}

			// Check if parameter is an JSON encoded object.
			property := reflect.New(field.Type).Interface()
			if refl.HasTaggedFields(property, tagJSON) && !refl.HasTaggedFields(property, string(in)) {
				propertySchema, err := r.Reflect(property,
					openapi.WithOperationCtx(oc, false, in),
					jsonschema.DefinitionsPrefix(definitionsPrefix),
					jsonschema.CollectDefinitions(r.collectDefinition(defNamePrefix)),
					jsonschema.RootRef,
				)
				if err != nil {
					return err
				}

				openapiSchema := SchemaOrRef{}
				openapiSchema.FromJSONSchema(propertySchema.ToSchemaOrBool())
				p.Schema = nil
				p.WithContentItem("application/json", MediaType{Schema: &openapiSchema})
			} else {
				ps, err := r.Reflect(reflect.New(field.Type).Interface(),
					openapi.WithOperationCtx(oc, false, in),
					jsonschema.InlineRefs)
				if err != nil {
					return err
				}

				if ps.HasType(jsonschema.Object) {
					p.WithStyle(string(QueryParameterStyleDeepObject)).WithExplode(true)
				}
			}

			err := refl.PopulateFieldsFromTags(&p, field.Tag)
			if err != nil {
				return err
			}

			if in == openapi.InPath {
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

			o.Parameters = append(o.Parameters, ParameterOrRef{Parameter: &p})

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

func (r *Reflector) collectDefinition(namePrefix string) func(name string, schema jsonschema.Schema) {
	return func(name string, schema jsonschema.Schema) {
		name = namePrefix + name

		if _, exists := r.SpecEns().ComponentsEns().SchemasEns().MapOfSchemaOrRefValues[name]; exists {
			return
		}

		s := SchemaOrRef{}
		s.FromJSONSchema(schema.ToSchemaOrBool())

		r.SpecEns().ComponentsEns().SchemasEns().WithMapOfSchemaOrRefValuesItem(name, s)
	}
}

func (r *Reflector) parseResponseHeader(resp *Response, oc openapi.OperationContext, cu openapi.ContentUnit) error {
	output := cu.Structure
	mapping := cu.FieldMapping(openapi.InHeader)

	if output == nil {
		return nil
	}

	res := make(map[string]HeaderOrRef)

	schema, err := r.Reflect(output,
		openapi.WithOperationCtx(oc, true, openapi.InHeader),
		jsonschema.InlineRefs,
		jsonschema.PropertyNameMapping(mapping),
		jsonschema.PropertyNameTag(tagHeader),
		jsonschema.InterceptProp(func(params jsonschema.InterceptPropParams) error {
			if !params.Processed {
				return nil
			}

			propertySchema := params.PropertySchema
			field := params.Field
			name := params.Name

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

func (r *Reflector) setupResponse(o *Operation, oc openapi.OperationContext) error {
	for _, cu := range oc.Response() {
		if cu.HTTPStatus == 0 {
			cu.HTTPStatus = 200
		}

		cu.ContentType = strings.Split(cu.ContentType, ";")[0]

		httpStatus := strconv.Itoa(cu.HTTPStatus)
		resp := o.Responses.MapOfResponseOrRefValues[httpStatus].Response

		if resp == nil {
			resp = &Response{}
		}

		if err := joinErrors(
			r.parseJSONResponse(resp, oc, cu),
			r.parseResponseHeader(resp, oc, cu),
		); err != nil {
			return err
		}

		if cu.ContentType != "" {
			r.ensureResponseContentType(resp, cu.ContentType, cu.Format)
		}

		if resp.Description == "" {
			resp.Description = http.StatusText(cu.HTTPStatus)
		}

		o.Responses.WithMapOfResponseOrRefValuesItem(httpStatus, ResponseOrRef{
			Response: resp,
		})
	}

	return nil
}

// SetupResponse sets up operation response.
func (r *Reflector) SetupResponse(oc OperationContext) error {
	return r.setupResponse(oc.Operation, toOpCtx(oc))
}

func (r *Reflector) ensureResponseContentType(resp *Response, contentType string, format string) {
	if _, ok := resp.Content[contentType]; !ok {
		if resp.Content == nil {
			resp.Content = map[string]MediaType{}
		}

		resp.Content[contentType] = mediaType(format)
	}
}

func (r *Reflector) parseJSONResponse(resp *Response, oc openapi.OperationContext, cu openapi.ContentUnit) error {
	output := cu.Structure
	contentType := cu.ContentType

	if output == nil {
		return nil
	}

	// Check if output structure exposes meaningful schema.
	if hasJSONBody, err := r.hasJSONBody(output); err == nil && !hasJSONBody {
		return nil
	}

	schema, err := r.Reflect(output,
		openapi.WithOperationCtx(oc, true, openapi.InBody),
		jsonschema.RootRef,
		jsonschema.DefinitionsPrefix("#/components/schemas/"),
		jsonschema.CollectDefinitions(r.collectDefinition("")),
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

	if contentType == "" {
		contentType = mimeJSON
	}

	resp.Content[contentType] = MediaType{
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

// OperationCtx retrieves operation context from reflect context.
func OperationCtx(rc *jsonschema.ReflectContext) (OperationContext, bool) {
	if oc, ok := openapi.OperationCtx(rc); ok {
		return fromOpCtx(oc), true
	}

	return OperationContext{}, false
}
