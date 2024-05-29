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

func validateClientConfig(config ClientConfig) error {
	var err error

	// region Auth

	// TODO centralise these values
	validAuthMethods := []string{"oauth2", "basic"}
	if !slices.Contains(validAuthMethods, config.AuthMethod) {
		return fmt.Errorf("auth method not valid: %s", config.AuthMethod)
	}

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

	default:
		return errors.New("you shouldn't be here")

	}

	// endregion

	// region Logging

	// Level
	validLogLevels := []string{
		"LogLeveDebug",
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

	// endregion

	return nil
}
