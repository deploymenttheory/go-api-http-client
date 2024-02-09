// http_client.go
/* The `http_client` package provides a configurable HTTP client tailored for interacting with specific APIs.
It supports different authentication methods, including "bearer" and "oauth". The client is designed with a
focus on concurrency management, structured error handling, and flexible configuration options.
The package offers a default timeout, custom backoff strategies, dynamic rate limiting,
and detailed logging capabilities. The main `Client` structure encapsulates all necessary components,
like the baseURL, authentication details, and an embedded standard HTTP client. */
package httpclient

import (
	"net/http"
	"sync"
	"time"

	"github.com/deploymenttheory/go-api-http-client/logger"
	"go.uber.org/zap"
)

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
	clientConfig               ClientConfig
	Logger                     logger.Logger
	ConcurrencyMgr             *ConcurrencyManager
	PerfMetrics                PerformanceMetrics
}

// Config holds configuration options for the HTTP Client.
type ClientConfig struct {
	Auth          AuthConfig        // User can either supply these values manually or pass from LoadAuthConfig/Env vars
	Environment   EnvironmentConfig // User can either supply these values manually or pass from LoadAuthConfig/Env vars
	ClientOptions ClientOptions     // Optional configuration options for the HTTP Client
}

// EnvironmentConfig represents the structure to read authentication details from a JSON configuration file.
type EnvironmentConfig struct {
	InstanceName       string `json:"InstanceName,omitempty"`
	OverrideBaseDomain string `json:"OverrideBaseDomain,omitempty"`
	APIType            string `json:"APIType,omitempty"`
}

// AuthConfig represents the structure to read authentication details from a JSON configuration file.
type AuthConfig struct {
	Username     string `json:"Username,omitempty"`
	Password     string `json:"Password,omitempty"`
	ClientID     string `json:"ClientID,omitempty"`
	ClientSecret string `json:"ClientSecret,omitempty"`
}

