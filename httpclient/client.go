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

	"github.com/deploymenttheory/go-api-http-client/apiintegrations/apihandler"
	"github.com/deploymenttheory/go-api-http-client/authenticationhandler"
	"github.com/deploymenttheory/go-api-http-client/concurrency"
	"github.com/deploymenttheory/go-api-http-client/helpers"

	"github.com/deploymenttheory/go-api-http-client/logger"
	"github.com/deploymenttheory/go-api-http-client/redirecthandler"
	"go.uber.org/zap"
)

// Client represents an HTTP client to interact with a specific API.
type Client struct {
	AuthMethod         string                                  // Specifies the authentication method: "bearer" or "oauth"
	Token              string                                  // Authentication Token
	Expiry             time.Time                               // Expiry time set for the auth token
	httpClient         *http.Client                            // Internal HTTP client
	clientConfig       ClientConfig                            // HTTP Client configuration
	Logger             logger.Logger                           // Logger for logging messages
	ConcurrencyHandler *concurrency.ConcurrencyHandler         // ConcurrencyHandler for managing concurrent requests
	APIHandler         apihandler.APIHandler                   // APIHandler interface used to define which API handler to use
	AuthTokenHandler   *authenticationhandler.AuthTokenHandler // AuthTokenHandler for managing authentication
}

// Config holds configuration options for the HTTP Client.
type ClientConfig struct {
	Auth          AuthConfig        // User can either supply these values manually or pass from LoadAuthConfig/Env vars
	Environment   EnvironmentConfig // User can either supply these values manually or pass from LoadAuthConfig/Env vars
	ClientOptions ClientOptions     // Optional configuration options for the HTTP Client
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
	APIType            string `json:"APIType,omitempty"`            // APIType specifies the type of API integration to use
	InstanceName       string `json:"InstanceName,omitempty"`       // Website Instance name without the root domain
	OverrideBaseDomain string `json:"OverrideBaseDomain,omitempty"` // Base domain override used when the default in the api handler isn't suitable
	TenantID           string `json:"TenantID,omitempty"`           // TenantID is the unique identifier for the tenant
	TenantName         string `json:"TenantName,omitempty"`         // TenantName is the name of the tenant
}

// ClientOptions holds optional configuration options for the HTTP Client.
type ClientOptions struct {
	Logging     LoggingConfig     // Configuration related to logging
	Cookies     CookieConfig      // Cookie handling settings
	Retry       RetryConfig       // Retry behavior configuration
	Concurrency ConcurrencyConfig // Concurrency configuration
	Timeout     TimeoutConfig     // Custom timeout settings
	Redirect    RedirectConfig    // Redirect handling settings
}

// LoggingConfig holds configuration options related to logging.
type LoggingConfig struct {
	LogLevel            string // Tiered logging level.
	LogOutputFormat     string // Output format of the logs. Use "JSON" for JSON format, "console" for human-readable format
	LogConsoleSeparator string // Separator in console output format.
	LogExportPath       string // Path to output logs to.
	HideSensitiveData   bool   // Whether sensitive fields should be hidden in logs.
}

// CookieConfig holds configuration related to cookie handling.
type CookieConfig struct {
	EnableCookieJar bool              // Enable or disable cookie jar
	CustomCookies   map[string]string `json:"CustomCookies,omitempty"` // Key-value pairs for setting specific cookies
}

// RetryConfig holds configuration related to retry behavior.
type RetryConfig struct {
	MaxRetryAttempts          int  // Maximum number of retry request attempts for retryable HTTP methods.
	EnableDynamicRateLimiting bool // Whether dynamic rate limiting should be enabled.
}

// ConcurrencyConfig holds configuration related to concurrency management.
type ConcurrencyConfig struct {
	MaxConcurrentRequests int // Maximum number of concurrent requests allowed.
}

// TimeoutConfig holds custom timeout settings.
// type TimeoutConfig struct {
// 	CustomTimeout            time.Duration // Custom timeout for the HTTP client
// 	TokenRefreshBufferPeriod time.Duration // Buffer period before token expiry to attempt token refresh
// 	TotalRetryDuration       time.Duration // Total duration to attempt retries
// }

