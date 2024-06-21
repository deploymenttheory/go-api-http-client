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
	Integration                  APIIntegration
	HideSensitiveData            bool
	CustomCookies                []*http.Cookie
	MaxRetryAttempts             int
	MaxConcurrentRequests        int
	EnableDynamicRateLimiting    bool
	CustomTimeout                time.Duration
	TokenRefreshBufferPeriod     time.Duration
	TotalRetryDuration           time.Duration // TODO do we need this now it's in the integration?
	MandatoryRequestDelay        time.Duration
	FollowRedirects              bool
	MaxRedirects                 int
	ConcurrencyManagementEnabled bool
	RetryEligiableRequests       bool
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
		log.Error("Failed to set up redirect handler", zap.Error(err))
		return nil, err
	}

	var concurrencyHandler *concurrency.ConcurrencyHandler
	if config.ConcurrencyManagementEnabled {
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
	)

	return client, nil

}