// ClientOptions holds optional configuration options for the HTTP Client.
type ClientOptions struct {
	LogLevel                  logger.LogLevel // Field for defining tiered logging level.
	HideSensitiveData         bool            // Field for defining whether sensitive fields should be hidden in logs.
	MaxRetryAttempts          int             // Config item defines the max number of retry request attempts for retryable HTTP methods.
	EnableDynamicRateLimiting bool            // Field for defining whether dynamic rate limiting should be enabled.
	MaxConcurrentRequests     int             // Field for defining the maximum number of concurrent requests allowed in the semaphore
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

// BuildClient creates a new HTTP client with the provided configuration.
func BuildClient(config ClientConfig) (*Client, error) {
	// Initialize the zap logger.
	log := logger.BuildLogger(config.ClientOptions.LogLevel)

	// Set the logger's level based on the provided configuration.
	log.SetLevel(config.ClientOptions.LogLevel)

	if config.ClientOptions.LogLevel < logger.LogLevelDebug || config.ClientOptions.LogLevel > logger.LogLevelFatal {
		return nil, log.Error("Invalid LogLevel setting", zap.Int("Provided LogLevel", int(config.ClientOptions.LogLevel)))
	}

	// Use the APIType from the config to determine which API handler to load
	apiHandler, err := LoadAPIHandler(config.Environment.APIType, log)
	if err != nil {
		return nil, log.Error("Failed to load API handler", zap.String("APIType", config.Environment.APIType), zap.Error(err))
	}

	log.Info("Initializing new HTTP client with the provided configuration")

	// Validate and set default values for the configuration
	if config.Environment.APIType == "" {
		return nil, log.Error("InstanceName cannot be empty")
	}

	if config.ClientOptions.MaxRetryAttempts < 0 {
		config.ClientOptions.MaxRetryAttempts = DefaultMaxRetryAttempts
		log.Info("MaxRetryAttempts was negative, set to default value", zap.Int("MaxRetryAttempts", DefaultMaxRetryAttempts))
	}

	if config.ClientOptions.MaxConcurrentRequests <= 0 {
		config.ClientOptions.MaxConcurrentRequests = DefaultMaxConcurrentRequests
		log.Info("MaxConcurrentRequests was negative or zero, set to default value", zap.Int("MaxConcurrentRequests", DefaultMaxConcurrentRequests))
	}

	if config.ClientOptions.TokenRefreshBufferPeriod < 0 {
		config.ClientOptions.TokenRefreshBufferPeriod = DefaultTokenBufferPeriod
		log.Info("TokenRefreshBufferPeriod was negative, set to default value", zap.Duration("TokenRefreshBufferPeriod", DefaultTokenBufferPeriod))
	}

	if config.ClientOptions.TotalRetryDuration <= 0 {
		config.ClientOptions.TotalRetryDuration = DefaultTotalRetryDuration
		log.Info("TotalRetryDuration was negative or zero, set to default value", zap.Duration("TotalRetryDuration", DefaultTotalRetryDuration))
	}

	if config.ClientOptions.TokenRefreshBufferPeriod == 0 {
		config.ClientOptions.TokenRefreshBufferPeriod = DefaultTokenBufferPeriod
		log.Info("TokenRefreshBufferPeriod not set, set to default value", zap.Duration("TokenRefreshBufferPeriod", DefaultTokenBufferPeriod))
	}

	if config.ClientOptions.TotalRetryDuration == 0 {
		config.ClientOptions.TotalRetryDuration = DefaultTotalRetryDuration
		log.Info("TotalRetryDuration not set, set to default value", zap.Duration("TotalRetryDuration", DefaultTotalRetryDuration))
	}

	if config.ClientOptions.CustomTimeout == 0 {
		config.ClientOptions.CustomTimeout = DefaultTimeout
		log.Info("CustomTimeout not set, set to default value", zap.Duration("CustomTimeout", DefaultTimeout))
	}

	// Determine the authentication method using the helper function
	authMethod, err := DetermineAuthMethod(config.Auth)
	if err != nil {
		log.Error("Failed to determine authentication method", zap.Error(err))
		return nil, err
	}

	// Create a new HTTP client with the provided configuration.
	client := &Client{
		APIHandler:         apiHandler,
		InstanceName:       config.Environment.InstanceName,
		AuthMethod:         authMethod,
		OverrideBaseDomain: config.Environment.OverrideBaseDomain,
		httpClient:         &http.Client{Timeout: config.ClientOptions.CustomTimeout},
		clientConfig:       config,
		Logger:             log,
		ConcurrencyMgr:     NewConcurrencyManager(config.ClientOptions.MaxConcurrentRequests, log, true),
		PerfMetrics:        PerformanceMetrics{},
	}

	// Log the client's configuration.
	log.Info("New API client initialized",
		zap.String("API Type", config.Environment.APIType),
		zap.String("Instance Name", client.InstanceName),
		zap.String("Override Base Domain", config.Environment.OverrideBaseDomain),
		zap.String("Authentication Method", authMethod),
		zap.String("Logging Level", config.ClientOptions.LogLevel.String()),
		zap.Bool("Hide Sensitive Data In Logs", config.ClientOptions.HideSensitiveData),
		zap.Int("Max Retry Attempts", config.ClientOptions.MaxRetryAttempts),
		zap.Int("Max Concurrent Requests", config.ClientOptions.MaxConcurrentRequests),
		zap.Bool("Enable Dynamic Rate Limiting", config.ClientOptions.EnableDynamicRateLimiting),
		zap.Duration("Token Refresh Buffer Period", config.ClientOptions.TokenRefreshBufferPeriod),
		zap.Duration("Total Retry Duration", config.ClientOptions.TotalRetryDuration),
		zap.Duration("Custom Timeout", config.ClientOptions.CustomTimeout),
	)

	return client, nil

}
