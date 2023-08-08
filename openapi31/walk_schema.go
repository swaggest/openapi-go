package openapi31

import (
	"fmt"
	"net/http"

	"github.com/swaggest/jsonschema-go"
	"github.com/swaggest/openapi-go"
	"github.com/swaggest/openapi-go/internal"
)

// WalkResponseJSONSchemas provides JSON schemas for response structure.
func (r *Reflector) WalkResponseJSONSchemas(cu openapi.ContentUnit, cb openapi.JSONSchemaCallback, done func(oc openapi.OperationContext)) error {
	oc := operationContext{
		OperationContext: internal.NewOperationContext(http.MethodGet, "/"),
		op:               &Operation{},
	}

	oc.AddRespStructure(nil, func(c *openapi.ContentUnit) {
		*c = cu
	})

	defer func() {
		if done != nil {
			done(oc)
		}
	}()

	op := oc.op

	if err := r.setupResponse(op, oc); err != nil {
		return err
	}

	var resp *Response

	for _, r := range op.Responses.MapOfResponseOrReferenceValues {
		resp = r.Response
	}

	if err := r.provideHeaderSchemas(resp, cb); err != nil {
		return err
	}

	for _, cont := range resp.Content {
		if cont.Schema == nil {
			continue
		}

		sm := ToJSONSchema(cont.Schema, r.Spec)

		if err := cb(openapi.InBody, "body", &sm, false); err != nil {
			return fmt.Errorf("response body schema: %w", err)
		}
	}

	return nil
}

func (r *Reflector) provideHeaderSchemas(resp *Response, cb openapi.JSONSchemaCallback) error {
	for name, h := range resp.Headers {
		if h.Header.Schema == nil {
			continue
		}

		hh := h.Header
		schema := ToJSONSchema(hh.Schema, r.Spec)

		required := false
		if hh.Required != nil && *hh.Required {
			required = true
		}

		name = http.CanonicalHeaderKey(name)

		if err := cb(openapi.InHeader, name, &schema, required); err != nil {
			return fmt.Errorf("response header schema (%s): %w", name, err)
		}
	}

	return nil
}

// WalkRequestJSONSchemas iterates over request parameters of a ContentUnit and call user function for param schemas.
func (r *Reflector) WalkRequestJSONSchemas(
	method string,
	cu openapi.ContentUnit,
	cb openapi.JSONSchemaCallback,
	done func(oc openapi.OperationContext),
) error {
	oc := operationContext{
		OperationContext: internal.NewOperationContext(method, "/"),
		op:               &Operation{},
	}

	oc.AddReqStructure(nil, func(c *openapi.ContentUnit) {
		*c = cu
	})

	defer func() {
		if done != nil {
			done(oc)
		}
	}()

	op := oc.op

	if err := r.setupRequest(op, oc); err != nil {
		return err
	}

	err := r.provideParametersJSONSchemas(op, cb)
	if err != nil {
		return err
	}

	if op.RequestBody == nil || op.RequestBody.RequestBody == nil {
		return nil
	}

	for ct, content := range op.RequestBody.RequestBody.Content {
		schema := ToJSONSchema(content.Schema, r.Spec)

		if ct == mimeJSON {
			err = cb(openapi.InBody, "body", &schema, false)
			if err != nil {
				return fmt.Errorf("request body schema: %w", err)
			}
		}

		if ct == mimeFormUrlencoded {
			if err = provideFormDataSchemas(schema, cb); err != nil {
				return err
			}
		}
	}

	return nil
}

func provideFormDataSchemas(schema jsonschema.SchemaOrBool, cb openapi.JSONSchemaCallback) error {
	for name, propertySchema := range schema.TypeObject.Properties {
		propertySchema := propertySchema

		if propertySchema.TypeObject != nil && len(schema.TypeObject.ExtraProperties) > 0 {
			cp := *propertySchema.TypeObject
			propertySchema.TypeObject = &cp
			propertySchema.TypeObject.ExtraProperties = schema.TypeObject.ExtraProperties
		}

		isRequired := false

		for _, req := range schema.TypeObject.Required {
			if req == name {
				isRequired = true

				break
			}
		}

		err := cb(openapi.InFormData, name, &propertySchema, isRequired)
		if err != nil {
			return fmt.Errorf("request body schema: %w", err)
		}
	}

	return nil
}

func (r *Reflector) provideParametersJSONSchemas(op *Operation, cb openapi.JSONSchemaCallback) error {
	for _, p := range op.Parameters {
		pp := p.Parameter

		required := false
		if pp.Required != nil && *pp.Required {
			required = true
		}

		sc := paramSchema(pp)

		if sc == nil {
			if err := cb(openapi.In(pp.In), pp.Name, nil, required); err != nil {
				return fmt.Errorf("schema for parameter (%s, %s): %w", pp.In, pp.Name, err)
			}

			continue
		}

		schema := ToJSONSchema(sc, r.Spec)

		if err := cb(openapi.In(pp.In), pp.Name, &schema, required); err != nil {
			return fmt.Errorf("schema for parameter (%s, %s): %w", pp.In, pp.Name, err)
		}
	}

	return nil
}

func paramSchema(p *Parameter) map[string]interface{} {
	sc := p.Schema

	if sc == nil {
		if jsc, ok := p.Content[mimeJSON]; ok {
			sc = jsc.Schema
		}
	}

	return sc
}
