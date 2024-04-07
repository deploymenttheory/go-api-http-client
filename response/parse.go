// response/parse.go
package response

import "strings"

// ParseContentTypeHeader parses the Content-Type header and returns the MIME type and any parameters.
func ParseContentTypeHeader(header string) (string, map[string]string) {
	return parseHeader(header)
}

// ParseContentDisposition parses the Content-Disposition header and returns the type and any parameters.
func ParseContentDisposition(header string) (string, map[string]string) {
	return parseHeader(header)
}

// parseHeader generalizes the parsing of headers like Content-Type and Content-Disposition.
// It extracts the main value (e.g., MIME type for Content-Type) and any parameters (like charset).
func parseHeader(header string) (string, map[string]string) {
	parts := strings.SplitN(header, ";", 2) // Split into main value and parameters
	mainValue := strings.TrimSpace(parts[0])

	params := make(map[string]string)
	if len(parts) > 1 { // Check if there are parameters
		for _, part := range strings.Split(parts[1], ";") {
			kv := strings.SplitN(part, "=", 2)
			if len(kv) == 2 {
				params[strings.TrimSpace(kv[0])] = strings.Trim(strings.TrimSpace(kv[1]), "\"")
			}
		}
	}

	return mainValue, params
}
