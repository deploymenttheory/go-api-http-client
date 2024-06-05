// httpclient/client_configuration.go
// Description: This file contains functions to load and validate configuration values from a JSON file or environment variables.
package httpclient

import (
	"errors"
	"fmt"
	"log"
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

// TODO migrate all the loose strings

// TODO LoadConfigFromFile Func
func LoadConfigFromFile(filepath string) (*ClientConfig, error) {
	return nil, nil
}

// TODO LoadConfigFromEnv Func
func LoadConfigFromEnv() (*ClientConfig, error) {
	return nil, nil
}

// TODO Review validateClientConfig
func validateClientConfig(config ClientConfig, populateDefaults bool) error {
	var err error

	if populateDefaults {
		log.Println("FEATURE PENDING")
		SetDefaultValuesClientConfig(&config)
	}

	// TODO adjust these strings to have links to documentation & centralise them
	if config.Integration == nil {
		return errors.New("no api integration supplied, please see documentation")
	}

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

	validLogFormats := []string{
		"json",
		"pretty",
	}

	if !slices.Contains(validLogFormats, config.LogOutputFormat) {
		return fmt.Errorf("invalid log output format: %s", config.LogOutputFormat)
	}

	// Log Export Path
	if config.ExportLogs {
		_, err = validateFilePath(config.LogExportPath)
		if err != nil {
			return err
		}
	}

	if config.MaxRetryAttempts < 0 {
		return errors.New("Max retry cannot be less than 0")
	}

	if config.MaxConcurrentRequests < 1 {
		return errors.New("maximum concurrent requests cannot be less than 1")
	}

	if config.CustomTimeout.Seconds() < 0 {
		return errors.New("timeout cannot be less than 0 seconds")
	}

	if config.TokenRefreshBufferPeriod.Seconds() < 0 {
		return errors.New("refresh buffer period cannot be less than 0 seconds")
	}

	if config.TotalRetryDuration.Seconds() < 0 {
		return errors.New("total retry duration cannot be less than 0 seconds")
	}

	if config.FollowRedirects {
		if DefaultMaxRedirects < 1 {
			return errors.New("max redirects cannot be less than 1")
		}
	}

	return nil
}

func SetDefaultValuesClientConfig(config *ClientConfig) {

	if config.LogLevel == "" {
		config.LogLevel = DefaultLogLevelString
	}

	if config.LogOutputFormat == "" {
		config.LogOutputFormat = DefaultLogOutputFormatString
	}

	if config.LogConsoleSeparator == "" {
		config.LogConsoleSeparator = DefaultLogConsoleSeparator
	}

	if !config.ExportLogs {
		config.ExportLogs = DefaultExportLogs
	}

	if config.LogExportPath == "" {
		config.LogExportPath = DefaultLogExportPath
	}

	if !config.HideSensitiveData {
		config.HideSensitiveData = DefaultHideSensitiveData
	}

	if config.MaxRetryAttempts == 0 {
		config.MaxRetryAttempts = DefaultMaxRetryAttempts
	}

	if config.MaxConcurrentRequests == 0 {
		config.MaxRetryAttempts = DefaultMaxConcurrentRequests
	}

	if !config.EnableDynamicRateLimiting {
		config.EnableDynamicRateLimiting = DefaultEnableDynamicRateLimiting
	}

	if config.CustomTimeout == 0 {
		config.CustomTimeout = DefaultCustomTimeout
	}

	if config.TokenRefreshBufferPeriod == 0 {
		config.TokenRefreshBufferPeriod = DefaultTokenRefreshBufferPeriod
	}

	if config.TotalRetryDuration == 0 {
		config.TotalRetryDuration = DefaultTotalRetryDuration
	}

	if !config.FollowRedirects {
		config.FollowRedirects = DefaultFollowRedirects
	}

	if config.MaxRedirects == 0 {
		config.MaxRedirects = DefaultMaxRedirects
	}

}
