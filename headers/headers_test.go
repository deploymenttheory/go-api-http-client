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
