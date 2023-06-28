package internal

import (
	"fmt"
	"regexp"
	"strings"
)

var regexFindPathParameter = regexp.MustCompile(`{([^}:]+)(:[^}]+)?(?:})`)

// SanitizeMethodPath validates method and parses path element names.
func SanitizeMethodPath(method, pathPattern string) (string, string, map[string]bool, error) {
	method = strings.ToLower(method)
	pathParametersSubmatches := regexFindPathParameter.FindAllStringSubmatch(pathPattern, -1)

	switch method {
	case "get", "put", "post", "delete", "options", "head", "patch", "trace":
		break
	default:
		return "", "", nil, fmt.Errorf("unexpected http method: %s", method)
	}

	pathParams := map[string]bool{}

	if len(pathParametersSubmatches) > 0 {
		for _, submatch := range pathParametersSubmatches {
			pathParams[submatch[1]] = true

			if submatch[2] != "" { // Remove gorilla.Mux-style regexp in path.
				pathPattern = strings.Replace(pathPattern, submatch[0], "{"+submatch[1]+"}", 1)
			}
		}
	}

	return method, pathPattern, pathParams, nil
}
