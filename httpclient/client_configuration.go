// httpclient/client_configuration.go
// Description: This file contains functions to load and validate configuration values from a JSON file or environment variables.
package httpclient

import (
	"errors"
	"fmt"
	"slices"
	"time"
)

const (
	DefaultLogLevelString            = "LogLevelInfo"
	DefaultLogOutputFormatString     = "pretty"
	DefaultLogConsoleSeparator       = "	"
	DefaultLogExportPath             = "/defaultlogs"
	DefaultMaxRetryAttempts          = 3
	DefaultMaxConcurrentRequests     = 1
	DefaultExportLogs                = false
	DefaultHideSensitiveData         = false
	DefaultEnableDynamicRateLimiting = false
	DefaultCustomTimeout             = 5 * time.Second
	DefaultTokenRefreshBufferPeriod  = 2 * time.Minute
	DefaultTotalRetryDuration        = 5 * time.Minute
	DefaultFollowRedirects           = false
	DefaultMaxRedirects              = 5
)

// LoadConfigFromFile loads configuration values from a JSON file into the ClientConfig struct.
// This function opens the specified configuration file, reads its content, and unmarshals the JSON data
// into the ClientConfig struct. It's designed to initialize the client configuration with values
// from a file, complementing or overriding defaults and environment variable settings.
func LoadConfigFromFile(filepath string) (*ClientConfig, error) {
	// TODO this whole function
	return nil, nil
}

func LoadConfigFromEnv() (*ClientConfig, error) {
	// TODO this whole function with settable env keys
	return nil, nil
}

// TODO try to get all the "valid list of x" strings out. Can't make them constants though? (and this func string)
func validateClientConfig(config ClientConfig, populateDefaults bool) error {
	var err error

	if populateDefaults {
		config, err = SetDefaultValuesClientConfig(config)
		if err != nil {
			return fmt.Errorf("failed to populate default values: %v", err)
		}
	}

	// region Interfaces
	// TODO adjust these strings to have links to documentation
	if config.Integration == nil {
		return errors.New("no api integration supplied, please see documentation")
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
	// TODO migrate these strings
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
	if *config.ExportLogs {
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
	if config.MaxRetryAttempts == nil {
		return errors.New("Max retry cannot be empty")
	} else if *config.MaxRetryAttempts < 0 {
		return errors.New("Max retry cannot be less than 0")
	}

	// Dynamic Rate Limiting
	// bool

	// Max Concurrent Requests
	if config.MaxConcurrentRequests == nil {
		return errors.New("Max retry cannot be empty")
	} else if *config.MaxConcurrentRequests < 0 {
		return errors.New("Max retry cannot be less than 0")
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
	if config.FollowRedirects != nil && *config.FollowRedirects {
		if DefaultMaxRedirects < 1 {
			return errors.New("max redirects cannot be less than 1")
		}
	}

	// endregion

	return nil
}

func SetDefaultValuesClientConfig(config ClientConfig) (ClientConfig, error) {

	if config.LogLevel == "" {
		config.LogLevel = DefaultLogLevelString
	}

	if config.LogOutputFormat == "" {
		config.LogOutputFormat = DefaultLogOutputFormatString
	}

	if config.LogConsoleSeparator == "" {
		config.LogConsoleSeparator = DefaultLogConsoleSeparator
	}

	if &config.ExportLogs == nil {
		defaultVal := DefaultExportLogs
		config.ExportLogs = &defaultVal
	}

	if config.LogExportPath == "" {
		config.LogExportPath = DefaultLogExportPath
	}

	if config.HideSensitiveData == nil {
		defaultVal := DefaultHideSensitiveData
		config.HideSensitiveData = &defaultVal
	}

	if config.MaxRetryAttempts == nil {
		defaultVal := DefaultMaxRetryAttempts
		config.MaxRetryAttempts = &defaultVal
	}

	if config.MaxConcurrentRequests == nil {
		defaultVal := DefaultMaxConcurrentRequests
		config.MaxRetryAttempts = &defaultVal
	}

	if config.EnableDynamicRateLimiting == nil {
		defaultVal := DefaultEnableDynamicRateLimiting
		config.EnableDynamicRateLimiting = &defaultVal
	}

	if config.CustomTimeout == nil {
		defaultVal := DefaultCustomTimeout
		config.CustomTimeout = &defaultVal
	}

	if config.TokenRefreshBufferPeriod == nil {
		defaultVal := DefaultTokenRefreshBufferPeriod
		config.TokenRefreshBufferPeriod = &defaultVal
	}

	if config.TotalRetryDuration == nil {
		defaultVal := DefaultTotalRetryDuration
		config.TotalRetryDuration = &defaultVal
	}

	if config.FollowRedirects == nil {
		defaultVal := DefaultFollowRedirects
		config.FollowRedirects = &defaultVal
	}

	if config.MaxRedirects == nil {
		defaultVal := DefaultMaxRedirects
		config.MaxRedirects = &defaultVal
	}

	return config, nil
}
