package openapi31

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/swaggest/openapi-go"
)

// ToParameterOrRef exposes Parameter in general form.
func (p Parameter) ToParameterOrRef() ParameterOrReference {
	return ParameterOrReference{
		Parameter: &p,
	}
}

// Operation retrieves method Operation from PathItem.
func (p PathItem) Operation(method string) (*Operation, error) {
	switch strings.ToUpper(method) {
	case http.MethodGet:
		return p.Get, nil
	case http.MethodPut:
		return p.Put, nil
	case http.MethodPost:
		return p.Post, nil
	case http.MethodDelete:
		return p.Delete, nil
	case http.MethodOptions:
		return p.Options, nil
	case http.MethodHead:
		return p.Head, nil
	case http.MethodPatch:
		return p.Patch, nil
	case http.MethodTrace:
		return p.Trace, nil
	default:
		return nil, fmt.Errorf("unexpected http method: %s", method)
	}
}

// SetOperation sets a method Operation to PathItem.
func (p *PathItem) SetOperation(method string, op *Operation) error {
	method = strings.ToUpper(method)

	switch method {
	case http.MethodGet:
		p.Get = op
	case http.MethodPut:
		p.Put = op
	case http.MethodPost:
		p.Post = op
	case http.MethodDelete:
		p.Delete = op
	case http.MethodOptions:
		p.Options = op
	case http.MethodHead:
		p.Head = op
	case http.MethodPatch:
		p.Patch = op
	case http.MethodTrace:
		p.Trace = op
	default:
		return fmt.Errorf("unexpected http method: %s", method)
	}

	return nil
}

// SetupOperation creates operation if it is not present and applies setup functions.
func (s *Spec) SetupOperation(method, path string, setup ...func(*Operation) error) error {
	method, path, pathParams, err := openapi.SanitizeMethodPath(method, path)
	if err != nil {
		return err
	}

	pathItem := s.PathsEns().MapOfPathItemValues[path]

	operation, err := pathItem.Operation(method)
	if err != nil {
		return err
	}

	if operation == nil {
		operation = &Operation{}
	}

	for _, f := range setup {
		if err := f(operation); err != nil {
			return err
		}
	}

	pathParamsMap := make(map[string]bool, len(pathParams))
	for _, p := range pathParams {
		pathParamsMap[p] = true
	}

	if err := operation.validatePathParams(pathParamsMap); err != nil {
		return err
	}

	if err := pathItem.SetOperation(method, operation); err != nil {
		return err
	}

	s.PathsEns().WithMapOfPathItemValuesItem(path, pathItem)

	return nil
}

func (o *Operation) validatePathParams(pathParams map[string]bool) error {
	paramIndex := make(map[string]bool, len(o.Parameters))

	var errs []string

	for _, p := range o.Parameters {
		if p.Parameter == nil {
			continue
		}

		if found := paramIndex[p.Parameter.Name+string(p.Parameter.In)]; found {
			errs = append(errs, "duplicate parameter in "+string(p.Parameter.In)+": "+p.Parameter.Name)

			continue
		}

		if found := pathParams[p.Parameter.Name]; !found && p.Parameter.In == ParameterInPath {
			errs = append(errs, "missing path parameter placeholder in url: "+p.Parameter.Name)

			continue
		}

		paramIndex[p.Parameter.Name+string(p.Parameter.In)] = true
	}

	for pathParam := range pathParams {
		if !paramIndex[pathParam+string(ParameterInPath)] {
			errs = append(errs, "undefined path parameter: "+pathParam)
		}
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, ", "))
	}

	return nil
}

// AddOperation validates and sets operation by path and method.
//
// It will fail if operation with method and path already exists.
func (s *Spec) AddOperation(method, path string, operation Operation) error {
	method = strings.ToLower(method)

	pathItem := s.PathsEns().MapOfPathItemValues[path]

	op, err := pathItem.Operation(method)
	if err != nil {
		return err
	}

	if op != nil {
		return fmt.Errorf("operation already exists: %s %s", method, path)
	}

	// Add "No Content" response if there are no responses configured.
	if len(operation.ResponsesEns().MapOfResponseOrReferenceValues) == 0 && operation.Responses.Default == nil {
		operation.Responses.WithMapOfResponseOrReferenceValuesItem(strconv.Itoa(http.StatusNoContent), ResponseOrReference{
			Response: &Response{
				Description: http.StatusText(http.StatusNoContent),
			},
		})
	}

	return s.SetupOperation(method, path, func(op *Operation) error {
		*op = operation

		return nil
	})
}

// UnknownParamIsForbidden indicates forbidden unknown parameters.
func (o Operation) UnknownParamIsForbidden(in ParameterIn) bool {
	f, ok := o.MapOfAnything[xForbidUnknown+string(in)].(bool)

	return f && ok
}

var _ openapi.SpecSchema = &Spec{}

// Title returns service title.
func (s *Spec) Title() string {
	return s.Info.Title
}

// SetTitle describes the service.
func (s *Spec) SetTitle(t string) {
	s.Info.Title = t
}

// Description returns service description.
func (s *Spec) Description() string {
	if s.Info.Description != nil {
		return *s.Info.Description
	}

	return ""
}

// SetDescription describes the service.
func (s *Spec) SetDescription(d string) {
	s.Info.WithDescription(d)
}

// Version returns service version.
func (s *Spec) Version() string {
	return s.Info.Version
}

// SetVersion describes the service.
func (s *Spec) SetVersion(v string) {
	s.Info.Version = v
}

// SetHTTPBasicSecurity sets security definition.
func (s *Spec) SetHTTPBasicSecurity(securityName string, description string) {
	s.ComponentsEns().WithSecuritySchemesItem(
		securityName,
		SecuritySchemeOrReference{
			SecurityScheme: (&SecurityScheme{
				HTTP: (&SecuritySchemeHTTP{}).WithScheme("basic"),
			}).WithDescription(description),
		},
	)
}

// SetAPIKeySecurity sets security definition.
func (s *Spec) SetAPIKeySecurity(securityName string, fieldName string, fieldIn openapi.In, description string) {
	s.ComponentsEns().WithSecuritySchemesItem(
		securityName,
		SecuritySchemeOrReference{
			SecurityScheme: (&SecurityScheme{
				APIKey: (&SecuritySchemeAPIKey{}).
					WithName(fieldName).
					WithIn(SecuritySchemeAPIKeyIn(fieldIn)),
			}).WithDescription(description),
		},
	)
}

// SetHTTPBearerTokenSecurity sets security definition.
func (s *Spec) SetHTTPBearerTokenSecurity(securityName string, format string, description string) {
	ss := (&SecurityScheme{
		HTTPBearer: (&SecuritySchemeHTTPBearer{}).
			WithScheme("bearer"),
	}).WithDescription(description)

	if format != "" {
		ss.HTTPBearer.WithBearerFormat(format)
	}

	s.ComponentsEns().WithSecuritySchemesItem(
		securityName,
		SecuritySchemeOrReference{
			SecurityScheme: ss,
		},
	)
}

// SetReference sets a reference and discards existing content.
func (r *ResponseOrReference) SetReference(ref string) {
	r.ReferenceEns().Ref = ref
	r.Response = nil
}

// SetReference sets a reference and discards existing content.
func (r *RequestBodyOrReference) SetReference(ref string) {
	r.ReferenceEns().Ref = ref
	r.RequestBody = nil
}
