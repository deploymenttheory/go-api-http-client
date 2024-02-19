package httpclient

import "net/http"

// IsIdempotentHTTPMethod checks if the given HTTP method is idempotent.
func IsIdempotentHTTPMethod(method string) bool {
	idempotentHTTPMethods := map[string]bool{
		http.MethodGet:    true,
		http.MethodPut:    true,
		http.MethodDelete: true,
	}

	return idempotentHTTPMethods[method]
}

// IsNonIdempotentHTTPMethod checks if the given HTTP method is non-idempotent.
// PATCH can be idempotent but often isn't used as such.
func IsNonIdempotentHTTPMethod(method string) bool {
	nonIdempotentHTTPMethods := map[string]bool{
		http.MethodPost:  true,
		http.MethodPatch: true,
	}

	return nonIdempotentHTTPMethods[method]
}
