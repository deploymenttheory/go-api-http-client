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
	"net/http"
	"time"

	"github.com/deploymenttheory/go-api-http-client/concurrency"
	"go.uber.org/zap"

	"github.com/deploymenttheory/go-api-http-client/redirecthandler"
)

const ()

// TODO all struct comments

// Master struct/object
type Client struct {
	// Config
	config *ClientConfig

	// Integration
	Integration *APIIntegration

	// Executor
	http *http.Client

	// Logger
	Sugar *zap.SugaredLogger

	// Concurrency Mananger
	Concurrency *concurrency.ConcurrencyHandler
}

// Options/Variables for Client
type ClientConfig struct {
	// Interface which implements the APIIntegration patterns. Integration handles all server/endpoint specific configuration, auth and vars.
	Integration APIIntegration

	// TODO
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

	// FollowRedirects allows the client to follow redirections when they're returned from a request.
	FollowRedirects bool `json:"follow_redirects"`

	// MaxRedirects is the maximum amount of redirects the client will follow before throwing an error.
	MaxRedirects int `json:"max_redirects"`

	// EnableConcurrencyManagement when false bypasses any concurrency management to allow for a simpler request flow.
	EnableConcurrencyManagement bool `json:"enable_concurrency_management"`

	// MandatoryRequestDelay is a short, usually sub 0.5 second, delay after every request as to not overwhelm an endpoint.
	// Can be set to nothing if you want to be lightning fast!
	MandatoryRequestDelay time.Duration

	// RetryEligiableRequests when false bypasses any retry logic for a simpler request flow.
	RetryEligiableRequests bool `json:"retry_eligiable_requests"`
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

	httpClient := &http.Client{
		Timeout: c.CustomTimeout,
	}

	// TODO refactor redirects
	if err := redirecthandler.SetupRedirectHandler(httpClient, c.FollowRedirects, c.MaxRedirects, c.Sugar); err != nil {
		return nil, fmt.Errorf("Failed to set up redirect handler: %v", err)
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
	} else {
		concurrencyHandler = nil
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

	return client, nil

}
