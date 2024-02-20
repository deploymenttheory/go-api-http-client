// httpclient_methods.go
package httpclient

/* Ref: https://www.rfc-editor.org/rfc/rfc7231#section-8.1.3

+---------+------+------------+
| Method  | Safe | Idempotent |
+---------+------+------------+
| CONNECT | no   | no         |
| DELETE  | no   | yes        |
| GET     | yes  | yes        |
| HEAD    | yes  | yes        |
| OPTIONS | yes  | yes        |
| POST    | no   | no         |
| PUT     | no   | yes        |
| TRACE   | yes  | yes        |
+---------+------+------------+
*/

import "net/http"

// IsIdempotentHTTPMethod checks if the given HTTP method is idempotent.
func IsIdempotentHTTPMethod(method string) bool {
	idempotentHTTPMethods := map[string]bool{
		http.MethodGet:     true,
		http.MethodPut:     true,
		http.MethodDelete:  true,
		http.MethodHead:    true,
		http.MethodOptions: true,
		http.MethodTrace:   true,
	}

	return idempotentHTTPMethods[method]
}

// IsNonIdempotentHTTPMethod checks if the given HTTP method is non-idempotent.
// PATCH can be idempotent but often isn't used as such.
func IsNonIdempotentHTTPMethod(method string) bool {
	nonIdempotentHTTPMethods := map[string]bool{
		http.MethodPost:    true,
		http.MethodPatch:   true,
		http.MethodConnect: true,
	}

	return nonIdempotentHTTPMethods[method]
}
