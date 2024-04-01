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
	"github.com/deploymenttheory/go-api-http-client/proxy"
	"github.com/deploymenttheory/go-api-http-client/redirecthandler"
	"go.uber.org/zap"
)

// Client represents an HTTP client to interact with a specific API.
type Client struct {
	APIHandler         APIHandler // APIHandler interface used to define which API handler to use
	InstanceName       string     // Website Instance name without the root domain
	AuthMethod         string     // Specifies the authentication method: "bearer" or "oauth"
	Token              string     // Authentication Token
	OverrideBaseDomain string     // Base domain override used when the default in the api handler isn't suitable
	Expiry             time.Time  // Expiry time set for the auth token
	httpClient         *http.Client
	tokenLock          sync.Mutex
	clientConfig       ClientConfig
	Logger             logger.Logger
	ConcurrencyMgr     *ConcurrencyManager
	PerfMetrics        PerformanceMetrics
}

// Config holds configuration options for the HTTP Client.
type ClientConfig struct {
	Auth          AuthConfig        // User can either supply these values manually or pass from LoadAuthConfig/Env vars
	Environment   EnvironmentConfig // User can either supply these values manually or pass from LoadAuthConfig/Env vars
	ClientOptions ClientOptions     // Optional configuration options for the HTTP Client
	Proxy         ProxyConfig       // Proxy configuration options for the HTTP Client
}

// AuthConfig represents the structure to read authentication details from a JSON configuration file.
type AuthConfig struct {
	Username     string `json:"Username,omitempty"`
	Password     string `json:"Password,omitempty"`
	ClientID     string `json:"ClientID,omitempty"`
	ClientSecret string `json:"ClientSecret,omitempty"`
}

// EnvironmentConfig represents the structure to read authentication details from a JSON configuration file.
type EnvironmentConfig struct {
	InstanceName       string `json:"InstanceName,omitempty"`
	OverrideBaseDomain string `json:"OverrideBaseDomain,omitempty"`
	APIType            string `json:"APIType,omitempty"`
}

// ClientOptions holds optional configuration options for the HTTP Client.
type ClientOptions struct {
	EnableCookieJar           bool   // Field to enable or disable cookie jar
	LogLevel                  string // Field for defining tiered logging level.
	LogOutputFormat           string // Field for defining the output format of the logs. Use "JSON" for JSON format, "console" for human-readable format
	LogConsoleSeparator       string // Field for defining the separator in console output format.
	HideSensitiveData         bool   // Field for defining whether sensitive fields should be hidden in logs.
	MaxRetryAttempts          int    // Config item defines the max number of retry request attempts for retryable HTTP methods.
	EnableDynamicRateLimiting bool   // Field for defining whether dynamic rate limiting should be enabled.
	MaxConcurrentRequests     int    // Field for defining the maximum number of concurrent requests allowed in the semaphore
	FollowRedirects           bool   // Flag to enable/disable following redirects
	MaxRedirects              int    // Maximum number of redirects to follow
	TokenRefreshBufferPeriod  time.Duration
	TotalRetryDuration        time.Duration
	CustomTimeout             time.Duration
}

type ProxyConfig struct {
	ProxyURL       string `json:"ProxyURL,omitempty"`
	ProxyUsername  string `json:"ProxyUsername,omitempty"`
	ProxyPassword  string `json:"ProxyPassword,omitempty"`
	ProxyAuthToken string `json:"ProxyAuthToken,omitempty"`
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

	// Parse the log level string to logger.LogLevel
	parsedLogLevel := logger.ParseLogLevelFromString(config.ClientOptions.LogLevel)

	// Initialize the logger with parsed config values
	log := logger.BuildLogger(parsedLogLevel, config.ClientOptions.LogOutputFormat, config.ClientOptions.LogConsoleSeparator)

	// Set the logger's level (optional if BuildLogger already sets the level based on the input)
	log.SetLevel(parsedLogLevel)

	// Use the APIType from the config to determine which API handler to load
	apiHandler, err := LoadAPIHandler(config.Environment.APIType, log)
	if err != nil {
		return nil, log.Error("Failed to load API handler", zap.String("APIType", config.Environment.APIType), zap.Error(err))
	}

	// Determine the authentication method using the helper function
	authMethod, err := DetermineAuthMethod(config.Auth)
	if err != nil {
		log.Error("Failed to determine authentication method", zap.Error(err))
		return nil, err
	}

	log.Info("Initializing new HTTP client with the provided configuration")

	// Initialize the internal HTTP client
	httpClient := &http.Client{
		Timeout: config.ClientOptions.CustomTimeout,
	}

	// Conditionally setup cookie jar
	if err := setupCookieJar(httpClient, config.ClientOptions.EnableCookieJar, log); err != nil {
		log.Error("Error setting up cookie jar", zap.Error(err))
		return nil, err
	}

	// Conditionally setup redirect handling
	if err := redirecthandler.SetupRedirectHandler(httpClient, config.ClientOptions.FollowRedirects, config.ClientOptions.MaxRedirects, log); err != nil {
		log.Error("Failed to set up redirect handler", zap.Error(err))
		return nil, err
	}

	// Conditionally Initialize the proxy if provided in the configuration
	if err := proxy.InitializeProxy(httpClient, config.Proxy.ProxyURL, config.Proxy.ProxyUsername, config.Proxy.ProxyPassword, config.Proxy.ProxyAuthToken, log); err != nil {
		return nil, err
	}

	// Create a new HTTP client with the provided configuration.
	client := &Client{
		APIHandler:         apiHandler,
		InstanceName:       config.Environment.InstanceName,
		AuthMethod:         authMethod,
		OverrideBaseDomain: config.Environment.OverrideBaseDomain,
		httpClient:         httpClient,
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
		zap.String("Logging Level", config.ClientOptions.LogLevel),
		zap.String("Log Encoding Format", config.ClientOptions.LogOutputFormat),
		zap.String("Log Separator", config.ClientOptions.LogConsoleSeparator),
		zap.Bool("Hide Sensitive Data In Logs", config.ClientOptions.HideSensitiveData),
		zap.Bool("Cookie Jar Enabled", config.ClientOptions.EnableCookieJar),
		zap.Int("Max Retry Attempts", config.ClientOptions.MaxRetryAttempts),
		zap.Int("Max Concurrent Requests", config.ClientOptions.MaxConcurrentRequests),
		zap.Bool("Follow Redirects", config.ClientOptions.FollowRedirects),
		zap.Int("Max Redirects", config.ClientOptions.MaxRedirects),
		zap.Bool("Enable Dynamic Rate Limiting", config.ClientOptions.EnableDynamicRateLimiting),
		zap.Duration("Token Refresh Buffer Period", config.ClientOptions.TokenRefreshBufferPeriod),
		zap.Duration("Total Retry Duration", config.ClientOptions.TotalRetryDuration),
		zap.Duration("Custom Timeout", config.ClientOptions.CustomTimeout),
	)

	return client, nil

}
