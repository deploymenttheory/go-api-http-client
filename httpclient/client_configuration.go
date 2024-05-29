// httpclient/client_configuration.go
// Description: This file contains functions to load and validate configuration values from a JSON file or environment variables.
package httpclient

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/deploymenttheory/go-api-http-client/helpers"
	"github.com/deploymenttheory/go-api-http-client/logger"
)

const (
	DefaultLogLevel                  = logger.LogLevelInfo
	DefaultMaxRetryAttempts          = 3
	DefaultEnableDynamicRateLimiting = true
	DefaultMaxConcurrentRequests     = 5
	DefaultTokenBufferPeriod         = helpers.JSONDuration(5 * time.Minute)
	DefaultTotalRetryDuration        = helpers.JSONDuration(5 * time.Minute)
	DefaultTimeout                   = helpers.JSONDuration(10 * time.Second)
	FollowRedirects                  = true
	MaxRedirects                     = 10
	ConfigFileExtension              = ".json"
)

// LoadConfigFromFile loads configuration values from a JSON file into the ClientConfig struct.
// This function opens the specified configuration file, reads its content, and unmarshals the JSON data
// into the ClientConfig struct. It's designed to initialize the client configuration with values
// from a file, complementing or overriding defaults and environment variable settings.
func LoadConfigFromFile(filepath string) (*ClientConfig, error) {

	filepath, err := validateFilePath(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to clean/validate filepath (%s): %v", filepath, err)
	}

	fileBytes, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read the configuration file: %s, error: %w", filepath, err)
	}

	var config ClientConfig
	if err := json.Unmarshal(fileBytes, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal the configuration file: %s, error: %w", filepath, err)
	}

	// TODO Set default values if necessary and validate the configuration

	return &config, nil
}

// LoadConfigFromEnv populates the ClientConfig structure with values from environment variables.
// It updates the configuration for authentication, environment specifics, and client options
// based on the presence of environment variables. For each configuration option, if an environment
// variable is set, its value is used; otherwise, the existing value in the ClientConfig structure
// is retained. It also sets default values if necessary and validates the final configuration,
// returning an error if the configuration is incomplete.
func LoadConfigFromEnv(config *ClientConfig) (*ClientConfig, error) {
	if config == nil {
		config = &ClientConfig{} // Initialize config if nil
	}

	// AuthConfig
	config.Auth.Username = getEnvOrDefault("USERNAME", config.Auth.Username)
	log.Printf("Username env value found and set to: %s", config.Auth.Username)

	config.Auth.Password = getEnvOrDefault("PASSWORD", config.Auth.Password)
	log.Printf("Password env value found and set")

	config.Auth.ClientID = getEnvOrDefault("CLIENT_ID", config.Auth.ClientID)
	log.Printf("ClientID env value found and set to: %s", config.Auth.ClientID)

	config.Auth.ClientSecret = getEnvOrDefault("CLIENT_SECRET", config.Auth.ClientSecret)
	log.Printf("ClientSecret env value found and set")

	// EnvironmentConfig
	config.Environment.APIType = getEnvOrDefault("API_TYPE", config.Environment.APIType)
	log.Printf("APIType env value found and set to: %s", config.Environment.APIType)

	config.Environment.InstanceName = getEnvOrDefault("INSTANCE_NAME", config.Environment.InstanceName)
	log.Printf("InstanceName env value found and set to: %s", config.Environment.InstanceName)

	config.Environment.OverrideBaseDomain = getEnvOrDefault("OVERRIDE_BASE_DOMAIN", config.Environment.OverrideBaseDomain)
	log.Printf("OverrideBaseDomain env value found and set to: %s", config.Environment.OverrideBaseDomain)

	config.Environment.TenantID = getEnvOrDefault("TENANT_ID", config.Environment.TenantID)
	log.Printf("TenantID env value found and set to: %s", config.Environment.TenantID)

	config.Environment.TenantName = getEnvOrDefault("TENANT_NAME", config.Environment.TenantName)
	log.Printf("TenantName env value found and set to: %s", config.Environment.TenantName)

	// ClientOptions

	// Logging
	config.LogLevel = getEnvOrDefault("LOG_LEVEL", config.LogLevel)
	log.Printf("LogLevel env value found and set to: %s", config.LogLevel)

	config.LogOutputFormat = getEnvOrDefault("LOG_OUTPUT_FORMAT", config.LogOutputFormat)
	log.Printf("LogOutputFormat env value found and set to: %s", config.LogOutputFormat)

	config.LogConsoleSeparator = getEnvOrDefault("LOG_CONSOLE_SEPARATOR", config.LogConsoleSeparator)
	log.Printf("LogConsoleSeparator env value found and set to: %s", config.LogConsoleSeparator)

	config.LogExportPath = getEnvOrDefault("LOG_EXPORT_PATH", config.LogExportPath)
	log.Printf("LogExportPath env value found and set to: %s", config.LogExportPath)

	config.HideSensitiveData = parseBool(getEnvOrDefault("HIDE_SENSITIVE_DATA", strconv.FormatBool(config.HideSensitiveData)))
	log.Printf("HideSensitiveData env value found and set to: %t", config.HideSensitiveData)

	// Cookies
	config.EnableCookieJar = parseBool(getEnvOrDefault("ENABLE_COOKIE_JAR", strconv.FormatBool(config.EnableCookieJar)))
	log.Printf("EnableCookieJar env value found and set to: %t", config.EnableCookieJar)

	// Load specific cookies from environment variable
	cookieStr := getEnvOrDefault("CUSTOM_COOKIES", "")
	if cookieStr != "" {
		config.CustomCookies = parseCookiesFromString(cookieStr)
		log.Printf("CustomCookies env value found and set")
	}

	// Retry
	config.MaxRetryAttempts = parseInt(getEnvOrDefault("MAX_RETRY_ATTEMPTS", strconv.Itoa(config.MaxRetryAttempts)), DefaultMaxRetryAttempts)
	log.Printf("MaxRetryAttempts env value found and set to: %d", config.MaxRetryAttempts)

	config.EnableDynamicRateLimiting = parseBool(getEnvOrDefault("ENABLE_DYNAMIC_RATE_LIMITING", strconv.FormatBool(config.EnableDynamicRateLimiting)))
	log.Printf("EnableDynamicRateLimiting env value found and set to: %t", config.EnableDynamicRateLimiting)

	// Concurrency
	config.MaxConcurrentRequests = parseInt(getEnvOrDefault("MAX_CONCURRENT_REQUESTS", strconv.Itoa(config.MaxConcurrentRequests)), DefaultMaxConcurrentRequests)
	log.Printf("MaxConcurrentRequests env value found and set to: %d", config.MaxConcurrentRequests)

	// timeouts
	config.TokenRefreshBufferPeriod = helpers.ParseJSONDuration(getEnvOrDefault("TOKEN_REFRESH_BUFFER_PERIOD", config.TokenRefreshBufferPeriod.String()), DefaultTokenBufferPeriod)
	log.Printf("TokenRefreshBufferPeriod env value found and set to: %s", config.TokenRefreshBufferPeriod)

	config.TotalRetryDuration = helpers.ParseJSONDuration(getEnvOrDefault("TOTAL_RETRY_DURATION", config.TotalRetryDuration.String()), DefaultTotalRetryDuration)
	log.Printf("TotalRetryDuration env value found and set to: %s", config.TotalRetryDuration)

	config.CustomTimeout = helpers.ParseJSONDuration(getEnvOrDefault("CUSTOM_TIMEOUT", config.CustomTimeout.String()), DefaultTimeout)
	log.Printf("CustomTimeout env value found and set to: %s", config.CustomTimeout)

	// Redirects
	config.FollowRedirects = parseBool(getEnvOrDefault("FOLLOW_REDIRECTS", strconv.FormatBool(config.FollowRedirects)))
	log.Printf("FollowRedirects env value set to: %t", config.FollowRedirects)

	config.MaxRedirects = parseInt(getEnvOrDefault("MAX_REDIRECTS", strconv.Itoa(config.MaxRedirects)), MaxRedirects)
	log.Printf("MaxRedirects env value set to: %d", config.MaxRedirects)

	// Set default values if necessary
	setLoggerDefaultValues(config)
	setClientDefaultValues(config)

	// Validate final configuration
	if err := validateMandatoryConfiguration(config); err != nil {
		return nil, err // Return the error if the configuration is incomplete
	}

	return config, nil
}

