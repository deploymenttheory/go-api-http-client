package httpclient

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/deploymenttheory/go-api-http-client/logger"
)

const (
	DefaultLogLevel                  = logger.LogLevelInfo
	DefaultMaxRetryAttempts          = 3
	DefaultEnableDynamicRateLimiting = true
	DefaultMaxConcurrentRequests     = 5
	DefaultTokenBufferPeriod         = 5 * time.Minute
	DefaultTotalRetryDuration        = 5 * time.Minute
	DefaultTimeout                   = 10 * time.Second
)

// LoadConfigFromFile loads configuration values from a JSON file into the ClientConfig struct.
// This function opens the specified configuration file, reads its content, and unmarshals the JSON data
// into the ClientConfig struct. It's designed to initialize the client configuration with values
// from a file, complementing or overriding defaults and environment variable settings.
// LoadConfigFromFile loads configuration values from a JSON file into the ClientConfig struct.
func LoadConfigFromFile(filePath string) (*ClientConfig, error) {
	// Read the entire file
	fileBytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read the configuration file: %s, error: %w", filePath, err)
	}

	// Initialize an instance of ClientConfig
	var config ClientConfig

	// Unmarshal the file content into the ClientConfig struct
	if err := json.Unmarshal(fileBytes, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal the configuration file: %s, error: %w", filePath, err)
	}

	log.Printf("Configuration successfully loaded from file: %s", filePath)

	// Set default values if necessary
	setLoggerDefaultValues(&config)
	setClientDefaultValues(&config)

	// Validate mandatory configuration fields
	if err := validateMandatoryConfiguration(&config); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

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
	config.Environment.InstanceName = getEnvOrDefault("INSTANCE_NAME", config.Environment.InstanceName)
	log.Printf("InstanceName env value found and set to: %s", config.Environment.InstanceName)

	config.Environment.OverrideBaseDomain = getEnvOrDefault("OVERRIDE_BASE_DOMAIN", config.Environment.OverrideBaseDomain)
	log.Printf("OverrideBaseDomain env value found and set to: %s", config.Environment.OverrideBaseDomain)

	config.Environment.APIType = getEnvOrDefault("API_TYPE", config.Environment.APIType)
	log.Printf("APIType env value found and set to: %s", config.Environment.APIType)

	// ClientOptions
	config.ClientOptions.LogLevel = getEnvOrDefault("LOG_LEVEL", config.ClientOptions.LogLevel)
	log.Printf("LogLevel env value found and set to: %s", config.ClientOptions.LogLevel)

	config.ClientOptions.LogOutputFormat = getEnvOrDefault("LOG_OUTPUT_FORMAT", config.ClientOptions.LogOutputFormat)
	log.Printf("LogOutputFormat env value found and set to: %s", config.ClientOptions.LogOutputFormat)

	config.ClientOptions.LogConsoleSeparator = getEnvOrDefault("LOG_CONSOLE_SEPARATOR", config.ClientOptions.LogConsoleSeparator)
	log.Printf("LogConsoleSeparator env value found and set to: %s", config.ClientOptions.LogConsoleSeparator)

	config.ClientOptions.HideSensitiveData = parseBool(getEnvOrDefault("HIDE_SENSITIVE_DATA", strconv.FormatBool(config.ClientOptions.HideSensitiveData)))
	log.Printf("HideSensitiveData env value found and set to: %t", config.ClientOptions.HideSensitiveData)

	config.ClientOptions.MaxRetryAttempts = parseInt(getEnvOrDefault("MAX_RETRY_ATTEMPTS", strconv.Itoa(config.ClientOptions.MaxRetryAttempts)), DefaultMaxRetryAttempts)
	log.Printf("MaxRetryAttempts env value found and set to: %d", config.ClientOptions.MaxRetryAttempts)

	config.ClientOptions.EnableDynamicRateLimiting = parseBool(getEnvOrDefault("ENABLE_DYNAMIC_RATE_LIMITING", strconv.FormatBool(config.ClientOptions.EnableDynamicRateLimiting)))
	log.Printf("EnableDynamicRateLimiting env value found and set to: %t", config.ClientOptions.EnableDynamicRateLimiting)

	config.ClientOptions.MaxConcurrentRequests = parseInt(getEnvOrDefault("MAX_CONCURRENT_REQUESTS", strconv.Itoa(config.ClientOptions.MaxConcurrentRequests)), DefaultMaxConcurrentRequests)
	log.Printf("MaxConcurrentRequests env value found and set to: %d", config.ClientOptions.MaxConcurrentRequests)

	config.ClientOptions.TokenRefreshBufferPeriod = parseDuration(getEnvOrDefault("TOKEN_REFRESH_BUFFER_PERIOD", config.ClientOptions.TokenRefreshBufferPeriod.String()), DefaultTokenBufferPeriod)
	log.Printf("TokenRefreshBufferPeriod env value found and set to: %s", config.ClientOptions.TokenRefreshBufferPeriod)

	config.ClientOptions.TotalRetryDuration = parseDuration(getEnvOrDefault("TOTAL_RETRY_DURATION", config.ClientOptions.TotalRetryDuration.String()), DefaultTotalRetryDuration)
	log.Printf("TotalRetryDuration env value found and set to: %s", config.ClientOptions.TotalRetryDuration)

	config.ClientOptions.CustomTimeout = parseDuration(getEnvOrDefault("CUSTOM_TIMEOUT", config.ClientOptions.CustomTimeout.String()), DefaultTimeout)
	log.Printf("CustomTimeout env value found and set to: %s", config.ClientOptions.CustomTimeout)

	// Set default values if necessary
	setLoggerDefaultValues(config)
	setClientDefaultValues(config)

	// Validate final configuration
	if err := validateMandatoryConfiguration(config); err != nil {
		return nil, err // Return the error if the configuration is incomplete
	}

	return config, nil
}

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

