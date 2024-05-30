// httpclient/client_configuration.go
// Description: This file contains functions to load and validate configuration values from a JSON file or environment variables.
package httpclient

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"slices"
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

	return &config, nil
}

func LoadConfigFromEnv() (*ClientConfig, error) {
	// TODO this whole function with settable env keys
	return nil, nil
}

// TODO try to get all the "valid list of x" strings out. Can't make them constants though? (and this func string)
func validateClientConfig(config ClientConfig) error {
	var err error

	// region Auth
	// Method
	validAuthMethods := []string{"oauth2", "basic"}
	if !slices.Contains(validAuthMethods, config.AuthMethod) {
		return fmt.Errorf("auth method not valid: %s", config.AuthMethod)
	}

	// Creds per Method
	switch config.AuthMethod {
	case "oauth2":
		err = validateValidClientID(config.ClientID)
		if err != nil {
			return err
		}
		err = validateClientSecret(config.ClientSecret)
		if err != nil {
			return err
		}
	case "basic":
		err = validateUsername(config.BasicAuthUsername)
		if err != nil {
			return err
		}

		err = validatePassword(config.BasicAuthPassword)
		if err != nil {
			return err
		}
	}

	// endregion

	// region Logging

	// Level
	validLogLevels := []string{
		"LogLevelDebug",
		"LogLevelInfo",
		"LogLevelWarn",
		"LogLevelError",
		"LogLevelPanic",
		"LogLevelFatal",
	}
	if !slices.Contains(validLogLevels, config.LogLevel) {
		return fmt.Errorf("invalid log level: %s", config.LogLevel)
	}

	// Console Format
	validLogFormats := []string{
		"json",
		"pretty",
	}

	if !slices.Contains(validLogFormats, config.LogOutputFormat) {
		return fmt.Errorf("invalid log output format: %s", config.LogOutputFormat)
	}

	// Log Console Separator
	// any string fine

	// Export Logs
	// bool

	// Log Export Path
	if config.ExportLogs {
		_, err = validateFilePath(config.LogExportPath)
		if err != nil {
			return err
		}
	}

	// Hide Sensitive Data
	// bool

	// endregion

	// region Cookies

	// CookieJar
	// bool

	// CustomCookies
	// no validation required

	// region Misc

	// Max Retry Attempts
	if config.MaxRetryAttempts < 0 {
		return errors.New("max retry attempts cannot be less than 0")
	}

	// Dynamic Rate Limiting
	// bool

	// Max Concurrent Requests
	if config.MaxConcurrentRequests < 0 {
		return errors.New("max concurrent requests cannot be less than 0")
	}

	// CustomTimeout
	if config.CustomTimeout.Seconds() < 0 {
		return errors.New("timeout cannot be less than 0 seconds")
	}

	// Token refesh buffer
	if config.TokenRefreshBufferPeriod.Seconds() < 0 {
		return errors.New("refresh buffer period cannot be less than 0 seconds")
	}

	if config.TotalRetryDuration.Seconds() < 0 {
		return errors.New("total retry duration cannot be less than 0 seconds")
	}

	// Follow redirects
	// bool

	// MaxRedirects
	if config.FollowRedirects {
		if MaxRedirects < 1 {
			return errors.New("max redirects cannot be less than 1")
		}
	}

	// endregion

	return nil
}
