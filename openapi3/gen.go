package openapi3

import (
	"mime/multipart"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/swaggest/jsonschema-go"
	"github.com/swaggest/jsonschema-go/refl"
)

type Generator struct {
	jsonschema.Generator
	Spec *Spec
}

func (g *Generator) SetRequest(o *Operation, input interface{}, httpMethod string) error {
	return refl.JoinErrors(
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

func (g *Generator) parseRequestBody(o *Operation, input interface{}, tag, mime string, httpMethod string) error {
	httpMethod = strings.ToUpper(httpMethod)

	if httpMethod == http.MethodGet || httpMethod == http.MethodHead || !refl.HasTaggedFields(input, tag) {
		return nil
	}

	hasFileUpload := false
	definitionPefix := ""
	if tag != "json" {
		definitionPefix += strings.Title(tag)
	}
	schema, err := g.Parse(input,
		jsonschema.DefinitionsPrefix("#/components/schemas/"+definitionPefix),
		jsonschema.PropertyNameTag(tag),
		jsonschema.HijackType(func(v reflect.Value, s *jsonschema.Schema) (bool, error) {
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

func (g *Generator) parseParametersIn(o *Operation, input interface{}, in ParameterIn) error {
	var jpc *jsonschema.ParseContext

	schema, err := g.Parse(input,
		jsonschema.DefinitionsPrefix("#/components/schemas/"),
		jsonschema.InlineRoot,
		jsonschema.PropertyNameTag(string(in)),
		func(pc *jsonschema.ParseContext) { jpc = pc },
	)
	if err != nil {
		return err
	}

	required := map[string]bool{}
	for _, name := range schema.Required {
		required[name] = true
	}

	for _, name := range jpc.WalkedProperties {
		prop, ok := schema.Properties[name]
		if !ok {
			continue
		}

		s := SchemaOrRef{}
		s.FromJSONSchema(prop)

		p := ParameterOrRef{
			Parameter: &Parameter{
				Name:            name,
				In:              in,
				Description:     prop.TypeObject.Description,
				Required:        nil,
				AllowEmptyValue: nil,
				Style:           nil,
				Explode:         nil,
				AllowReserved:   nil,
				Schema:          &s,
				Content:         nil,
				Example:         nil,
				Examples:        nil,
				Location:        nil,
				MapOfAnything:   nil,
			},
		}

		if s.Schema != nil {
			p.Parameter.Deprecated = s.Schema.Deprecated
		}

		if in == ParameterInPath || required[name] {
			p.Parameter.WithRequired(true)
		}

		o.Parameters = append(o.Parameters, p)
	}

	return nil
}

func (g *Generator) parseResponseHeader(output interface{}) (map[string]HeaderOrRef, error) {
	var jpc *jsonschema.ParseContext

	schema, err := g.Parse(output,
		jsonschema.DefinitionsPrefix("#/components/headers/"),
		jsonschema.InlineRoot,
		jsonschema.PropertyNameTag("header"),
		func(pc *jsonschema.ParseContext) { jpc = pc },
	)
	if err != nil {
		return nil, err
	}

	required := map[string]bool{}
	for _, name := range schema.Required {
		required[name] = true
	}

	res := make(map[string]HeaderOrRef, len(schema.Properties))

	for _, name := range jpc.WalkedProperties {
		prop, ok := schema.Properties[name]
		if !ok {
			continue
		}

		s := SchemaOrRef{}
		s.FromJSONSchema(prop)

		header := Header{
			Description:     prop.TypeObject.Description,
			Deprecated:      s.Schema.Deprecated,
			AllowEmptyValue: nil,
			Explode:         nil,
			AllowReserved:   nil,
			Schema:          &s,
			Content:         nil,
			Example:         nil,
			Examples:        nil,
			MapOfAnything:   nil,
		}

		if required[name] {
			header.WithRequired(true)
		}

		res[name] = HeaderOrRef{
			Header: &header,
		}
	}

	return res, nil
}

func (g *Generator) SetJSONResponse(o *Operation, output interface{}, httpStatus int) error {
	schema, err := g.Parse(output, jsonschema.DefinitionsPrefix("#/components/schemas/"))
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
