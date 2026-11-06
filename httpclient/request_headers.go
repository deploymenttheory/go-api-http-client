package httpclient

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/net/http/httpguts"
)

var (
	errInvalidRequestHeader      = errors.New("invalid custom request header")
	errConflictingRequestHeader  = errors.New("custom header conflicts with managed bearer-token header")
	errInvalidHeaderOrigin       = errors.New("configured request headers require an HTTP or HTTPS base URL")
	errCrossOriginHeaderRedirect = errors.New("cross-origin redirect blocked for configured request headers")
	errTooManyHeaderRedirects    = errors.New("stopped after 10 redirects")
)

// RequestHeaderConfig configures headers for an API origin, including token bootstrap.
// BearerTokenHeader defaults to Authorization. A different name frees Authorization
// for a reverse proxy's Basic credentials without freezing the API's refreshed token.
type RequestHeaderConfig struct {
	CustomHeaders     map[string]string
	BearerTokenHeader string
}

// HTTPClient returns a copy of client with this origin's header policy installed.
// Existing transports, TLS settings, cookies and timeouts are preserved. Use it for
// bootstrap requests as well as ClientConfig.RequestHeaders for normal API requests.
func (c RequestHeaderConfig) HTTPClient(client http.Client, baseURL string) (http.Client, error) {
	tokenHeader := c.BearerTokenHeader
	if tokenHeader == "" {
		tokenHeader = "Authorization"
	}
	if !httpguts.ValidHeaderFieldName(tokenHeader) {
		return http.Client{}, fmt.Errorf("%w: bearer-token header name", errInvalidRequestHeader)
	}
	tokenHeader = http.CanonicalHeaderKey(tokenHeader)
	headers := make(http.Header, len(c.CustomHeaders))
	for name, value := range c.CustomHeaders {
		if !httpguts.ValidHeaderFieldName(name) || !httpguts.ValidHeaderFieldValue(value) {
			return http.Client{}, errInvalidRequestHeader
		}
		name = http.CanonicalHeaderKey(name)
		if name == tokenHeader {
			return http.Client{}, fmt.Errorf("%w: %s", errConflictingRequestHeader, name)
		}
		if _, exists := headers[name]; exists {
			return http.Client{}, fmt.Errorf("%w: duplicate header name %s", errInvalidRequestHeader, name)
		}
		headers.Set(name, value)
	}
	if len(headers) == 0 && tokenHeader == "Authorization" {
		return client, nil
	}
	origin, err := url.Parse(baseURL)
	if err != nil || origin.Host == "" || (origin.Scheme != "http" && origin.Scheme != "https") {
		return http.Client{}, errInvalidHeaderOrigin
	}
	base := client.Transport
	if base == nil {
		base = http.DefaultTransport
	}
	client.Transport = &requestHeaderTransport{base: base, origin: origin, headers: headers, tokenHeader: tokenHeader}
	previousRedirect := client.CheckRedirect
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		// Go does not recognize a custom token header as an authentication header.
		// Never let a redirect forward it or the proxy credentials to another origin.
		if !sameHeaderOrigin(origin, req.URL) {
			return errCrossOriginHeaderRedirect
		}
		if previousRedirect != nil {
			return previousRedirect(req, via)
		}
		if len(via) >= 10 {
			return errTooManyHeaderRedirects
		}
		return nil
	}
	return client, nil
}

type requestHeaderTransport struct {
	base        http.RoundTripper
	origin      *url.URL
	headers     http.Header
	tokenHeader string
}

func (t *requestHeaderTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if !sameHeaderOrigin(t.origin, req.URL) {
		return t.base.RoundTrip(req)
	}
	request := req.Clone(req.Context())
	scheme, _, _ := strings.Cut(request.Header.Get("Authorization"), " ")
	if t.tokenHeader != "Authorization" && strings.EqualFold(scheme, "Bearer") {
		request.Header.Set(t.tokenHeader, request.Header.Get("Authorization"))
		request.Header.Del("Authorization")
		// net/http applies URL Basic auth before RoundTrip, when Authorization
		// still held the bearer token. Restore the proxy credentials after moving it.
		if request.URL.User != nil {
			password, _ := request.URL.User.Password()
			request.SetBasicAuth(request.URL.User.Username(), password)
		}
	}
	for name, values := range t.headers {
		request.Header[name] = append([]string(nil), values...)
	}
	resp, err := t.base.RoundTrip(request)
	if err != nil {
		return resp, fmt.Errorf("send request with configured headers: %w", err)
	}
	return resp, nil
}

func (t *requestHeaderTransport) CloseIdleConnections() {
	if transport, ok := t.base.(interface{ CloseIdleConnections() }); ok {
		transport.CloseIdleConnections()
	}
}

func sameHeaderOrigin(a, b *url.URL) bool {
	return strings.EqualFold(a.Scheme, b.Scheme) && strings.EqualFold(a.Host, b.Host)
}
