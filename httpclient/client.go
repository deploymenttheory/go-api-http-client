// httpclient/client.go
/* The `http_client` package provides a configurable HTTP client tailored for interacting with specific APIs.
It supports different authentication methods, including "bearer" and "oauth". The client is designed with a
focus on concurrency management, structured error handling, and flexible configuration options.
The package offers a default timeout, custom backoff strategies, dynamic rate limiting,
and detailed logging capabilities. The main `Client` structure encapsulates all necessary components,
like the baseURL, authentication details, and an embedded standard HTTP client. */
package httpclient

import (
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"

	"github.com/deploymenttheory/go-api-http-client/concurrency"
	"go.uber.org/zap"
)

// HTTPExecutor is an interface which wraps http.Client to allow mocking.
type HTTPExecutor interface {

	// Inherited
	CloseIdleConnections()
	Do(req *http.Request) (*http.Response, error)
	Get(url string) (resp *http.Response, err error)
	Head(url string) (resp *http.Response, err error)
	Post(url string, contentType string, body io.Reader) (resp *http.Response, err error)
	PostForm(url string, data url.Values) (resp *http.Response, err error)

	// Additional
	SetCookieJar(jar http.CookieJar)
	SetCookies(url *url.URL, cookies []*http.Cookie)
	SetCustomTimeout(time.Duration)
	Cookies(*url.URL) []*http.Cookie
	SetRedirectPolicy(*func(req *http.Request, via []*http.Request) error)
}

// Master struct/object
type Client struct {
	config      *ClientConfig
	Integration *APIIntegration
	http        HTTPExecutor
	Sugar       *zap.SugaredLogger
	Concurrency *concurrency.ConcurrencyHandler
}

// Options/Variables for Client
type ClientConfig struct {
	// Interface which implements the APIIntegration patterns. Integration handles all server/endpoint specific configuration, auth and vars.
	Integration APIIntegration

	// Sugar is the logger from Zap.
	Sugar *zap.SugaredLogger

	// Wether or not empty values will be set or an error thrown for missing items.
	PopulateDefaultValues bool

	// HideSenitiveData controls if sensitive data will be visible in logs. Debug option which should be True in production use.
	HideSensitiveData bool `json:"hide_sensitive_data"`

	// CustomCookies allows implementation of persistent, session wide cookies.
	CustomCookies []*http.Cookie

	// MaxRetry Attempts limits the amount of retries the client will perform on requests which are deemd retriable.
	MaxRetryAttempts int `json:"max_retry_attempts"`

	// MaxConcurrentRequests limits the amount of Semaphore tokens available to the client and therefor limits concurrent requests.
	MaxConcurrentRequests int `json:"max_concurrent_requests"`

	// EnableDynamicRateLimiting // TODO because I don't know.
	EnableDynamicRateLimiting bool `json:"enable_dynamic_rate_limiting"`

	// CustomTimeout // TODO also because I don't know.
	CustomTimeout time.Duration

	// TokenRefreshBufferPeriod is the duration of time before the token expires in which it's deemed
	// more sensible to replace the token rather then carry on using it.
	TokenRefreshBufferPeriod time.Duration

	// TotalRetryDuration // TODO maybe this should be called context?
	TotalRetryDuration time.Duration

	// EnableCustomRedirectLogic allows the client to follow redirections when they're returned from a request.
	// Toggleable for debug reasons only
	CustomRedirectPolicy *func(req *http.Request, via []*http.Request) error

	// MaxRedirects is the maximum amount of redirects the client will follow before throwing an error.
	MaxRedirects int `json:"max_redirects"`

	// EnableConcurrencyManagement when false bypasses any concurrency management to allow for a simpler request flow.
	EnableConcurrencyManagement bool `json:"enable_concurrency_management"`

	// MandatoryRequestDelay is a short, usually sub 0.5 second, delay after every request as to not overwhelm an endpoint.
	// Can be set to nothing if you want to be lightning fast!
	MandatoryRequestDelay time.Duration

	// RetryEligiableRequests when false bypasses any retry logic for a simpler request flow.
	RetryEligiableRequests bool `json:"retry_eligiable_requests"`

	HTTPExecutor HTTPExecutor
}

// BuildClient creates a new HTTP client with the provided configuration.
func (c *ClientConfig) Build() (*Client, error) {
	if c.Sugar == nil {
		zapLogger, err := zap.NewProduction()
		if err != nil {
			return nil, err
		}

		c.Sugar = zapLogger.Sugar()
		c.Sugar.Info("No logger provided. Defaulting to Sugared Zap Production Logger")
	}

	c.Sugar.Debug("validating configuration")

	err := c.validateClientConfig()
	if err != nil {
		return nil, fmt.Errorf("invalid configuration: %v", err)
	}

	c.Sugar.Debug("configuration valid")

	httpClient := c.HTTPExecutor

	cookieJar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	httpClient.SetCookieJar(cookieJar)

	if c.CustomRedirectPolicy != nil {
		httpClient.SetRedirectPolicy(c.CustomRedirectPolicy)
	}

	// TODO refactor concurrency
	var concurrencyHandler *concurrency.ConcurrencyHandler
	if c.EnableConcurrencyManagement {
		concurrencyMetrics := &concurrency.ConcurrencyMetrics{}
		concurrencyHandler = concurrency.NewConcurrencyHandler(
			c.MaxConcurrentRequests,
			c.Sugar,
			concurrencyMetrics,
		)
	}

	client := &Client{
		Integration: &c.Integration,
		http:        httpClient,
		config:      c,
		Sugar:       c.Sugar,
		Concurrency: concurrencyHandler,
	}

	if len(client.config.CustomCookies) > 0 {
		client.Sugar.Debug("setting custom cookies")
		client.loadCustomCookies()
	}

	client.Sugar.Infof("client init complete: %+v", client)

	return client, nil

}
