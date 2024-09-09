package openapi3

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/swaggest/openapi-go"
)

// ToParameterOrRef exposes Parameter in general form.
func (p Parameter) ToParameterOrRef() ParameterOrRef {
	return ParameterOrRef{
		Parameter: &p,
	}
}

// WithOperation sets Operation to PathItem.
//
// Deprecated: use Spec.AddOperation.
func (p *PathItem) WithOperation(method string, operation Operation) *PathItem {
	return p.WithMapOfOperationValuesItem(strings.ToLower(method), operation)
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

// SetupOperation creates operation if it is not present and applies setup functions.
func (s *Spec) SetupOperation(method, path string, setup ...func(*Operation) error) error {
	method, path, pathParams, err := openapi.SanitizeMethodPath(method, path)
	if err != nil {
		return err
	}

	pathItem := s.Paths.MapOfPathItemValues[path]
	operation := pathItem.MapOfOperationValues[method]

	for _, f := range setup {
		if err := f(&operation); err != nil {
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

	pathItem.WithMapOfOperationValuesItem(method, operation)

	s.Paths.WithMapOfPathItemValuesItem(path, pathItem)

	return nil
}

// AddOperation validates and sets operation by path and method.
//
// It will fail if operation with method and path already exists.
func (s *Spec) AddOperation(method, path string, operation Operation) error {
	method = strings.ToLower(method)
	pathItem := s.Paths.MapOfPathItemValues[path]

	if _, found := pathItem.MapOfOperationValues[method]; found {
		return fmt.Errorf("operation already exists: %s %s", method, path)
	}

	// Add "No Content" response if there are no responses configured.
	if len(operation.Responses.MapOfResponseOrRefValues) == 0 && operation.Responses.Default == nil {
		operation.Responses.WithMapOfResponseOrRefValuesItem(strconv.Itoa(http.StatusNoContent), ResponseOrRef{
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
	s.ComponentsEns().SecuritySchemesEns().WithMapOfSecuritySchemeOrRefValuesItem(
		securityName,
		SecuritySchemeOrRef{
			SecurityScheme: &SecurityScheme{
				HTTPSecurityScheme: (&HTTPSecurityScheme{}).WithScheme("basic").WithDescription(description),
			},
		},
	)
}

// SetAPIKeySecurity sets security definition.
func (s *Spec) SetAPIKeySecurity(securityName string, fieldName string, fieldIn openapi.In, description string) {
	s.ComponentsEns().SecuritySchemesEns().WithMapOfSecuritySchemeOrRefValuesItem(
		securityName,
		SecuritySchemeOrRef{
			SecurityScheme: &SecurityScheme{
				APIKeySecurityScheme: (&APIKeySecurityScheme{}).
					WithName(fieldName).
					WithIn(APIKeySecuritySchemeIn(fieldIn)).
					WithDescription(description),
			},
		},
	)
}

// SetHTTPBearerTokenSecurity sets security definition.
func (s *Spec) SetHTTPBearerTokenSecurity(securityName string, format string, description string) {
	ss := &SecurityScheme{
		HTTPSecurityScheme: (&HTTPSecurityScheme{}).
			WithScheme("bearer").
			WithDescription(description),
	}

	if format != "" {
		ss.HTTPSecurityScheme.WithBearerFormat(format)
	}

	s.ComponentsEns().SecuritySchemesEns().WithMapOfSecuritySchemeOrRefValuesItem(
		securityName,
		SecuritySchemeOrRef{
			SecurityScheme: ss,
		},
	)
}

// SetReference sets a reference and discards existing content.
func (r *ResponseOrRef) SetReference(ref string) {
	r.ResponseReferenceEns().Ref = ref
	r.Response = nil
}

// SetReference sets a reference and discards existing content.
func (r *RequestBodyOrRef) SetReference(ref string) {
	r.RequestBodyReferenceEns().Ref = ref
	r.RequestBody = nil
}
