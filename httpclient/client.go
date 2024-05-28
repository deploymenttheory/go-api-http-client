// httpclient/client.go
/* The `http_client` package provides a configurable HTTP client tailored for interacting with specific APIs.
It supports different authentication methods, including "bearer" and "oauth". The client is designed with a
focus on concurrency management, structured error handling, and flexible configuration options.
The package offers a default timeout, custom backoff strategies, dynamic rate limiting,
and detailed logging capabilities. The main `Client` structure encapsulates all necessary components,
like the baseURL, authentication details, and an embedded standard HTTP client. */
package httpclient

import (
	"net/http"
	"time"

	"github.com/deploymenttheory/go-api-http-client/authenticationhandler"
	"github.com/deploymenttheory/go-api-http-client/concurrency"
	"github.com/deploymenttheory/go-api-http-client/helpers"

	"github.com/deploymenttheory/go-api-http-client/logger"
	"github.com/deploymenttheory/go-api-http-client/redirecthandler"
	"go.uber.org/zap"
)

// region main
// Master struct/object
type Client struct {
	AuthMethod         string
	AuthToken          string
	AuthTokenExpiry    time.Time
	http               *http.Client
	config             ClientConfig
	Logger             logger.Logger
	ConcurrencyHandler *concurrency.ConcurrencyHandler
	APIHandler         APIHandler
	AuthTokenHandler   *authenticationhandler.AuthTokenHandler
}

//endregion

// region clientconfig

// Setup Options for Client
type ClientConfig struct {
	Auth          AuthConfig
	Environment   EnvironmentConfig
	ClientOptions ClientOptions
}

//endregion

// region authconfig

// Struct to map auth config from JSON.
type AuthConfig struct {
	Username     string `json:"Username,omitempty"`
	Password     string `json:"Password,omitempty"`
	ClientID     string `json:"ClientID,omitempty"`
	ClientSecret string `json:"ClientSecret,omitempty"`
}

//endregion

// region envconfig

// Struct to map environment config from JSON
type EnvironmentConfig struct {
	APIType            string `json:"APIType,omitempty"`            // APIType specifies the type of API integration to use // QUERY what are the types?!
	InstanceName       string `json:"InstanceName,omitempty"`       // Website Instance name without the root domain // NOTE Jamf specific
	OverrideBaseDomain string `json:"OverrideBaseDomain,omitempty"` // Base domain override used when the default in the api handler isn't suitable // NOTE ??
	TenantID           string `json:"TenantID,omitempty"`           // TenantID is the unique identifier for the tenant // QUERY what tenant?
	TenantName         string `json:"TenantName,omitempty"`         // TenantName is the name of the tenant // QUERY ?!?!
}

//endregion

// region clientoptions

// ClientOptions holds optional configuration options for the HTTP Client. // NOTE how is this difference from ClientConfig?!
type ClientOptions struct {
	Logging     LoggingConfig     // Configuration related to logging
	Cookies     CookieConfig      // Cookie handling settings
	Retry       RetryConfig       // Retry behavior configuration
	Concurrency ConcurrencyConfig // Concurrency configuration
	Timeout     TimeoutConfig     // Custom timeout settings
	Redirect    RedirectConfig    // Redirect handling settings
}

//endregion

// region loggingconfig

type LoggingConfig struct {
	LogLevel            string // Tiered logging level.
	LogOutputFormat     string // Output format of the logs. Use "JSON" for JSON format, "console" for human-readable format
	LogConsoleSeparator string // Separator in console output format.
	LogExportPath       string // Path to output logs to.
	HideSensitiveData   bool   // Whether sensitive fields should be hidden in logs.
}

//endregion

// region cookieconfig
// CookieConfig holds configuration related to cookie handling. // QUERY is this needed?
type CookieConfig struct {
	EnableCookieJar bool              // Enable or disable cookie jar
	CustomCookies   map[string]string `json:"CustomCookies,omitempty"` // Key-value pairs for setting specific cookies
}

//endregion

// region retryconfig
// RetryConfig holds configuration related to retry behavior. // QUERY is this needed?!
type RetryConfig struct {
	MaxRetryAttempts          int  // Maximum number of retry request attempts for retryable HTTP methods.
	EnableDynamicRateLimiting bool // Whether dynamic rate limiting should be enabled.
}

//endregion

// region concurrencyconfig
// ConcurrencyConfig holds configuration related to concurrency management. // QUERY and this?!
type ConcurrencyConfig struct {
	MaxConcurrentRequests int // Maximum number of concurrent requests allowed.
}

//endregion

// region timeoutconfig

type TimeoutConfig struct {
	CustomTimeout            helpers.JSONDuration // Custom timeout for the HTTP client
	TokenRefreshBufferPeriod helpers.JSONDuration // Buffer period before token expiry to attempt token refresh
	TotalRetryDuration       helpers.JSONDuration // Total duration to attempt retries
}

//endregion

// region redirectconfig
// RedirectConfig holds configuration related to redirect handling. // QUERY is this needed?!
type RedirectConfig struct {
	FollowRedirects bool // Enable or disable following redirects
	MaxRedirects    int  // Maximum number of redirects to follow
}

//endregion