// validateMandatoryConfiguration checks if any essential configuration fields are missing,
// and returns an error with details about the missing configurations.
// This ensures the caller can understand what specific configurations need attention.

// setClientDefaultValues sets default values for the client configuration options if none are provided.
// It checks each configuration option and sets it to the default value if it is either negative, zero,
// or not set. This function ensures that the configuration adheres to expected minimums or defaults,
// enhancing robustness and fault tolerance. It uses the standard log package for logging, ensuring that
// default value settings are transparent before the zap logger is initialized.

// Helper function to get environment variable or default value
func getEnvOrDefault(envKey string, defaultValue string) string {
	if value, exists := os.LookupEnv(envKey); exists {
		return value
	}
	return defaultValue
}

// Helper function to parse boolean from environment variable
func parseBool(value string) bool {
	result, err := strconv.ParseBool(value)
	if err != nil {
		return false
	}
	return result
}

// Helper function to parse int from environment variable
func parseInt(value string, defaultVal int) int {
	result, err := strconv.Atoi(value)
	if err != nil {
		return defaultVal
	}
	return result
}

// Helper function to parse duration from environment variable
func parseDuration(value string, defaultVal time.Duration) time.Duration {
	result, err := time.ParseDuration(value)
	if err != nil {
		return defaultVal
	}
	return result
}

// parseCookiesFromString parses a semi-colon separated string of key=value pairs into a map.
func parseCookiesFromString(cookieStr string) map[string]string {
	cookies := make(map[string]string)
	pairs := strings.Split(cookieStr, ";")
	for _, pair := range pairs {
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) == 2 {
			key := strings.TrimSpace(kv[0])
			value := strings.TrimSpace(kv[1])
			cookies[key] = value
		}
	}
	return cookies
}