type TimeoutConfig struct {
	CustomTimeout            helpers.JSONDuration // Custom timeout for the HTTP client
	TokenRefreshBufferPeriod helpers.JSONDuration // Buffer period before token expiry to attempt token refresh
	TotalRetryDuration       helpers.JSONDuration // Total duration to attempt retries
}

// RedirectConfig holds configuration related to redirect handling.
type RedirectConfig struct {
	FollowRedirects bool // Enable or disable following redirects
	MaxRedirects    int  // Maximum number of redirects to follow
}

// BuildClient creates a new HTTP client with the provided configuration.
func BuildClient(config ClientConfig) (*Client, error) {

	// Parse the log level string to logger.LogLevel
	parsedLogLevel := logger.ParseLogLevelFromString(config.ClientOptions.Logging.LogLevel)

	// Initialize the logger with parsed config values
	log := logger.BuildLogger(parsedLogLevel, config.ClientOptions.Logging.LogOutputFormat, config.ClientOptions.Logging.LogConsoleSeparator, config.ClientOptions.Logging.LogExportPath)

	// Set the logger's level (optional if BuildLogger already sets the level based on the input)
	log.SetLevel(parsedLogLevel)

	// Use the APIType from the config to determine which API handler to load
	apiHandler, err := apihandler.LoadAPIHandler(config.Environment.APIType, config.Environment.InstanceName, config.Environment.TenantID, config.Environment.TenantName, log)
	if err != nil {
		log.Error("Failed to load API handler", zap.String("APIType", config.Environment.APIType), zap.Error(err))
		return nil, err
	}

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

	log.Info("Initializing new HTTP client with the provided configuration")

	// Initialize the internal HTTP client
	httpClient := &http.Client{
		Timeout: config.ClientOptions.Timeout.CustomTimeout.Duration(),
	}

	// Conditionally setup cookie jar
	// if err := SetupCookieJar(httpClient, config, log); err != nil {
	// 	log.Error("Error setting up cookie jar", zap.Error(err))
	// 	return nil, err
	// }

	// Conditionally setup redirect handling
	if err := redirecthandler.SetupRedirectHandler(httpClient, config.ClientOptions.Redirect.FollowRedirects, config.ClientOptions.Redirect.MaxRedirects, log); err != nil {
		log.Error("Failed to set up redirect handler", zap.Error(err))
		return nil, err
	}

	// Initialize ConcurrencyMetrics specifically for ConcurrencyHandler
	concurrencyMetrics := &concurrency.ConcurrencyMetrics{}

	// Initialize the ConcurrencyHandler with the newly created ConcurrencyMetrics
	concurrencyHandler := concurrency.NewConcurrencyHandler(
		config.ClientOptions.Concurrency.MaxConcurrentRequests,
		log,
		concurrencyMetrics,
	)

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

	return client, nil

}

// // SetupCookieJar sets up the cookie jar for the HTTP client if enabled in the configuration.
// func SetupCookieJar(client *http.Client, clientConfig ClientConfig, log logger.Logger) error {
// 	if clientConfig.ClientOptions.Cookies.EnableCookieJar {
// 		jar, err := cookiejar.New(nil) // nil options use default options
// 		if err != nil {
// 			log.Error("Failed to create cookie jar", zap.Error(err))
// 			return fmt.Errorf("setupCookieJar failed: %w", err) // Wrap and return the error
// 		}

// 		if clientConfig.ClientOptions.Cookies.CustomCookies != nil {
// 			var CookieList []*http.Cookie
// 			CookieList = make([]*http.Cookie, 0)
// 			for k, v := range clientConfig.ClientOptions.Cookies.CustomCookies {
// 				newCookie := &http.Cookie{
// 					Name:  k,
// 					Value: v,
// 				}
// 				CookieList = append(CookieList, newCookie)
// 			}

// 			cookieUrl, err := url.Parse(fmt.Sprintf("http://%s.jamfcloud.com", clientConfig.Environment.InstanceName))
// 			if err != nil {
// 				return err
// 			}

// 			jar.SetCookies(cookieUrl, CookieList)
// 		}

// 		client.Jar = jar
// 	}
// 	return nil
// }
