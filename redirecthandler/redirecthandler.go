package redirecthandler

import (
	"fmt"
	"net/http"
	"net/url"
	"sync"

	"go.uber.org/zap"
)

// RedirectHandler contains configurations for handling HTTP redirects.
type RedirectHandler struct {
	Logger             *zap.SugaredLogger           // Logger instance for logging.
	MaxRedirects       int                          // Maximum allowed redirects to prevent infinite loops.
	VisitedURLs        map[string]int               // Tracks visited URLs to detect loops.
	VisitedURLsMutex   sync.RWMutex                 // Mutex for safe concurrent access to VisitedURLs.
	SensitiveHeaders   []string                     // Headers to be removed on cross-domain redirects.
	PermanentRedirects map[string]string            // Cache for permanent redirects
	PermRedirectsMutex sync.RWMutex                 // Mutex for safe concurrent access to PermanentRedirects
	RedirectHistories  map[*http.Request][]*url.URL // Map to track redirect history for each request
}

// NewRedirectHandler creates a new instance of RedirectHandler.
func NewRedirectHandler(logger *zap.SugaredLogger, maxRedirects int) *RedirectHandler {
	return &RedirectHandler{
		Logger:             logger,
		MaxRedirects:       maxRedirects,
		VisitedURLs:        make(map[string]int),
		SensitiveHeaders:   []string{"Authorization", "Cookie"},
		PermanentRedirects: make(map[string]string),
		RedirectHistories:  make(map[*http.Request][]*url.URL),
	}
}

// AddSensitiveHeader allows adding configurable sensitive headers.
func (r *RedirectHandler) AddSensitiveHeader(header string) {
	r.SensitiveHeaders = append(r.SensitiveHeaders, header)
}

// WithRedirectHandling applies the redirect handling policy to an http.Client.
func (r *RedirectHandler) WithRedirectHandling(client *http.Client) {
	client.CheckRedirect = r.checkRedirect
}

// checkRedirect implements the redirect handling logic.
func (r *RedirectHandler) checkRedirect(req *http.Request, via []*http.Request) error {
	defer r.clearRedirectHistory(req)

	// Enforce max redirects
	if len(via) >= r.MaxRedirects {
		r.Logger.Warn("Maximum redirects reached", zap.Int("maxRedirects", r.MaxRedirects))
		return &MaxRedirectsError{MaxRedirects: r.MaxRedirects}
	}

	// Return if disallowed method
	// TODO why?
	if req.Method == http.MethodPost || req.Method == http.MethodPatch {
		r.Logger.Warn("Redirect attempted on non-idempotent method, not following", zap.String("method", req.Method))
		return http.ErrUseLastResponse
	}

	// Check for cached permanent redirect
	// TODO why do we need to cache these?
	urlString, ok := r.checkPermanentRedirect(req.URL.String())
	if ok && (req.Method == http.MethodGet || req.Method == http.MethodHead) {
		parsedURL, err := url.Parse(urlString)
		if err != nil {
			// TODO is there ever a time where the cached one will be invalid?
			r.Logger.Error("Failed to parse URL from cache", zap.String("url", urlString), zap.Error(err))
		} else {
			req.URL = parsedURL
			r.Logger.Info("Using cached permanent redirect", zap.String("originalURL", urlString), zap.String("redirectURL", parsedURL.String()))
			return nil
		}
	}

	// Track redirect history for the current request
	r.RedirectHistories[req] = append(r.RedirectHistories[req], req.URL)

	// Check for redirect loops by analyzing the history
	if redirectLoop(r.RedirectHistories[req]) {
		r.Logger.Error("Redirect loop detected", zap.Any("redirectHistory", r.RedirectHistories[req]))
		return fmt.Errorf("redirect loop detected: %v", r.RedirectHistories[req])
	}

	lastResponse := via[len(via)-1].Response
	if lastResponse.StatusCode == http.StatusPermanentRedirect || lastResponse.StatusCode == http.StatusTemporaryRedirect {
		location, err := lastResponse.Location()
		if err != nil {
			r.Logger.Error("Failed to get location from redirect response", zap.Error(err))
			return err
		}

		newReqURL, err := r.resolveRedirectURL(req.URL, location)
		if err != nil {
			r.Logger.Error("Failed to resolve redirect URL", zap.Error(err))
			return err
		}

		if newReqURL.Host != req.URL.Host {
			r.secureRequest(req)
		}

		if lastResponse.StatusCode == http.StatusPermanentRedirect {
			r.cachePermanentRedirect(req.URL.String(), newReqURL.String())
		}

		if lastResponse.StatusCode == http.StatusSeeOther {
			r.adjustForSeeOther(req)
		}

		r.Logger.Info("Redirecting request", zap.String("originalURL", req.URL.String()), zap.String("newURL", newReqURL.String()), zap.Int("redirectCount", len(via)))
		req.URL = newReqURL
		return nil
	}

	// Clear redirect history if redirect is successful
	if len(via) > 0 && lastResponse.StatusCode >= 200 && lastResponse.StatusCode < 400 {
		redirectedReq := via[len(via)-1]
		r.clearRedirectHistory(redirectedReq)
	}

	return http.ErrUseLastResponse
}

