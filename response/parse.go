// response/parse.go
package response

import "strings"

// parseHeader generalizes the parsing of headers like Content-Type and Content-Disposition.
// It extracts the main value (e.g., MIME type for Content-Type) and any parameters (like charset).
// TODO we need to talk about what this does.
func parseDispositionHeader(header string) (string, map[string]string) {
	parts := strings.SplitN(header, ";", 2)
	mainValue := strings.TrimSpace(parts[0])

	params := make(map[string]string)
	if len(parts) > 1 {
		for _, part := range strings.Split(parts[1], ";") {
			kv := strings.SplitN(part, "=", 2)
			if len(kv) == 2 {
				params[strings.TrimSpace(kv[0])] = strings.Trim(strings.TrimSpace(kv[1]), "\"")
			}
		}
	}

	return mainValue, params
}
