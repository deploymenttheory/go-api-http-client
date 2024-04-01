package redirecthandler

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/deploymenttheory/go-api-http-client/logger"
	"github.com/deploymenttheory/go-api-http-client/mocklogger"
	"github.com/stretchr/testify/assert"
)

// TestRedirectHandler_CheckRedirect tests the checkRedirect method of the RedirectHandler.
// It covers various scenarios including redirect loop detection, maximum redirects limit,
// resolving relative redirects, cross-domain security measures, and handling of 303 See Other response.
func TestRedirectHandler_CheckRedirect(t *testing.T) {
	mockLogger := mocklogger.NewMockLogger()

	// Set the mock logger to capture logs at all levels
	mockLogger.SetLevel(logger.LogLevelDebug)

	redirectHandler := NewRedirectHandler(mockLogger, 10)

	reqURL, _ := url.Parse("http://example.com")
	req := &http.Request{URL: reqURL, Method: http.MethodPost}
	resp := &http.Response{
		Status:     "303 See Other",
		StatusCode: http.StatusSeeOther,
		Header:     http.Header{"Location": []string{"http://example.com/new"}},
	}

	t.Run("Redirect Loop Detection", func(t *testing.T) {
		redirectHandler.VisitedURLs = map[string]int{"http://example.com": 1}
		err := redirectHandler.checkRedirect(req, []*http.Request{{}, {}})
		assert.Equal(t, http.ErrUseLastResponse, err)
		// Verify that a warning log for redirect loop was recorded
		assert.Contains(t, mockLogger.Calls[0].Arguments.String(0), "Detected redirect loop")
	})

	t.Run("Maximum Redirects Reached", func(t *testing.T) {
		redirectHandler.VisitedURLs = map[string]int{}
		redirectHandler.MaxRedirects = 1
		err := redirectHandler.checkRedirect(req, []*http.Request{{}, {}})
		assert.Equal(t, http.ErrUseLastResponse, err)
		// Verify that a warning log for max redirects was recorded
		assert.Contains(t, mockLogger.Calls[1].Arguments.String(0), "Stopped after maximum redirects")
	})

	t.Run("Resolve Relative Redirects", func(t *testing.T) {
		redirectHandler.MaxRedirects = 10
		err := redirectHandler.checkRedirect(req, []*http.Request{{}, {}})
		assert.Nil(t, err)
		assert.Equal(t, "http://example.com/new", req.URL.String())
	})

	t.Run("Cross-Domain Security Measures", func(t *testing.T) {
		reqURL, _ = url.Parse("http://example.com")
		req = &http.Request{URL: reqURL, Method: http.MethodPost}
		resp.Header.Set("Location", "http://anotherdomain.com/new")
		err := redirectHandler.checkRedirect(req, []*http.Request{{}, {}})
		assert.Nil(t, err)
		// Ensure sensitive headers are removed and corresponding log is recorded
		assert.Empty(t, req.Header.Get("Authorization"))
		assert.Contains(t, mockLogger.Calls[2].Arguments.String(0), "Removed sensitive header")
	})

	t.Run("Handling 303 See Other", func(t *testing.T) {
		reqURL, _ = url.Parse("http://example.com")
		req = &http.Request{URL: reqURL, Method: http.MethodPost}
		resp.Header.Set("Location", "http://example.com/new")
		err := redirectHandler.checkRedirect(req, []*http.Request{{}, {}})
		assert.Nil(t, err)
		assert.Equal(t, http.MethodGet, req.Method)
		// Ensure no body, no GetBody, correct ContentLength, no Content-Type header, and a log is recorded
		assert.Nil(t, req.Body)
		assert.Nil(t, req.GetBody)
		assert.Equal(t, int64(0), req.ContentLength)
		assert.Empty(t, req.Header.Get("Content-Type"))
		assert.Contains(t, mockLogger.Calls[3].Arguments.String(0), "Changed request method to GET")
	})
}

