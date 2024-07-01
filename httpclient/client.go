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

	"github.com/deploymenttheory/go-api-http-client/logger"
	"github.com/deploymenttheory/go-api-http-client/redirecthandler"
	"go.uber.org/zap"
)

// TODO all struct comments

// Master struct/object
type Client struct {
	config ClientConfig
	http   *http.Client

	AuthToken       string
	AuthTokenExpiry time.Time
	Logger          logger.Logger
	Concurrency     *concurrency.ConcurrencyHandler
	Integration     *APIIntegration
}

// Options/Variables for Client
type ClientConfig struct {
	// Interface which implements the APIIntegration patterns. Integration handles all server/endpoint specific configuration, auth and vars.
	Integration APIIntegration

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
func BuildClient(config ClientConfig, populateDefaultValues bool, log logger.Logger) (*Client, error) {
	err := validateClientConfig(config, populateDefaultValues)
	if err != nil {
		return nil, fmt.Errorf("invalid configuration: %v", err)
	}

	log.Info(fmt.Sprintf("initializing new http client, auth: %s", config.Integration.GetFQDN()))

	httpClient := &http.Client{
		Timeout: config.CustomTimeout,
	}

	// TODO refactor redirects
	if err := redirecthandler.SetupRedirectHandler(httpClient, config.FollowRedirects, config.MaxRedirects, log); err != nil {
		return nil, fmt.Errorf("Failed to set up redirect handler: %v", err)
	}

	var concurrencyHandler *concurrency.ConcurrencyHandler
	if config.EnableConcurrencyManagement {
		concurrencyMetrics := &concurrency.ConcurrencyMetrics{}
		concurrencyHandler = concurrency.NewConcurrencyHandler(
			config.MaxConcurrentRequests,
			log,
			concurrencyMetrics,
		)
	} else {
		concurrencyHandler = nil
	}

	client := &Client{
		Integration: &config.Integration,
		http:        httpClient,
		config:      config,
		Logger:      log,
		Concurrency: concurrencyHandler,
	}

	if len(client.config.CustomCookies) > 0 {
		client.loadCustomCookies(config.CustomCookies)
	}

	log.Debug("New API client initialized",
		zap.String("Authentication Method", (*client.Integration).GetAuthMethodDescriptor()),
		zap.Bool("Hide Sensitive Data In Logs", config.HideSensitiveData),
		zap.Int("Max Retry Attempts", config.MaxRetryAttempts),
		zap.Bool("Enable Dynamic Rate Limiting", config.EnableDynamicRateLimiting),
		zap.Int("Max Concurrent Requests", config.MaxConcurrentRequests),
		zap.Bool("Follow Redirects", config.FollowRedirects),
		zap.Int("Max Redirects", config.MaxRedirects),
		zap.Duration("Token Refresh Buffer Period", config.TokenRefreshBufferPeriod),
		zap.Duration("Total Retry Duration", config.TotalRetryDuration),
		zap.Duration("Custom Timeout", config.CustomTimeout),
		zap.Bool("Enable Concurrency Management", config.EnableConcurrencyManagement),
	)

	return client, nil

}
