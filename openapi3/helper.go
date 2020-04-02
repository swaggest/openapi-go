package openapi3

import "strings"

func (p *PathItem) WithOperation(method string, operation Operation) *PathItem {
	return p.WithMapOfOperationValuesItem(strings.ToLower(method), operation)
}
