package openapi

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/swaggest/jsonschema-go"
)

// In defines value location in HTTP content.
type In string

// In values enumeration.
const (
	InPath     = In("path")
	InQuery    = In("query")
	InHeader   = In("header")
	InCookie   = In("cookie")
	InFormData = In("formData")
	InBody     = In("body")
)

// ContentOption configures ContentUnit.
type ContentOption func(cu *ContentUnit)

// ContentUnit defines HTTP content.
type ContentUnit struct {
	Structure   interface{}
	ContentType string
	Format      string

	// HTTPStatus can have values 100-599 for single status, or 1-5 for status families (e.g. 2XX)
	HTTPStatus int

	// IsDefault indicates default response.
	IsDefault bool

	Description string

	// Customize allows fine control over prepared content entities.
	// The cor value can be asserted to one of these types:
	// *openapi3.RequestBodyOrRef
	// *openapi3.ResponseOrRef
	// *openapi31.RequestBodyOrReference
	// *openapi31.ResponseOrReference
	Customize func(cor ContentOrReference)

	fieldMapping map[In]map[string]string
}

// ContentOrReference defines content entity that can be a reference.
type ContentOrReference interface {
	SetReference(ref string)
}

// WithCustomize is a ContentUnit option.
func WithCustomize(customize func(cor ContentOrReference)) ContentOption {
	return func(cu *ContentUnit) {
		cu.Customize = customize
	}
}

// WithReference is a ContentUnit option.
func WithReference(ref string) ContentOption {
	return func(cu *ContentUnit) {
		cu.Customize = func(cor ContentOrReference) {
			cor.SetReference(ref)
		}
	}
}

// ContentUnitPreparer defines self-contained ContentUnit.
type ContentUnitPreparer interface {
	SetupContentUnit(cu *ContentUnit)
}

// WithContentType is a ContentUnit option.
func WithContentType(contentType string) func(cu *ContentUnit) {
	return func(cu *ContentUnit) {
		cu.ContentType = contentType
	}
}

// WithHTTPStatus is a ContentUnit option.
func WithHTTPStatus(httpStatus int) func(cu *ContentUnit) {
	return func(cu *ContentUnit) {
		cu.HTTPStatus = httpStatus
	}
}

// SetFieldMapping sets custom field mapping.
func (c *ContentUnit) SetFieldMapping(in In, fieldToParamName map[string]string) {
	if len(fieldToParamName) == 0 {
		return
	}

	if c.fieldMapping == nil {
		c.fieldMapping = make(map[In]map[string]string)
	}

	c.fieldMapping[in] = fieldToParamName
}

// FieldMapping returns custom field mapping.
func (c ContentUnit) FieldMapping(in In) map[string]string {
	return c.fieldMapping[in]
}

// OperationContext defines operation and processing state.
type OperationContext interface {
	OperationInfo
	OperationInfoReader
	OperationState

	Method() string
	PathPattern() string

	Request() []ContentUnit
	Response() []ContentUnit

	AddReqStructure(i interface{}, options ...ContentOption)
	AddRespStructure(o interface{}, options ...ContentOption)

	UnknownParamsAreForbidden(in In) bool
}

// OperationInfo extends OperationContext with general information.
type OperationInfo interface {
	SetTags(tags ...string)
	SetIsDeprecated(isDeprecated bool)
	SetSummary(summary string)
	SetDescription(description string)
	SetID(operationID string)

	AddSecurity(securityName string, scopes ...string)
}

// OperationInfoReader exposes current state of operation context.
type OperationInfoReader interface {
	Tags() []string
	IsDeprecated() bool
	Summary() string
	Description() string
	ID() string
}

// OperationState extends OperationContext with processing state information.
type OperationState interface {
	IsProcessingResponse() bool
	ProcessingIn() In

	SetIsProcessingResponse(isProcessingResponse bool)
	SetProcessingIn(in In)
}

type ocCtxKey struct{}

// WithOperationCtx is a jsonschema.ReflectContext option.
func WithOperationCtx(oc OperationContext, isProcessingResponse bool, in In) func(rc *jsonschema.ReflectContext) {
	return func(rc *jsonschema.ReflectContext) {
		oc.SetIsProcessingResponse(isProcessingResponse)
		oc.SetProcessingIn(in)

		rc.Context = context.WithValue(rc.Context, ocCtxKey{}, oc)
	}
}

// OperationCtx retrieves operation context from reflect context.
func OperationCtx(rc *jsonschema.ReflectContext) (OperationContext, bool) {
	if oc, ok := rc.Value(ocCtxKey{}).(OperationContext); ok {
		return oc, true
	}

	return nil, false
}

var regexFindPathParameter = regexp.MustCompile(`{([^}:]+)(:[^}]+)?(?:})`)

// SanitizeMethodPath validates method and parses path element names.
func SanitizeMethodPath(method, pathPattern string) (cleanMethod string, cleanPath string, pathParams []string, err error) {
	method = strings.ToLower(method)
	pathParametersSubmatches := regexFindPathParameter.FindAllStringSubmatch(pathPattern, -1)

	switch method {
	case "get", "put", "post", "delete", "options", "head", "patch", "trace":
		break
	default:
		return "", "", nil, fmt.Errorf("unexpected http method: %s", method)
	}

	if len(pathParametersSubmatches) > 0 {
		for _, submatch := range pathParametersSubmatches {
			pathParams = append(pathParams, submatch[1])

			if submatch[2] != "" { // Remove gorilla.Mux-style regexp in path.
				pathPattern = strings.Replace(pathPattern, submatch[0], "{"+submatch[1]+"}", 1)
			}
		}
	}

	return method, pathPattern, pathParams, nil
}
