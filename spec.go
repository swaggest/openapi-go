package openapi

// SpecSchema abstracts OpenAPI schema implementation to generalize multiple revisions.
type SpecSchema interface {
	Title() string
	Description() string
	Version() string

	SetTitle(t string)
	SetDescription(d string)
	SetVersion(v string)

	SetHTTPBasicSecurity(securityName string, description string)
	SetAPIKeySecurity(securityName string, fieldName string, fieldIn In, description string)
	SetHTTPBearerTokenSecurity(securityName string, format string, description string)
}
