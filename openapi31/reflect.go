package openapi31

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"regexp"
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

// NewReflector creates an instance of OpenAPI 3.1 reflector.
func NewReflector() *Reflector {
	r := &Reflector{}
	r.SpecEns()

	return r
}

// NewOperationContext initializes openapi.OperationContext to be prepared
// and added later with Reflector.AddOperation.
func (r *Reflector) NewOperationContext(method, pathPattern string) (openapi.OperationContext, error) {
	method, pathPattern, pathParams, err := openapi.SanitizeMethodPath(method, pathPattern)
	if err != nil {
		return nil, err
	}

	pathItem := r.SpecEns().PathsEns().MapOfPathItemValues[pathPattern]

	operation, err := pathItem.Operation(method)
	if err != nil {
		return nil, err
	}

	if operation != nil {
		return nil, fmt.Errorf("operation already exists: %s %s", method, pathPattern)
	}

	operation = &Operation{}

	pathParamsMap := make(map[string]bool, len(pathParams))
	for _, p := range pathParams {
		pathParamsMap[p] = true
	}

	oc := operationContext{
		OperationContext: internal.NewOperationContext(method, pathPattern),
		op:               operation,
		pathParams:       pathParamsMap,
	}

	return oc, nil
}

// ResolveJSONSchemaRef builds JSON Schema from OpenAPI Component Schema reference.
//
// Can be used in jsonschema.Schema IsTrivial().
func (r *Reflector) ResolveJSONSchemaRef(ref string) (s jsonschema.SchemaOrBool, found bool) {
	if r.Spec == nil || r.Spec.Components == nil || r.Spec.Components.Schemas == nil ||
		!strings.HasPrefix(ref, componentsSchemas) {
		return s, false
	}

	ref = strings.TrimPrefix(ref, componentsSchemas)
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

type operationContext struct {
	*internal.OperationContext

	op *Operation

	pathParams map[string]bool
}

// OperationExposer grants access to underlying *Operation.
type OperationExposer interface {
	Operation() *Operation
}

func (o operationContext) AddSecurity(securityName string, scopes ...string) {
	if scopes == nil {
		scopes = []string{}
	}

	o.op.Security = append(o.op.Security, map[string][]string{securityName: scopes})
}

func (o operationContext) SetTags(tags ...string) {
	o.op.WithTags(tags...)
}

func (o operationContext) Tags() []string {
	return o.op.Tags
}

func (o operationContext) SetIsDeprecated(isDeprecated bool) {
	o.op.WithDeprecated(isDeprecated)
}

func (o operationContext) IsDeprecated() bool {
	return o.op.Deprecated != nil && *o.op.Deprecated
}

func (o operationContext) SetSummary(summary string) {
	o.op.WithSummary(summary)
}

func (o operationContext) Summary() string {
	if o.op.Summary == nil {
		return ""
	}

	return *o.op.Summary
}

func (o operationContext) SetDescription(description string) {
	o.op.WithDescription(description)
}

func (o operationContext) Description() string {
	if o.op.Description == nil {
		return ""
	}

	return *o.op.Description
}

func (o operationContext) SetID(operationID string) {
	o.op.WithID(operationID)
}

func (o operationContext) ID() string {
	if o.op.ID == nil {
		return ""
	}

	return *o.op.ID
}

func (o operationContext) UnknownParamsAreForbidden(in openapi.In) bool {
	return o.op.UnknownParamIsForbidden(ParameterIn(in))
}

// Operation returns OpenAPI 3 operation for customization.
func (o operationContext) Operation() *Operation {
	return o.op
}

// AddOperation configures operation request and response schema.
func (r *Reflector) AddOperation(oc openapi.OperationContext) error {
	c, ok := oc.(operationContext)
	if !ok {
		return fmt.Errorf("wrong operation context %T received, %T expected", oc, operationContext{})
	}

	if err := r.setupRequest(c.op, oc); err != nil {
		return fmt.Errorf("setup request %s %s: %w", oc.Method(), oc.PathPattern(), err)
	}

	if err := c.op.validatePathParams(c.pathParams); err != nil {
		return fmt.Errorf("validate path params %s %s: %w", oc.Method(), oc.PathPattern(), err)
	}

	if err := r.setupResponse(c.op, oc); err != nil {
		return fmt.Errorf("setup response %s %s: %w", oc.Method(), oc.PathPattern(), err)
	}

	return r.SpecEns().AddOperation(oc.Method(), oc.PathPattern(), *c.op)
}

func (r *Reflector) setupRequest(o *Operation, oc openapi.OperationContext) error {
	for _, cu := range oc.Request() {
		switch cu.ContentType {
		case "":
			if err := joinErrors(
				r.parseRequestBody(o, oc, cu, mimeFormUrlencoded, oc.Method(), cu.FieldMapping(openapi.InFormData), tagFormData, tagForm),
				r.parseParameters(o, oc, cu),
				r.parseRequestBody(o, oc, cu, mimeJSON, oc.Method(), nil, tagJSON),
			); err != nil {
				return err
			}
		case mimeJSON:
			if err := joinErrors(
				r.parseParameters(o, oc, cu),
				r.parseRequestBody(o, oc, cu, mimeJSON, oc.Method(), nil, tagJSON),
			); err != nil {
				return err
			}
		case mimeFormUrlencoded, mimeMultipart:
			if err := joinErrors(
				r.parseRequestBody(o, oc, cu, mimeFormUrlencoded, oc.Method(), cu.FieldMapping(openapi.InFormData), tagFormData, tagForm),
				r.parseParameters(o, oc, cu),
			); err != nil {
				return err
			}
		default:
			r.stringRequestBody(o, cu.ContentType, cu.Format)
		}

		if cu.Description != "" && o.RequestBody != nil && o.RequestBody.RequestBody != nil {
			o.RequestBody.RequestBody.WithDescription(cu.Description)
		}

		if cu.Customize != nil && o.RequestBody != nil {
			cu.Customize(o.RequestBody)
		}
	}

	return nil
}

const (
	tagJSON            = "json"
	tagFormData        = "formData"
	tagForm            = "form"
	mimeJSON           = "application/json"
	mimeFormUrlencoded = "application/x-www-form-urlencoded"
	mimeMultipart      = "multipart/form-data"

	componentsSchemas = "#/components/schemas/"
)

func mediaType(format string) MediaType {
	schema := jsonschema.String.ToSchemaOrBool()
	if format != "" {
		schema.TypeObject.WithFormat(format)
	}

	sm, err := schema.ToSimpleMap()
	if err != nil {
		panic("BUG: " + err.Error())
	}

	mt := MediaType{
		Schema: sm,
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
	schema, hasFileUpload, err := internal.ReflectRequestBody(
		true,
		r.JSONSchemaReflector(),
		cu,
		httpMethod,
		mapping,
		tag,
		additionalTags,
		openapi.WithOperationCtx(oc, false, "body"),
		jsonschema.DefinitionsPrefix(componentsSchemas),
	)
	if err != nil || schema == nil {
		return err
	}

	definitions := schema.Definitions
	schema.Definitions = nil

	sm, err := schema.ToSchemaOrBool().ToSimpleMap()
	if err != nil {
		return err
	}

	mt := MediaType{
		Schema: sm,
	}

	for name, def := range definitions {
		sm, err := def.ToSimpleMap()
		if err != nil {
			return err
		}

		r.SpecEns().ComponentsEns().WithSchemasItem(name, sm)
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
	if refl.IsSliceOrMap(c.Structure) {
		return nil
	}

	s, err := internal.ReflectParametersIn(
		r.JSONSchemaReflector(), oc, c, in, r.collectDefinition(), func(params jsonschema.InterceptPropParams) error {
			if !params.Processed || len(params.Path) > 1 {
				return nil
			}

			name := params.Name
			propertySchema := params.PropertySchema
			field := params.Field

			sm, err := propertySchema.ToSchemaOrBool().ToSimpleMap()
			if err != nil {
				return err
			}

			p := Parameter{
				Name:        name,
				In:          ParameterIn(in),
				Description: propertySchema.Description,
				Schema:      sm,
				Content:     nil,
			}

			collectionFormat := ""
			refl.ReadStringTag(field.Tag, "collectionFormat", &collectionFormat)

			switch collectionFormat {
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
			if collectionFormat == "json" || //nolint:nestif
				(refl.HasTaggedFields(property, tagJSON) && !refl.HasTaggedFields(property, string(in))) {
				propertySchema, err := r.Reflect(property,
					openapi.WithOperationCtx(oc, false, in),
					jsonschema.DefinitionsPrefix(componentsSchemas),
					jsonschema.CollectDefinitions(r.collectDefinition()),
					jsonschema.RootRef,
					sanitizeDefName,
				)
				if err != nil {
					return err
				}

				sm, err := propertySchema.ToSchemaOrBool().ToSimpleMap()
				if err != nil {
					return err
				}

				p.Schema = nil
				p.WithContentItem("application/json", MediaType{Schema: sm})
			} else {
				ps, err := r.Reflect(reflect.New(field.Type).Interface(),
					openapi.WithOperationCtx(oc, false, in),
					jsonschema.InlineRefs,
					sanitizeDefName,
				)
				if err != nil {
					return err
				}

				if ps.HasType(jsonschema.Object) {
					p.WithStyle(ParameterStyleDeepObject).WithExplode(true)
				}
			}

			err = refl.PopulateFieldsFromTags(&p, field.Tag)
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

			o.Parameters = append(o.Parameters, ParameterOrReference{Parameter: &p})

			return nil
		}, additionalTags...,
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

var defNameSanitizer = regexp.MustCompile(`[^a-zA-Z0-9.\-_]+`)

func sanitizeDefName(rc *jsonschema.ReflectContext) {
	jsonschema.InterceptDefName(func(_ reflect.Type, defaultDefName string) string {
		return defNameSanitizer.ReplaceAllString(defaultDefName, "")
	})(rc)
}

func (r *Reflector) collectDefinition() func(name string, schema jsonschema.Schema) {
	return func(name string, schema jsonschema.Schema) {
		if _, exists := r.SpecEns().ComponentsEns().Schemas[name]; exists {
			return
		}

		sm, err := schema.ToSchemaOrBool().ToSimpleMap()
		if err != nil {
			panic("BUG:" + err.Error())
		}

		r.SpecEns().ComponentsEns().WithSchemasItem(name, sm)
	}
}

func (r *Reflector) parseResponseHeader(resp *Response, oc openapi.OperationContext, cu openapi.ContentUnit) error {
	if cu.Structure == nil {
		return nil
	}

	res := make(map[string]HeaderOrReference)

	schema, err := internal.ReflectResponseHeader(r.JSONSchemaReflector(), oc, cu,
		func(params jsonschema.InterceptPropParams) error {
			if !params.Processed || len(params.Path) > 1 { // only top-level fields (including embedded).
				return nil
			}

			propertySchema := params.PropertySchema
			field := params.Field
			name := params.Name

			sm, err := propertySchema.ToSchemaOrBool().ToSimpleMap()
			if err != nil {
				return err
			}

			header := Header{
				Description:   propertySchema.Description,
				Deprecated:    isDeprecated(propertySchema.ToSchemaOrBool()),
				Schema:        sm,
				Content:       nil,
				Example:       nil,
				Examples:      nil,
				MapOfAnything: nil,
			}

			err = refl.PopulateFieldsFromTags(&header, field.Tag)
			if err != nil {
				return err
			}

			res[name] = HeaderOrReference{
				Header: &header,
			}

			return nil
		},
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

func (r *Reflector) setupResponse(o *Operation, oc openapi.OperationContext) error {
	for _, cu := range oc.Response() {
		if cu.HTTPStatus == 0 && !cu.IsDefault {
			cu.HTTPStatus = http.StatusOK
		}

		cu.ContentType = strings.Split(cu.ContentType, ";")[0]

		httpStatus := strconv.Itoa(cu.HTTPStatus)
		resp := o.ResponsesEns().MapOfResponseOrReferenceValues[httpStatus].Response

		switch {
		case cu.IsDefault:
			httpStatus = "default"

			if o.Responses.Default == nil {
				o.Responses.Default = &ResponseOrReference{}
			}

			resp = o.Responses.Default.Response
		case cu.HTTPStatus > 0 && cu.HTTPStatus < 6:
			httpStatus = strconv.Itoa(cu.HTTPStatus) + "XX"
			resp = o.Responses.MapOfResponseOrReferenceValues[httpStatus].Response
		}

		if resp == nil {
			resp = &Response{}
		}

		if strings.ToUpper(oc.Method()) != http.MethodHead {
			if err := joinErrors(
				r.parseJSONResponse(resp, oc, cu),
				r.parseResponseHeader(resp, oc, cu),
			); err != nil {
				return err
			}

			if cu.ContentType != "" {
				r.ensureResponseContentType(resp, cu.ContentType, cu.Format)
			}
		} else {
			// Only headers with HEAD method.
			if err := r.parseResponseHeader(resp, oc, cu); err != nil {
				return err
			}
		}

		if cu.Description != "" {
			resp.Description = cu.Description
		}

		if resp.Description == "" {
			resp.Description = http.StatusText(cu.HTTPStatus)
		}

		ror := ResponseOrReference{Response: resp}

		if cu.Customize != nil {
			cu.Customize(&ror)
		}

		if cu.IsDefault {
			o.Responses.Default = &ror
		} else {
			o.Responses.WithMapOfResponseOrReferenceValuesItem(httpStatus, ror)
		}
	}

	return nil
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
	sch, err := internal.ReflectJSONResponse(
		r.JSONSchemaReflector(),
		cu.Structure,
		openapi.WithOperationCtx(oc, true, openapi.InBody),
		jsonschema.DefinitionsPrefix(componentsSchemas),
		jsonschema.CollectDefinitions(r.collectDefinition()),
	)

	if err != nil || sch == nil {
		return err
	}

	sm, err := sch.ToSchemaOrBool().ToSimpleMap()
	if err != nil {
		return err
	}

	if resp.Content == nil {
		resp.Content = map[string]MediaType{}
	}

	contentType := cu.ContentType
	if contentType == "" {
		contentType = mimeJSON
	}

	mt := resp.Content[contentType]
	mt.Schema = sm

	resp.Content[contentType] = mt

	if sch.Description != nil && resp.Description == "" {
		resp.Description = *sch.Description
	}

	return nil
}

// SpecSchema returns OpenAPI spec schema.
func (r *Reflector) SpecSchema() openapi.SpecSchema {
	return r.SpecEns()
}

// JSONSchemaReflector provides access to a low-level struct reflector.
func (r *Reflector) JSONSchemaReflector() *jsonschema.Reflector {
	return &r.Reflector
}