// TestRedirectHandler_ResolveRedirectURL tests the resolveRedirectURL method of the RedirectHandler.
// It checks the correct resolution of absolute and relative URLs including those with query parameters and fragments.
func TestRedirectHandler_ResolveRedirectURL(t *testing.T) {
	redirectHandler := RedirectHandler{}

	t.Run("Absolute URL", func(t *testing.T) {
		reqURL, _ := url.Parse("http://example.com")
		redirectURL, _ := url.Parse("http://newexample.com/path")
		newReqURL, err := redirectHandler.resolveRedirectURL(reqURL, redirectURL)
		assert.Nil(t, err)
		assert.Equal(t, redirectURL.String(), newReqURL.String())
	})

	t.Run("Relative URL", func(t *testing.T) {
		reqURL, _ := url.Parse("http://example.com/current")
		redirectURL, _ := url.Parse("/newpath")
		newReqURL, err := redirectHandler.resolveRedirectURL(reqURL, redirectURL)
		assert.Nil(t, err)
		assert.Equal(t, "http://example.com/newpath", newReqURL.String())
	})

	t.Run("Relative URL with Query and Fragment", func(t *testing.T) {
		reqURL, _ := url.Parse("http://example.com/current?param=value#fragment")
		redirectURL, _ := url.Parse("newpath?newparam=newvalue#newfragment")
		newReqURL, err := redirectHandler.resolveRedirectURL(reqURL, redirectURL)
		assert.Nil(t, err)
		assert.Equal(t, "http://example.com/newpath?newparam=newvalue#newfragment", newReqURL.String())
	})
}

// TestRedirectHandler_SecureRequest tests the secureRequest method of the RedirectHandler.
// It verifies that sensitive headers are correctly removed when a request is redirected to a different domain.
func TestRedirectHandler_SecureRequest(t *testing.T) {
	mockLogger := mocklogger.NewMockLogger()
	mockLogger.SetLevel(logger.LogLevelDebug)

	redirectHandler := RedirectHandler{Logger: mockLogger}
	req := &http.Request{Header: http.Header{"Authorization": []string{"token"}, "Cookie": []string{"session"}}}

	t.Run("Secure Cross-Domain Redirect", func(t *testing.T) {
		redirectHandler.secureRequest(req)
		// Ensure sensitive headers are removed and log messages were recorded
		assert.Empty(t, req.Header.Get("Authorization"))
		assert.Empty(t, req.Header.Get("Cookie"))
		assert.Contains(t, mockLogger.Calls[0].Arguments.String(0), "Removed sensitive header")
	})
}

// Test for Redirect Loop Detection - This test ensures that the redirect handler correctly identifies and stops redirect loops.
func TestRedirectLoopDetection(t *testing.T) {
	// Setup
	mockLogger := mocklogger.NewMockLogger()
	handler := NewRedirectHandler(mockLogger, 5)
	loopURL, _ := url.Parse("http://example.com/loop")
	req := &http.Request{URL: loopURL}

	// Simulate a redirect loop by adding the same URL to the history multiple times
	handler.RedirectHistories[req] = []*url.URL{loopURL, loopURL}

	// Test
	err := handler.checkRedirect(req, []*http.Request{req, req})

	// Assertions
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "redirect loop detected")
	// Verify log message for loop detection
	assert.Contains(t, mockLogger.Calls[0].Arguments.String(0), "Redirect loop detected")
}

// TestRedirectHistoryCleanup - This test ensures that the redirect history for each request is properly cleaned up to prevent memory leaks.
func TestRedirectHistoryCleanup(t *testing.T) {
	// Setup
	mockLogger := mocklogger.NewMockLogger()
	handler := NewRedirectHandler(mockLogger, 5)
	req := &http.Request{URL: &url.URL{Path: "/test"}}

	// Simulate adding some history
	handler.RedirectHistories[req] = []*url.URL{{Path: "/redirect1"}, {Path: "/redirect2"}}

	// Perform a redirect that will trigger the cleanup
	handler.checkRedirect(req, []*http.Request{req})

	// Assertions
	_, exists := handler.RedirectHistories[req]
	assert.False(t, exists)
}

// TestMaxRedirectsReached - This test checks that the handler stops redirects after reaching the maximum limit.
func TestMaxRedirectsReached(t *testing.T) {
	// Setup
	mockLogger := mocklogger.NewMockLogger()
	handler := NewRedirectHandler(mockLogger, 1) // Set max redirects to 1
	req := &http.Request{URL: &url.URL{Path: "/start"}}
	via := []*http.Request{{}, {}} // Simulate one redirect has already occurred

	// Test
	err := handler.checkRedirect(req, via)

	// Assertions
	assert.NotNil(t, err)
	assert.IsType(t, &MaxRedirectsError{}, err)
	maxRedirectsError := err.(*MaxRedirectsError)
	assert.Equal(t, 1, maxRedirectsError.MaxRedirects)
	// Verify log message for max redirects reached
	assert.Contains(t, mockLogger.Calls[0].Arguments.String(0), "Maximum redirects reached")
}