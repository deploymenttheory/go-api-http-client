package httpclient

import (
	"net/http"
)

// DoRequestNoRetry executes an HTTP request exactly once, bypassing the retry
// machinery that DoRequest applies to idempotent methods.
//
// DoRequest treats GET, PUT, DELETE, HEAD, OPTIONS and TRACE as safe to replay
// and retries them on transient failures (500, 502, 503, 504) and on 429. That
// holds for most of an API surface, but not for endpoints using optimistic
// locking, where the request body carries a version token the server consumes
// on the first successful write.
//
// For those endpoints a replay is unsafe in a way retries cannot detect. If the
// original request reached the server and was applied, but the response was
// lost or reported a fault, the retry resubmits a body whose token has already
// been spent and is rejected with a conflict — turning a write that succeeded
// into a reported conflict, and leaving the caller unable to tell that outcome
// apart from a genuine concurrent modification.
//
// Callers performing such writes should use this method and decide for
// themselves whether to retry, having first re-read the resource to obtain a
// current token.
//
// Behaviour is otherwise identical to DoRequest: the same concurrency control,
// authentication, and response handling apply. The caller is responsible for
// closing the response body when it is non-nil.
func (c *Client) DoRequestNoRetry(method, endpoint string, body, out any) (*http.Response, error) {
	return c.requestNoRetries(method, endpoint, body, out)
}
