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

// Config holds configuration options for the HTTP Client.
type Config struct {
	// Required
	Auth        AuthConfig        // User can either supply these values manually or pass from LoadAuthConfig/Env vars
	Environment EnvironmentConfig // User can either supply these values manually or pass from LoadAuthConfig/Env vars
	// Optional
	LogLevel                  logger.LogLevel // Field for defining tiered logging level.
	MaxRetryAttempts          int             // Config item defines the max number of retry request attempts for retryable HTTP methods.
	EnableDynamicRateLimiting bool
	MaxConcurrentRequests     int // Field for defining the maximum number of concurrent requests allowed in the semaphore
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

// EnvironmentConfig represents the structure to read authentication details from a JSON configuration file.
type EnvironmentConfig struct {
	InstanceName       string `json:"instanceName,omitempty"`
	OverrideBaseDomain string `json:"overrideBaseDomain,omitempty"`
	APIType            string `json:"apiType,omitempty"`
}

// AuthConfig represents the structure to read authentication details from a JSON configuration file.
type AuthConfig struct {
	Username     string `json:"username,omitempty"`
	Password     string `json:"password,omitempty"`
	ClientID     string `json:"clientID,omitempty"`
	ClientSecret string `json:"clientSecret,omitempty"`
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
	Logger                     logger.Logger
	ConcurrencyMgr             *ConcurrencyManager
	PerfMetrics                PerformanceMetrics
}

// BuildClient creates a new HTTP client with the provided configuration.
func BuildClient(config Config) (*Client, error) {
	// Initialize the zap logger.
	log := logger.BuildLogger(config.LogLevel)

	// Set the logger's level based on the provided configuration.
	log.SetLevel(config.LogLevel)

	if config.LogLevel < logger.LogLevelDebug || config.LogLevel > logger.LogLevelFatal {
		return nil, log.Error("Invalid LogLevel setting", zap.Int("Provided LogLevel", int(config.LogLevel)))
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

	if config.MaxRetryAttempts < 0 {
		config.MaxRetryAttempts = DefaultMaxRetryAttempts
		log.Info("MaxRetryAttempts was negative, set to default value", zap.Int("MaxRetryAttempts", DefaultMaxRetryAttempts))
	}

	if config.MaxConcurrentRequests <= 0 {
		config.MaxConcurrentRequests = DefaultMaxConcurrentRequests
		log.Info("MaxConcurrentRequests was negative or zero, set to default value", zap.Int("MaxConcurrentRequests", DefaultMaxConcurrentRequests))
	}

	if config.TokenRefreshBufferPeriod < 0 {
		config.TokenRefreshBufferPeriod = DefaultTokenBufferPeriod
		log.Info("TokenRefreshBufferPeriod was negative, set to default value", zap.Duration("TokenRefreshBufferPeriod", DefaultTokenBufferPeriod))
	}

	if config.TotalRetryDuration <= 0 {
		config.TotalRetryDuration = DefaultTotalRetryDuration
		log.Info("TotalRetryDuration was negative or zero, set to default value", zap.Duration("TotalRetryDuration", DefaultTotalRetryDuration))
	}

	if config.TokenRefreshBufferPeriod == 0 {
		config.TokenRefreshBufferPeriod = DefaultTokenBufferPeriod
		log.Info("TokenRefreshBufferPeriod not set, set to default value", zap.Duration("TokenRefreshBufferPeriod", DefaultTokenBufferPeriod))
	}

	if config.TotalRetryDuration == 0 {
		config.TotalRetryDuration = DefaultTotalRetryDuration
		log.Info("TotalRetryDuration not set, set to default value", zap.Duration("TotalRetryDuration", DefaultTotalRetryDuration))
	}

	if config.CustomTimeout == 0 {
		config.CustomTimeout = DefaultTimeout
		log.Info("CustomTimeout not set, set to default value", zap.Duration("CustomTimeout", DefaultTimeout))
	}

	// Determine the authentication method
	AuthMethod := "unknown"
	if config.Auth.Username != "" && config.Auth.Password != "" {
		AuthMethod = "bearer"
	} else if config.Auth.ClientID != "" && config.Auth.ClientSecret != "" {
		AuthMethod = "oauth"
	} else {
		return nil, log.Error("Invalid AuthConfig", zap.String("Username", config.Auth.Username), zap.String("ClientID", config.Auth.ClientID))
	}

	// Create a new HTTP client with the provided configuration.
	client := &Client{
		InstanceName:   config.Environment.APIType,
		APIHandler:     apiHandler,
		AuthMethod:     AuthMethod,
		httpClient:     &http.Client{Timeout: config.CustomTimeout},
		config:         config,
		Logger:         log,
		ConcurrencyMgr: NewConcurrencyManager(config.MaxConcurrentRequests, log, true),
		PerfMetrics:    PerformanceMetrics{},
	}

	// Log the client's configuration.
	log.Info("New API client initialized",
		zap.String("API Service", config.Environment.APIType),
		zap.String("Instance Name", client.InstanceName),
		zap.String("OverrideBaseDomain", config.Environment.OverrideBaseDomain),
		zap.String("AuthMethod", AuthMethod),
		zap.Int("MaxRetryAttempts", config.MaxRetryAttempts),
		zap.Int("MaxConcurrentRequests", config.MaxConcurrentRequests),
		zap.Bool("EnableDynamicRateLimiting", config.EnableDynamicRateLimiting),
		zap.Duration("TokenRefreshBufferPeriod", config.TokenRefreshBufferPeriod),
		zap.Duration("TotalRetryDuration", config.TotalRetryDuration),
		zap.Duration("CustomTimeout", config.CustomTimeout),
		zap.String("LogLevel", config.LogLevel.String()),
	)

	return client, nil

}
