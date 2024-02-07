// http_client.go
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
	"sync"
	"time"

	"go.uber.org/zap"
)

// Config holds configuration options for the HTTP Client.
type Config struct {
	// Required
	InstanceName string
	Auth         AuthConfig // User can either supply these values manually or pass from LoadAuthConfig/Env vars
	APIType      string     `json:"apiType"`
	// Optional
	LogLevel                  LogLevel // Field for defining tiered logging level.
	MaxRetryAttempts          int      // Config item defines the max number of retry request attempts for retryable HTTP methods.
	EnableDynamicRateLimiting bool
	Logger                    Logger // Field for the packages initailzed logger
	MaxConcurrentRequests     int    // Field for defining the maximum number of concurrent requests allowed in the semaphore
	TokenRefreshBufferPeriod  time.Duration
	TotalRetryDuration        time.Duration
	CustomTimeout             time.Duration
}

// ClientPerformanceMetrics captures various metrics related to the client's
// interactions with the API, providing insights into its performance and behavior.
type PerformanceMetrics struct {
	TotalRequests        int64
	TotalRetries         int64
	TotalRateLimitErrors int64
	TotalResponseTime    time.Duration
	TokenWaitTime        time.Duration
	lock                 sync.Mutex
}

// ClientAuthConfig represents the structure to read authentication details from a JSON configuration file.
type AuthConfig struct {
	InstanceName       string `json:"instanceName,omitempty"`
	OverrideBaseDomain string `json:"overrideBaseDomain,omitempty"`
	Username           string `json:"username,omitempty"`
	Password           string `json:"password,omitempty"`
	ClientID           string `json:"clientID,omitempty"`
	ClientSecret       string `json:"clientSecret,omitempty"`
}

// Client represents an HTTP client to interact with a specific API.
type Client struct {
	APIHandler                 APIHandler                 // APIHandler interface used to define which API handler to use
	InstanceName               string                     // Website Instance name without the root domain
	AuthMethod                 string                     // Specifies the authentication method: "bearer" or "oauth"
	Token                      string                     // Authentication Token
	OverrideBaseDomain         string                     // Base domain override used when the default in the api handler isn't suitable
	OAuthCredentials           OAuthCredentials           // ClientID / Client Secret
	BearerTokenAuthCredentials BearerTokenAuthCredentials // Username and Password for Basic Authentication
	Expiry                     time.Time                  // Expiry time set for the auth token
	httpClient                 *http.Client
	tokenLock                  sync.Mutex
	config                     Config
	logger                     Logger
	ConcurrencyMgr             *ConcurrencyManager
	PerfMetrics                PerformanceMetrics
}

// BuildClient creates a new HTTP client with the provided configuration.
func BuildClient(config Config) (*Client, error) {
	// Use the Logger interface type for the logger variable
	var logger Logger
	if config.Logger == nil {
		logger = NewDefaultLogger()
	} else {
		logger = config.Logger
	}

	// Set the logger's level based on the provided configuration if present
	logger.SetLevel(config.LogLevel)

	// Validate LogLevel
	if config.LogLevel < LogLevelNone || config.LogLevel > LogLevelDebug {
		return nil, fmt.Errorf("invalid LogLevel setting: %d", config.LogLevel)
	}

	// Use the APIType from the config to determine which API handler to load
	apiHandler, err := LoadAPIHandler(config, config.APIType)
	if err != nil {
		logger.Error("Failed to load API handler", zap.String("APIType", config.APIType), zap.Error(err))
		return nil, err // Return the original error without wrapping it in fmt.Errorf
	}

	logger.Info("Initializing new HTTP client", zap.String("InstanceName", config.InstanceName), zap.String("APIType", config.APIType), zap.Int("LogLevel", int(config.LogLevel)))
	// Validate and set default values for the configuration
	if config.InstanceName == "" {
		return nil, fmt.Errorf("instanceName cannot be empty")
	}

	if config.MaxRetryAttempts < 0 {
		config.MaxRetryAttempts = DefaultMaxRetryAttempts
		logger.Info("MaxRetryAttempts was negative, set to default value", zap.Int("MaxRetryAttempts", DefaultMaxRetryAttempts))
	}

	if config.MaxConcurrentRequests <= 0 {
		config.MaxConcurrentRequests = DefaultMaxConcurrentRequests
		logger.Info("MaxConcurrentRequests was negative or zero, set to default value", zap.Int("MaxConcurrentRequests", DefaultMaxConcurrentRequests))
	}

	if config.TokenRefreshBufferPeriod < 0 {
		config.TokenRefreshBufferPeriod = DefaultTokenBufferPeriod
		logger.Info("TokenRefreshBufferPeriod was negative, set to default value", zap.Duration("TokenRefreshBufferPeriod", DefaultTokenBufferPeriod))
	}

	if config.TotalRetryDuration <= 0 {
		config.TotalRetryDuration = DefaultTotalRetryDuration
		logger.Info("TotalRetryDuration was negative or zero, set to default value", zap.Duration("TotalRetryDuration", DefaultTotalRetryDuration))
	}

	if config.TokenRefreshBufferPeriod == 0 {
		config.TokenRefreshBufferPeriod = DefaultTokenBufferPeriod
		logger.Info("TokenRefreshBufferPeriod not set, set to default value", zap.Duration("TokenRefreshBufferPeriod", DefaultTokenBufferPeriod))
	}

	if config.TotalRetryDuration == 0 {
		config.TotalRetryDuration = DefaultTotalRetryDuration
		logger.Info("TotalRetryDuration not set, set to default value", zap.Duration("TotalRetryDuration", DefaultTotalRetryDuration))
	}

	if config.CustomTimeout == 0 {
		config.CustomTimeout = DefaultTimeout
		logger.Info("CustomTimeout not set, set to default value", zap.Duration("CustomTimeout", DefaultTimeout))
	}

	// Determine the authentication method
	AuthMethod := "unknown"
	if config.Auth.Username != "" && config.Auth.Password != "" {
		AuthMethod = "bearer"
	} else if config.Auth.ClientID != "" && config.Auth.ClientSecret != "" {
		AuthMethod = "oauth"
	} else {
		return nil, fmt.Errorf("invalid AuthConfig")
	}

	client := &Client{
		InstanceName:   config.InstanceName,
		APIHandler:     apiHandler,
		AuthMethod:     AuthMethod,
		httpClient:     &http.Client{Timeout: config.CustomTimeout},
		config:         config,
		logger:         logger,
		ConcurrencyMgr: NewConcurrencyManager(config.MaxConcurrentRequests, logger, true),
		PerfMetrics:    PerformanceMetrics{},
	}

	// Get auth token
	_, err = client.ValidAuthTokenCheck()
	if err != nil {
		logger.Error("Failed to validate or obtain auth token", zap.Error(err))
		return nil, fmt.Errorf("failed to validate auth: %w", err)
	}

	go client.StartMetricEvaluation()

	logger.Info("New client initialized", zap.String("InstanceName", client.InstanceName), zap.String("AuthMethod", AuthMethod), zap.Int("MaxRetryAttempts", config.MaxRetryAttempts), zap.Int("MaxConcurrentRequests", config.MaxConcurrentRequests), zap.Bool("EnableDynamicRateLimiting", config.EnableDynamicRateLimiting))

	return client, nil

}
