// headers/headers_test.go
package headers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/deploymenttheory/go-api-http-client/authenticationhandler"
	"github.com/stretchr/testify/assert"
)

// TestSetAuthorization verifies that the SetAuthorization method correctly sets
// the "Authorization" header of the HTTP request. The header should be prefixed
// with "Bearer " followed by the token provided by the authTokenHandler.
func TestSetAuthorization(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)

	token := "test-token"
	authTokenHandler := &authenticationhandler.AuthTokenHandler{Token: token}

	// Create HeaderHandler without a mock logger since logging is not being tested
	headerHandler := NewHeaderHandler(req, nil, nil, authTokenHandler)
	headerHandler.SetAuthorization()

	expectedHeaderValue := "Bearer " + token
	assert.Equal(t, expectedHeaderValue, req.Header.Get("Authorization"), "Authorization header should be correctly set")
}

// TestSetContentType verifies that the SetContentType method correctly sets
// the "Content-Type" header of the HTTP request. This header should reflect
// the content type passed to the method.
func TestSetContentType(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)

	contentType := "application/json"
	// Create HeaderHandler without a mock logger since logging is not being tested
	headerHandler := NewHeaderHandler(req, nil, nil, nil)
	headerHandler.SetContentType(contentType)

	assert.Equal(t, contentType, req.Header.Get("Content-Type"), "Content-Type header should be correctly set")
}

// TestSetAccept verifies that the SetAccept method correctly sets the "Accept"
// header of the HTTP request. This header indicates the media types that the
// client is willing to receive from the server.
func TestSetAccept(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)

	acceptHeader := "application/json"
	headerHandler := NewHeaderHandler(req, nil, nil, nil)
	headerHandler.SetAccept(acceptHeader)

	assert.Equal(t, acceptHeader, req.Header.Get("Accept"), "Accept header should be correctly set")
}

// TestSetUserAgent verifies that the SetUserAgent method correctly sets the
// "User-Agent" header of the HTTP request. This header should reflect the user
// agent string passed to the method.
func TestSetUserAgent(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)

	userAgent := "MyCustomUserAgent/1.0"
	headerHandler := NewHeaderHandler(req, nil, nil, nil)
	headerHandler.SetUserAgent(userAgent)

	assert.Equal(t, userAgent, req.Header.Get("User-Agent"), "User-Agent header should be correctly set")
}

// TestSetCacheControlHeader verifies that the SetCacheControlHeader function
// correctly sets the "Cache-Control" header of the HTTP request. This header
// contains directives for caching mechanisms in requests and responses.
func TestSetCacheControlHeader(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)

	cacheControlValue := "no-cache"
	SetCacheControlHeader(req, cacheControlValue)

	assert.Equal(t, cacheControlValue, req.Header.Get("Cache-Control"), "Cache-Control header should be correctly set")
}

func TestSetConditionalHeaders(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	ifModifiedSince := "Wed, 21 Oct 2015 07:28:00 GMT"
	ifNoneMatch := `"etag-value"`

	SetConditionalHeaders(req, ifModifiedSince, ifNoneMatch)

	assert.Equal(t, ifModifiedSince, req.Header.Get("If-Modified-Since"), "If-Modified-Since header should be correctly set")
	assert.Equal(t, ifNoneMatch, req.Header.Get("If-None-Match"), "If-None-Match header should be correctly set")
}

func TestSetAcceptEncodingHeader(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	acceptEncodingValue := "gzip, deflate"

	SetAcceptEncodingHeader(req, acceptEncodingValue)

	assert.Equal(t, acceptEncodingValue, req.Header.Get("Accept-Encoding"), "Accept-Encoding header should be correctly set")
}

func TestSetRefererHeader(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	refererValue := "http://previous-page.com"

	SetRefererHeader(req, refererValue)

	assert.Equal(t, refererValue, req.Header.Get("Referer"), "Referer header should be correctly set")
}

func TestSetXForwardedForHeader(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	xForwardedForValue := "123.45.67.89"

	SetXForwardedForHeader(req, xForwardedForValue)

	assert.Equal(t, xForwardedForValue, req.Header.Get("X-Forwarded-For"), "X-Forwarded-For header should be correctly set")
}

func TestSetCustomHeader(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	headerName := "X-Custom-Header"
	headerValue := "CustomValue"

	SetCustomHeader(req, headerName, headerValue)

	assert.Equal(t, headerValue, req.Header.Get(headerName), "Custom header should be correctly set")
}

// TestSetRequestHeaders verifies that standard headers, including a custom Authorization header,
// are set correctly on the HTTP request based on headers provided by a mock APIHandler.
// TODO need to implement MockAPIHandler
// func TestSetRequestHeaders(t *testing.T) {
// 	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
// 	mockAPIHandler := new(MockAPIHandler) // Assume you've implemented MockAPIHandler

// 	// Mock APIHandler to return a set of standard headers
// 	standardHeaders := map[string]string{
// 		"Content-Type":  "application/json",
// 		"Custom-Header": "custom-value",
// 	}
// 	mockAPIHandler.On("GetAPIRequestHeaders", "test-endpoint").Return(standardHeaders)

// 	authTokenHandler := &authenticationhandler.AuthTokenHandler{Token: "test-token"}
// 	headerHandler := NewHeaderHandler(req, nil, nil, authTokenHandler)
// 	headerHandler.SetRequestHeaders("test-endpoint")

// 	assert.Equal(t, "Bearer test-token", req.Header.Get("Authorization"), "Authorization header should be correctly set with Bearer token")
// 	assert.Equal(t, "application/json", req.Header.Get("Content-Type"), "Content-Type header should be set to application/json")
// 	assert.Equal(t, "custom-value", req.Header.Get("Custom-Header"), "Custom-Header should be set to custom-value")

// 	mockAPIHandler.AssertExpectations(t)
// }

// TestHeadersToString verifies that the HeadersToString function correctly formats
// HTTP headers into a string, with each header on a new line.
func TestHeadersToString(t *testing.T) {
	headers := http.Header{
		"Content-Type": []string{"application/json"},
		"Accept":       []string{"application/xml"},
	}

	expected := "Content-Type: application/json\nAccept: application/xml"
	result := HeadersToString(headers)

	assert.Equal(t, expected, result, "Headers should be correctly formatted into a string")
}

// TestCheckDeprecationHeader verifies that the CheckDeprecationHeader function
// can detect the presence of a Deprecation header in the HTTP response.
// TODO need to implement MockLogger
// func TestCheckDeprecationHeader(t *testing.T) {
// 	resp := &http.Response{
// 		Header: make(http.Header),
// 	}
// 	deprecationDate := "Fri, 01 Jan 2100 00:00:00 GMT"
// 	resp.Header.Set("Deprecation", deprecationDate)

// 	// Normally, you would check for a log entry here, but we're skipping logging.
// 	// This test will simply ensure the function can run without error.
// 	CheckDeprecationHeader(resp, nil) // Passing nil as logger since we're not testing logging

// 	assert.Equal(t, deprecationDate, resp.Header.Get("Deprecation"), "Deprecation header should be detected")
// }
