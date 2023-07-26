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
	if len(operation.Responses.MapOfResponseOrRefValues) == 0 {
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