// resolveRedirectURL resolves the redirect location URL against the current request URL.
func (r *RedirectHandler) resolveRedirectURL(reqURL *url.URL, redirectURL *url.URL) (*url.URL, error) {
	if !redirectURL.IsAbs() {
		redirectURL.Scheme = reqURL.Scheme
	}
	return redirectURL, nil
}

// secureRequest removes sensitive headers from the request if the new destination is a different domain.
func (r *RedirectHandler) secureRequest(req *http.Request) {
	for _, header := range r.SensitiveHeaders {
		req.Header.Del(header)
	}
}

// adjustForSeeOther adjusts the request for "303 See Other" responses.
func (r *RedirectHandler) adjustForSeeOther(req *http.Request) {
	req.Method = http.MethodGet
	req.Body = nil
	req.GetBody = nil
	req.ContentLength = 0
	req.Header.Del("Content-Type")
}

// RedirectLoopError represents an error when a redirect loop is detected.
type RedirectLoopError struct {
	URL string
}

// RedirectLoopError defines an error for when a redirect loop is detected.
func (e *RedirectLoopError) Error() string {
	return fmt.Sprintf("redirect loop detected at %s", e.URL)
}

// MaxRedirectsError represents an error when the maximum number of redirects is reached.
type MaxRedirectsError struct {
	MaxRedirects int
}

// MaxRedirectsError defines an error for when the maximum number of redirects is reached.
func (e *MaxRedirectsError) Error() string {
	return fmt.Sprintf("maximum redirects reached: %d", e.MaxRedirects)
}

// cachePermanentRedirect caches the permanent redirect location.
func (r *RedirectHandler) cachePermanentRedirect(originalURL, redirectURL string) {
	r.PermRedirectsMutex.Lock()
	defer r.PermRedirectsMutex.Unlock()

	r.PermanentRedirects[originalURL] = redirectURL
}

// checkPermanentRedirect checks if there's a cached redirect for the given URL.
func (r *RedirectHandler) checkPermanentRedirect(originalURL string) (string, bool) {
	r.PermRedirectsMutex.RLock()
	defer r.PermRedirectsMutex.RUnlock()

	url, exists := r.PermanentRedirects[originalURL]
	return url, exists
}

// redirectLoop checks if there's a loop in the redirect history.
func redirectLoop(history []*url.URL) bool {
	var urls []string
	for _, v := range history {
		urls = append(urls, v.String())
	}

	// if duplicates found at different indexes in loop. I don't think it's pretty but it works.
	for i, j := range urls {
		for k, l := range urls {
			if i != k {
				if j == l {
					return true
				}
			}
		}
	}

	return false
}

// clearRedirectHistory clears the redirect history for a given request to prevent memory leaks.
func (r *RedirectHandler) clearRedirectHistory(req *http.Request) {
	r.VisitedURLsMutex.Lock()
	delete(r.RedirectHistories, req)
	r.VisitedURLsMutex.Unlock()
}

// GetRedirectHistory returns the redirect history for a given request.
func (r *RedirectHandler) GetRedirectHistory(req *http.Request) []*url.URL {
	r.VisitedURLsMutex.RLock()
	defer r.VisitedURLsMutex.RUnlock()

	return r.RedirectHistories[req]
}

// SetupRedirectHandler configures the HTTP client for redirect handling based on the client configuration.
func SetupRedirectHandler(client *http.Client, maxRedirects int, log *zap.SugaredLogger) {
	redirectHandler := NewRedirectHandler(log, maxRedirects)
	redirectHandler.WithRedirectHandling(client)
	log.Info("Redirect handling enabled", zap.Int("MaxRedirects", maxRedirects))

}
