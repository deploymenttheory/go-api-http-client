// httpmethod/httpmethod.go
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
	methods := map[string]bool{
		http.MethodGet:     true,
		http.MethodPut:     true,
		http.MethodDelete:  true,
		http.MethodHead:    true,
		http.MethodOptions: true,
		http.MethodTrace:   true,
		http.MethodPost:    false,
		http.MethodPatch:   false,
		http.MethodConnect: false,
	}

	return methods[method]
}
