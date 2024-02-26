package httpclient

import (
	"bufio"
	"encoding/json"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/deploymenttheory/go-api-http-client/logger"
	"go.uber.org/zap"
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

func SetClientConfiguration(log *zap.Logger, configFilePath string) (*ClientConfig, error) {
	config := &ClientConfig{}

	// Load config values from environment variables
	loadConfigFromEnv(config)

	// Load config values from file if necessary
	if validateConfigCompletion(config) && configFilePath != "" {
		if err := config.loadConfigFromFile(configFilePath, log); err != nil { // Updated to the correct function name
			log.Error("Failed to load configuration from file", zap.String("path", configFilePath), zap.Error(err))
			return nil, err
		}
	}

	// Validate the configuration
	if err := validateClientConfig(config, log); err != nil {
		log.Error("Configuration validation failed", zap.Error(err))
		return nil, err
	}

	// Set default values where necessary
	setClientDefaultValues(config, log)

	return config, nil
}

func loadConfigFromEnv(config *ClientConfig) {
	// AuthConfig
	config.Auth.ClientID = getEnvOrDefault("CLIENT_ID", config.Auth.ClientID)
	config.Auth.ClientSecret = getEnvOrDefault("CLIENT_SECRET", config.Auth.ClientSecret)

	// EnvironmentConfig
	config.Environment.InstanceName = getEnvOrDefault("INSTANCE_NAME", config.Environment.InstanceName)
	config.Environment.OverrideBaseDomain = getEnvOrDefault("OVERRIDE_BASE_DOMAIN", config.Environment.OverrideBaseDomain)
	config.Environment.APIType = getEnvOrDefault("API_TYPE", config.Environment.APIType)

	// ClientOptions
	config.ClientOptions.LogLevel = getEnvOrDefault("LOG_LEVEL", config.ClientOptions.LogLevel)
	config.ClientOptions.LogOutputFormat = getEnvOrDefault("LOG_OUTPUT_FORMAT", config.ClientOptions.LogOutputFormat)
	config.ClientOptions.LogConsoleSeparator = getEnvOrDefault("LOG_CONSOLE_SEPARATOR", config.ClientOptions.LogConsoleSeparator)
	config.ClientOptions.HideSensitiveData = parseBool(getEnvOrDefault("HIDE_SENSITIVE_DATA", strconv.FormatBool(config.ClientOptions.HideSensitiveData)))
	config.ClientOptions.MaxRetryAttempts = parseInt(getEnvOrDefault("MAX_RETRY_ATTEMPTS", strconv.Itoa(config.ClientOptions.MaxRetryAttempts)), DefaultMaxRetryAttempts)
	config.ClientOptions.EnableDynamicRateLimiting = parseBool(getEnvOrDefault("ENABLE_DYNAMIC_RATE_LIMITING", strconv.FormatBool(config.ClientOptions.EnableDynamicRateLimiting)))
	config.ClientOptions.MaxConcurrentRequests = parseInt(getEnvOrDefault("MAX_CONCURRENT_REQUESTS", strconv.Itoa(config.ClientOptions.MaxConcurrentRequests)), DefaultMaxConcurrentRequests)
	config.ClientOptions.TokenRefreshBufferPeriod = parseDuration(getEnvOrDefault("TOKEN_REFRESH_BUFFER_PERIOD", config.ClientOptions.TokenRefreshBufferPeriod.String()), DefaultTokenBufferPeriod)
	config.ClientOptions.TotalRetryDuration = parseDuration(getEnvOrDefault("TOTAL_RETRY_DURATION", config.ClientOptions.TotalRetryDuration.String()), DefaultTotalRetryDuration)
	config.ClientOptions.CustomTimeout = parseDuration(getEnvOrDefault("CUSTOM_TIMEOUT", config.ClientOptions.CustomTimeout.String()), DefaultTimeout)
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

// validateConfigCompletion checks if any essential configuration fields are missing,
// indicating the configuration might be incomplete and may require loading from additional sources.
func validateConfigCompletion(config *ClientConfig) bool {
	// Check if essential fields are missing; additional fields can be checked as needed
	return config.Auth.ClientID == "" || config.Auth.ClientSecret == "" ||
		config.Environment.InstanceName == "" || config.Environment.APIType == "" ||
		config.ClientOptions.LogLevel == "" || config.ClientOptions.LogOutputFormat == "" ||
		config.ClientOptions.LogConsoleSeparator == ""
}

func setClientDefaultValues(config *ClientConfig, log *zap.Logger) {
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

	// Log completion of setting default values
	log.Info("Default values set for client configuration")
}

// loadFromFile loads configuration values from a JSON file into the ClientConfig struct.
func (config *ClientConfig) loadConfigFromFile(filePath string, log *zap.Logger) error {
	// Open the configuration file
	file, err := os.Open(filePath)
	if err != nil {
		log.Error("Failed to open the configuration file", zap.String("filePath", filePath), zap.Error(err))
		return err
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	var builder strings.Builder

	// Read the file content
	for {
		part, _, err := reader.ReadLine()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Error("Failed to read the configuration file", zap.String("filePath", filePath), zap.Error(err))
			return err
		}
		builder.Write(part)
	}

	// Unmarshal JSON content into the ClientConfig struct
	err = json.Unmarshal([]byte(builder.String()), config)
	if err != nil {
		log.Error("Failed to unmarshal the configuration file", zap.String("filePath", filePath), zap.Error(err))
		return err
	}

	log.Info("Configuration successfully loaded from file", zap.String("filePath", filePath))
	return nil
}

// validateClientConfig checks the client configuration and logs any issues using zap logger.
func validateClientConfig(config *ClientConfig, log *zap.Logger) {
	// Helper function to log validation errors
	logValidationError := func(msg string) {
		log.Error(msg)
	}

	// Flag to track if any validation errors were found
	var hasErrors bool

	// Validate AuthConfig
	if config.Auth.ClientID == "" {
		logValidationError("Validation error: ClientID in AuthConfig is required")
		hasErrors = true
	}
	if config.Auth.ClientSecret == "" {
		logValidationError("Validation error: ClientSecret in AuthConfig is required")
		hasErrors = true
	}

	// Validate EnvironmentConfig
	if config.Environment.InstanceName == "" {
		logValidationError("Validation error: InstanceName in EnvironmentConfig is required")
		hasErrors = true
	}
	if config.Environment.APIType == "" {
		logValidationError("Validation error: APIType in EnvironmentConfig is required")
		hasErrors = true
	}

	// Validate ClientOptions
	if config.ClientOptions.LogLevel == "" {
		logValidationError("Validation error: LogLevel in ClientOptions is required")
		hasErrors = true
	}
	if config.ClientOptions.LogOutputFormat == "" {
		logValidationError("Validation error: LogOutputFormat in ClientOptions is required")
		hasErrors = true
	}
	if config.ClientOptions.MaxRetryAttempts < 0 {
		logValidationError("Validation error: MaxRetryAttempts in ClientOptions must not be negative")
		hasErrors = true
	}
	if config.ClientOptions.MaxConcurrentRequests <= 0 {
		logValidationError("Validation error: MaxConcurrentRequests in ClientOptions must be greater than 0")
		hasErrors = true
	}
	if config.ClientOptions.TokenRefreshBufferPeriod < 0 {
		logValidationError("Validation error: TokenRefreshBufferPeriod in ClientOptions must not be negative")
		hasErrors = true
	}
	if config.ClientOptions.TotalRetryDuration < 0 {
		logValidationError("Validation error: TotalRetryDuration in ClientOptions must not be negative")
		hasErrors = true
	}
	if config.ClientOptions.CustomTimeout <= 0 {
		logValidationError("Validation error: CustomTimeout in ClientOptions must be greater than 0")
		hasErrors = true
	}

	// Log a summary error if any validation errors were found
	if hasErrors {
		log.Error("Configuration validation failed with one or more errors")
	}
}
