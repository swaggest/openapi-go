package openapi3

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
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

var regexFindPathParameter = regexp.MustCompile(`{([^}:]+)(:[^/]+)?(?:})`)

// SetupOperation creates operation if it is not present and applies setup functions.
func (s *Spec) SetupOperation(method, path string, setup ...func(*Operation) error) error {
	method = strings.ToLower(method)
	pathParametersSubmatches := regexFindPathParameter.FindAllStringSubmatch(path, -1)

	switch method {
	case "get", "put", "post", "delete", "options", "head", "patch", "trace":
		break
	default:
		return fmt.Errorf("unexpected http method: %s", method)
	}

	pathItem := s.Paths.MapOfPathItemValues[path]
	pathParams := map[string]bool{}

	if len(pathParametersSubmatches) > 0 {
		for _, submatch := range pathParametersSubmatches {
			pathParams[submatch[1]] = true

			if submatch[2] != "" { // Remove gorilla.Mux-style regexp in path
				path = strings.Replace(path, submatch[0], "{"+submatch[1]+"}", 1)
			}
		}
	}

	var errs []string

	operation := pathItem.MapOfOperationValues[method]

	for _, f := range setup {
		err := f(&operation)
		if err != nil {
			return err
		}
	}

	paramIndex := make(map[string]bool, len(operation.Parameters))

	for _, p := range operation.Parameters {
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
		return errors.New("operation with method and path already exists")
	}

	return s.SetupOperation(method, path, func(op *Operation) error {
		*op = operation
		return nil
	})
}

// WithHTTPBearerSecurity adds HTTP Bearer security definition.
func (s *Spec) WithHTTPBearerSecurity(name, bearerFormat, description string) *Spec {
	hss := HTTPSecurityScheme{}

	hss.WithScheme("bearer")

	if bearerFormat != "" {
		hss.WithBearerFormat(bearerFormat)
	}

	if description != "" {
		hss.WithDescription(description)
	}

	s.ComponentsEns().SecuritySchemesEns().WithMapOfSecuritySchemeOrRefValuesItem(
		name,
		SecuritySchemeOrRef{
			SecurityScheme: &SecurityScheme{
				HTTPSecurityScheme: &hss,
			},
		},
	)

	return s
}

// WithHTTPBasicSecurity adds HTTP Basic Auth security definition.
func (s *Spec) WithHTTPBasicSecurity(name, description string) *Spec {
	hss := HTTPSecurityScheme{}

	hss.WithScheme("basic")

	if description != "" {
		hss.WithDescription(description)
	}

	s.ComponentsEns().SecuritySchemesEns().WithMapOfSecuritySchemeOrRefValuesItem(
		name,
		SecuritySchemeOrRef{
			SecurityScheme: &SecurityScheme{
				HTTPSecurityScheme: &hss,
			},
		},
	)

	return s
}
