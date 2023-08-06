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

func (p PathItem) Operation(method string) (*Operation, error) {
	method = strings.ToUpper(method)

	switch method {
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

// SetTitle describes the service.
func (s *Spec) SetTitle(t string) {
	s.Info.Title = t
}

// SetDescription describes the service.
func (s *Spec) SetDescription(d string) {
	s.Info.WithDescription(d)
}

// SetVersion describes the service.
func (s *Spec) SetVersion(v string) {
	s.Info.Version = v
}
