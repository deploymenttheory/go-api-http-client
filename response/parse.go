// response/parse.go
package response

import "strings"

// ParseContentTypeHeader parses the Content-Type header and extracts the MIME type and parameters.
func ParseContentTypeHeader(header string) (string, map[string]string) {
	parts := strings.Split(header, ";")
	mimeType := strings.TrimSpace(parts[0])
	params := make(map[string]string)
	for _, part := range parts[1:] {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) == 2 {
			params[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
		}
	}
	return mimeType, params
}