// validateMandatoryConfiguration checks if any essential configuration fields are missing,
// and returns an error with details about the missing configurations.
// This ensures the caller can understand what specific configurations need attention.
func validateMandatoryConfiguration(config *ClientConfig) error {
	var missingFields []string

	// Check for mandatory fields related to the environment
	if config.Environment.InstanceName == "" {
		missingFields = append(missingFields, "Environment.InstanceName")
	}
	if config.Environment.APIType == "" {
		missingFields = append(missingFields, "Environment.APIType")
	}

	// Check for mandatory fields related to the client options
	if config.ClientOptions.LogLevel == "" {
		missingFields = append(missingFields, "ClientOptions.LogLevel")
	}
	if config.ClientOptions.LogOutputFormat == "" {
		missingFields = append(missingFields, "ClientOptions.LogOutputFormat")
	}
	if config.ClientOptions.LogConsoleSeparator == "" {
		missingFields = append(missingFields, "ClientOptions.LogConsoleSeparator")
	}

	// Check for either OAuth credentials pair or Username and Password pair
	usingOAuth := config.Auth.ClientID != "" && config.Auth.ClientSecret != ""
	usingBasicAuth := config.Auth.Username != "" && config.Auth.Password != ""

	if !(usingOAuth || usingBasicAuth) {
		if config.Auth.ClientID == "" {
			missingFields = append(missingFields, "Auth.ClientID")
		}
		if config.Auth.ClientSecret == "" {
			missingFields = append(missingFields, "Auth.ClientSecret")
		}
		if config.Auth.Username == "" {
			missingFields = append(missingFields, "Auth.Username")
		}
		if config.Auth.Password == "" {
			missingFields = append(missingFields, "Auth.Password")
		}
	}

	// If there are missing fields, construct and return an error message detailing what is missing
	if len(missingFields) > 0 {
		errorMessage := fmt.Sprintf("Mandatory configuration missing: %s. Ensure that either OAuth credentials (ClientID and ClientSecret) or Basic Auth credentials (Username and Password) are fully provided.", strings.Join(missingFields, ", "))
		return fmt.Errorf(errorMessage)
	}

	// If no fields are missing, return nil indicating the configuration is complete
	return nil
}

// setClientDefaultValues sets default values for the client configuration options if none are provided.
// It checks each configuration option and sets it to the default value if it is either negative, zero,
// or not set. This function ensures that the configuration adheres to expected minimums or defaults,
// enhancing robustness and fault tolerance. It uses the standard log package for logging, ensuring that
// default value settings are transparent before the zap logger is initialized.
func setClientDefaultValues(config *ClientConfig) {
	if config.ClientOptions.MaxRetryAttempts < 0 {
		config.ClientOptions.MaxRetryAttempts = DefaultMaxRetryAttempts
		log.Printf("MaxRetryAttempts was negative, set to default value: %d", DefaultMaxRetryAttempts)
	}

	if config.ClientOptions.MaxConcurrentRequests <= 0 {
		config.ClientOptions.MaxConcurrentRequests = DefaultMaxConcurrentRequests
		log.Printf("MaxConcurrentRequests was negative or zero, set to default value: %d", DefaultMaxConcurrentRequests)
	}

	if config.ClientOptions.TokenRefreshBufferPeriod < 0 {
		config.ClientOptions.TokenRefreshBufferPeriod = DefaultTokenBufferPeriod
		log.Printf("TokenRefreshBufferPeriod was negative, set to default value: %s", DefaultTokenBufferPeriod)
	}

	if config.ClientOptions.TotalRetryDuration <= 0 {
		config.ClientOptions.TotalRetryDuration = DefaultTotalRetryDuration
		log.Printf("TotalRetryDuration was negative or zero, set to default value: %s", DefaultTotalRetryDuration)
	}

	if config.ClientOptions.TokenRefreshBufferPeriod == 0 {
		config.ClientOptions.TokenRefreshBufferPeriod = DefaultTokenBufferPeriod
		log.Printf("TokenRefreshBufferPeriod not set, set to default value: %s", DefaultTokenBufferPeriod)
	}

	if config.ClientOptions.TotalRetryDuration == 0 {
		config.ClientOptions.TotalRetryDuration = DefaultTotalRetryDuration
		log.Printf("TotalRetryDuration not set, set to default value: %s", DefaultTotalRetryDuration)
	}

	if config.ClientOptions.CustomTimeout == 0 {
		config.ClientOptions.CustomTimeout = DefaultTimeout
		log.Printf("CustomTimeout not set, set to default value: %s", DefaultTimeout)
	}

	// Log completion of setting default values
	log.Println("Default values set for client configuration")
}

// setLoggerDefaultValues sets default values for the client logger configuration options if none are provided.
// It checks each configuration option and sets it to the default value if it is either negative, zero,
// or not set. It also logs each default value being set.
func setLoggerDefaultValues(config *ClientConfig) {
	// Set default value if none is provided
	if config.ClientOptions.LogConsoleSeparator == "" {
		config.ClientOptions.LogConsoleSeparator = ","
		log.Println("LogConsoleSeparator not set, set to default value: ,")
	}

	// Log completion of setting default values
	log.Println("Default values set for logger configuration")
}
