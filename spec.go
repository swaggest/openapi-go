package openapi

// SpecSchema abstracts OpenAPI schema implementation to generalize multiple revisions.
type SpecSchema interface {
	SetTitle(t string)
	SetDescription(d string)
	SetVersion(v string)
}
