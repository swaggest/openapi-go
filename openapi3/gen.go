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

func (g *Reflector) SetRequest(o *Operation, input interface{}, httpMethod string) error {
	return joinErrors(
		g.parseParametersIn(o, input, ParameterInQuery),
		g.parseParametersIn(o, input, ParameterInPath),
		g.parseParametersIn(o, input, ParameterInCookie),
		g.parseParametersIn(o, input, ParameterInHeader),
		g.parseRequestBody(o, input, "json", mimeJSON, httpMethod),
		g.parseRequestBody(o, input, "formData", mimeFormUrlencoded, httpMethod),
	)
}

var (
	typeOfMultipartFile       = reflect.TypeOf((*multipart.File)(nil)).Elem()
	typeOfMultipartFileHeader = reflect.TypeOf((*multipart.FileHeader)(nil)).Elem()
)

const (
	mimeJSON           = "application/json"
	mimeFormUrlencoded = "application/x-www-form-urlencoded"
	mimeMultipart      = "multipart/form-data"
)

func (g *Reflector) parseRequestBody(o *Operation, input interface{}, tag, mime string, httpMethod string) error {
	httpMethod = strings.ToUpper(httpMethod)

	if httpMethod == http.MethodGet || httpMethod == http.MethodHead || !refl.HasTaggedFields(input, tag) {
		return nil
	}

	hasFileUpload := false
	definitionPefix := ""
	if tag != "json" {
		definitionPefix += strings.Title(tag)
	}
	schema, err := g.Reflect(input,
		jsonschema.DefinitionsPrefix("#/components/schemas/"+definitionPefix),
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

	mt := MediaType{
		Schema: &SchemaOrRef{
			SchemaReference: &SchemaReference{Ref: *schema.Ref},
		},
	}

	for name, def := range schema.Definitions {
		if g.Spec.Components == nil {
			g.Spec.Components = &Components{}
		}
		if g.Spec.Components.Schemas == nil {
			g.Spec.Components.Schemas = &ComponentsSchemas{}
		}
		s := SchemaOrRef{}
		s.FromJSONSchema(def)

		g.Spec.Components.Schemas.WithMapOfSchemaOrRefValuesItem(definitionPefix+name, s)
	}

	if o.RequestBody == nil {
		o.RequestBody = &RequestBodyOrRef{}
	}

	if o.RequestBody.RequestBody == nil {
		o.RequestBody.RequestBody = &RequestBody{}
	}

	if o.RequestBody.RequestBody.Content == nil {
		o.RequestBody.RequestBody.Content = map[string]MediaType{}
	}

	if mime == mimeFormUrlencoded && hasFileUpload {
		mime = mimeMultipart
	}

	o.RequestBody.RequestBody.Content[mime] = mt

	return nil
}

func (g *Reflector) parseParametersIn(o *Operation, input interface{}, in ParameterIn) error {
	_, err := g.Reflect(input,
		jsonschema.DefinitionsPrefix("#/components/schemas/"),
		jsonschema.InlineRoot,
		jsonschema.PropertyNameTag(string(in)),
		func(pc *jsonschema.ReflectContext) {
			pc.InterceptProperty = func(name string, field reflect.StructField, propertySchema *jsonschema.Schema) error {
				s := SchemaOrRef{}
				s.FromJSONSchema(propertySchema.ToSchemaOrBool())

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
			}
		},
	)
	if err != nil {
		return err
	}

	return nil
}

func (g *Reflector) parseResponseHeader(output interface{}) (map[string]HeaderOrRef, error) {
	res := make(map[string]HeaderOrRef)

	_, err := g.Reflect(output,
		jsonschema.DefinitionsPrefix("#/components/headers/"),
		jsonschema.InlineRoot,
		jsonschema.PropertyNameTag("header"),
		func(pc *jsonschema.ReflectContext) {
			pc.InterceptProperty = func(name string, field reflect.StructField, propertySchema *jsonschema.Schema) error {
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
			}
		},
	)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (g *Reflector) SetJSONResponse(o *Operation, output interface{}, httpStatus int) error {
	schema, err := g.Reflect(output, jsonschema.DefinitionsPrefix("#/components/schemas/"))
	if err != nil {
		return err
	}

	if o.Responses.MapOfResponseOrRefValues == nil {
		o.Responses.MapOfResponseOrRefValues = make(map[string]ResponseOrRef, 1)
	}

	oaiSchema := SchemaOrRef{}
	oaiSchema.FromJSONSchema(schema.ToSchemaOrBool())

	resp := Response{
		Content: map[string]MediaType{
			"application/json": {
				Schema:        &oaiSchema,
				Example:       nil,
				Examples:      nil,
				Encoding:      nil,
				MapOfAnything: nil,
			},
		},
	}

	resp.Headers, err = g.parseResponseHeader(output)
	if err != nil {
		return err
	}

	for name, def := range schema.Definitions {
		if g.Spec.Components == nil {
			g.Spec.Components = &Components{}
		}
		if g.Spec.Components.Schemas == nil {
			g.Spec.Components.Schemas = &ComponentsSchemas{}
		}
		s := SchemaOrRef{}
		s.FromJSONSchema(def)

		g.Spec.Components.Schemas.WithMapOfSchemaOrRefValuesItem(name, s)
	}

	o.Responses.MapOfResponseOrRefValues[strconv.Itoa(httpStatus)] = ResponseOrRef{
		Response: &resp,
	}

	return nil
}
