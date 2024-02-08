package httpclient

import (
	"fmt"
	"net/http"
	"strings"
)

// HeadersToString converts an http.Header map to a single string representation.
func HeadersToString(headers http.Header) string {
	var headerStrings []string

	// Iterate over the map and append each key-value pair to the slice
	for name, values := range headers {
		// Combine each header's key with its values, which are joined by a comma
		headerStrings = append(headerStrings, fmt.Sprintf("%s: %s", name, strings.Join(values, ", ")))
	}

	// Join all header strings into a single string
	return strings.Join(headerStrings, "; ")
}
