package redirecthandler

import (
	"net/http"
	"net/url"
	"sync"

	"github.com/deploymenttheory/go-api-http-client/logger"
	"github.com/deploymenttheory/go-api-http-client/status"
	"go.uber.org/zap"
)

// RedirectHandler handles HTTP redirects within an http.Client.
// It provides features such as redirect loop detection, security enhancements,
// and integration with client settings for fine-grained control over redirect behavior.
type RedirectHandler struct {
	Logger           logger.Logger
	MaxRedirects     int
	VisitedURLs      map[string]int
	VisitedURLsMutex sync.Mutex
	SensitiveHeaders []string
}

// NewRedirectHandler creates a new instance of RedirectHandler with the provided logger
// and maximum number of redirects. It initializes internal structures and is ready to use.
func NewRedirectHandler(logger logger.Logger, maxRedirects int) *RedirectHandler {
	return &RedirectHandler{
		Logger:           logger,
		MaxRedirects:     maxRedirects,
		VisitedURLs:      make(map[string]int),
		SensitiveHeaders: []string{"Authorization", "Cookie"}, // Add other sensitive headers if needed
	}
}

// WithRedirectHandling applies the redirect handling policy to an http.Client.
// It sets the CheckRedirect function on the client to use the handler's logic.
func (r *RedirectHandler) WithRedirectHandling(client *http.Client) {
	client.CheckRedirect = r.checkRedirect
}

// checkRedirect is the core function that implements the redirect handling logic.
// It is set as the CheckRedirect function on an http.Client and is called whenever
// the client encounters a 3XX response. It enforces the max redirects limit,
// detects redirect loops, applies security measures for cross-domain redirects,
// resolves relative redirects, and optimizes performance.
func (r *RedirectHandler) checkRedirect(req *http.Request, via []*http.Request) error {
	// Redirect Loop Detection
	r.VisitedURLsMutex.Lock()
	defer r.VisitedURLsMutex.Unlock()
	if _, exists := r.VisitedURLs[req.URL.String()]; exists {
		r.Logger.Warn("Detected redirect loop", zap.String("url", req.URL.String()))
		return http.ErrUseLastResponse
	}
	r.VisitedURLs[req.URL.String()]++

	if len(via) >= r.MaxRedirects {
		r.Logger.Warn("Stopped after maximum redirects", zap.Int("maxRedirects", r.MaxRedirects))
		return http.ErrUseLastResponse
	}

	lastResponse := via[len(via)-1].Response
	if status.IsRedirectStatusCode(lastResponse.StatusCode) {
		location, err := lastResponse.Location()
		if err != nil {
			r.Logger.Error("Failed to get location from redirect response", zap.Error(err))
			return err
		}

		// Resolve relative redirects against the current request URL
		newReqURL, err := r.resolveRedirectURL(req.URL, location)
		if err != nil {
			r.Logger.Error("Failed to resolve redirect URL", zap.Error(err))
			return err
		}

		// Security Measures
		if newReqURL.Host != req.URL.Host {
			r.secureRequest(req)
		}

		// Handling 303 See Other
		if lastResponse.StatusCode == http.StatusSeeOther {
			req.Method = http.MethodGet
			req.Body = nil
			req.GetBody = nil
			req.ContentLength = 0
			req.Header.Del("Content-Type")
			r.Logger.Info("Changed request method to GET for 303 See Other response")
		}

		req.URL = newReqURL
		r.Logger.Info("Redirecting request", zap.String("newURL", newReqURL.String()))
		return nil
	}

	return http.ErrUseLastResponse
}

// resolveRedirectURL resolves the redirect location URL against the current request URL
// to handle relative redirects accurately.
func (r *RedirectHandler) resolveRedirectURL(reqURL *url.URL, redirectURL *url.URL) (*url.URL, error) {
	if redirectURL.IsAbs() {
		return redirectURL, nil // Absolute URL, no need to resolve
	}

	// Relative URL, resolve against the current request URL
	absoluteURL := *reqURL
	absoluteURL.Path = redirectURL.Path
	absoluteURL.RawQuery = redirectURL.RawQuery
	absoluteURL.Fragment = redirectURL.Fragment
	return &absoluteURL, nil
}

// secureRequest removes sensitive headers from the request if the new destination is a different domain.
func (r *RedirectHandler) secureRequest(req *http.Request) {
	for _, header := range r.SensitiveHeaders {
		req.Header.Del(header)
		r.Logger.Info("Removed sensitive header due to domain change", zap.String("header", header))
	}
}