// BuildClient creates a new HTTP client with the provided configuration.
func BuildClient(config ClientConfig) (*Client, error) {

	//region Logging
	// I'm not going down this rabbit hole yet.
	parsedLogLevel := logger.ParseLogLevelFromString(config.ClientOptions.Logging.LogLevel)
	log := logger.BuildLogger(parsedLogLevel, config.ClientOptions.Logging.LogOutputFormat, config.ClientOptions.Logging.LogConsoleSeparator, config.ClientOptions.Logging.LogExportPath)
	log.SetLevel(parsedLogLevel)

	//endregion

	//region API Handler

	// Use the APIType from the config to determine which API handler to load
	apiHandler, err := GetAPIHandler(config.Environment.APIType, config.Environment.InstanceName, config.Environment.TenantID, config.Environment.TenantName, log)
	if err != nil {
		log.Error("Failed to load API handler", zap.String("APIType", config.Environment.APIType), zap.Error(err))
		return nil, err
	}

	//endregion

	//region Auth

	// Determine the authentication method using the helper function
	authMethod, err := DetermineAuthMethod(config.Auth)
	if err != nil {
		log.Error("Failed to determine authentication method", zap.Error(err))
		return nil, err
	}

	// Initialize AuthTokenHandler
	clientCredentials := authenticationhandler.ClientCredentials{
		Username:     config.Auth.Username,
		Password:     config.Auth.Password,
		ClientID:     config.Auth.ClientID,
		ClientSecret: config.Auth.ClientSecret,
	}

	authTokenHandler := authenticationhandler.NewAuthTokenHandler(
		log,
		authMethod,
		clientCredentials,
		config.Environment.InstanceName,
		config.ClientOptions.Logging.HideSensitiveData,
	)

	//endregion

	//region HTTP

	log.Info("Initializing new HTTP client with the provided configuration")

	// Initialize the internal HTTP client
	httpClient := &http.Client{
		Timeout: config.ClientOptions.Timeout.CustomTimeout.Duration(),
	}

	//endregion

	// region COOKIES

	// Conditionally setup cookie jar
	// if err := SetupCookieJar(httpClient, config, log); err != nil {
	// 	log.Error("Error setting up cookie jar", zap.Error(err))
	// 	return nil, err
	// }

	//endregion

	//region Redirect?

	// Conditionally setup redirect handling
	if err := redirecthandler.SetupRedirectHandler(httpClient, config.ClientOptions.Redirect.FollowRedirects, config.ClientOptions.Redirect.MaxRedirects, log); err != nil {
		log.Error("Failed to set up redirect handler", zap.Error(err))
		return nil, err
	}

	//endregion

	//region Concurrency

	// Initialize ConcurrencyMetrics specifically for ConcurrencyHandler
	concurrencyMetrics := &concurrency.ConcurrencyMetrics{}

	// Initialize the ConcurrencyHandler with the newly created ConcurrencyMetrics
	concurrencyHandler := concurrency.NewConcurrencyHandler(
		config.ClientOptions.Concurrency.MaxConcurrentRequests,
		log,
		concurrencyMetrics,
	)

	//endregion

	//region Create

	// Create a new HTTP client with the provided configuration.
	client := &Client{
		APIHandler:         apiHandler,
		AuthMethod:         authMethod,
		httpClient:         httpClient,
		clientConfig:       config,
		Logger:             log,
		ConcurrencyHandler: concurrencyHandler,
		AuthTokenHandler:   authTokenHandler,
	}

	//endregion

	//region LoggingOut

	// Log the client's configuration.
	log.Info("New API client initialized",
		zap.String("API Type", config.Environment.APIType),
		zap.String("Instance Name", config.Environment.InstanceName),
		zap.String("Override Base Domain", config.Environment.OverrideBaseDomain),
		zap.String("Tenant ID", config.Environment.TenantID),
		zap.String("Tenant Name", config.Environment.TenantName),
		zap.String("Authentication Method", authMethod),
		zap.String("Logging Level", config.ClientOptions.Logging.LogLevel),
		zap.String("Log Encoding Format", config.ClientOptions.Logging.LogOutputFormat),
		zap.String("Log Separator", config.ClientOptions.Logging.LogConsoleSeparator),
		zap.Bool("Hide Sensitive Data In Logs", config.ClientOptions.Logging.HideSensitiveData),
		zap.Bool("Cookie Jar Enabled", config.ClientOptions.Cookies.EnableCookieJar),
		zap.Int("Max Retry Attempts", config.ClientOptions.Retry.MaxRetryAttempts),
		zap.Bool("Enable Dynamic Rate Limiting", config.ClientOptions.Retry.EnableDynamicRateLimiting),
		zap.Int("Max Concurrent Requests", config.ClientOptions.Concurrency.MaxConcurrentRequests),
		zap.Bool("Follow Redirects", config.ClientOptions.Redirect.FollowRedirects),
		zap.Int("Max Redirects", config.ClientOptions.Redirect.MaxRedirects),
		zap.Duration("Token Refresh Buffer Period", config.ClientOptions.Timeout.TokenRefreshBufferPeriod.Duration()),
		zap.Duration("Total Retry Duration", config.ClientOptions.Timeout.TotalRetryDuration.Duration()),
		zap.Duration("Custom Timeout", config.ClientOptions.Timeout.CustomTimeout.Duration()),
	)

	//endregion

	return client, nil

}
