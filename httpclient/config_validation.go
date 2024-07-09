// httpclient/client_configuration.go
// Description: This file contains functions to load and validate configuration values from a JSON file or environment variables.
package httpclient

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	DefaultLogLevelString              = "LogLevelInfo"
	DefaultLogOutputFormatString       = "pretty"
	DefaultLogConsoleSeparator         = "	"
	DefaultLogExportPath               = "/defaultlogs"
	DefaultMaxRetryAttempts            = 3
	DefaultMaxConcurrentRequests       = 1
	DefaultExportLogs                  = false
	DefaultHideSensitiveData           = false
	DefaultEnableDynamicRateLimiting   = false
	DefaultCustomTimeout               = 5 * time.Second
	DefaultTokenRefreshBufferPeriod    = 2 * time.Minute
	DefaultTotalRetryDuration          = 5 * time.Minute
	DefaultFollowRedirects             = false
	DefaultMaxRedirects                = 5
	DefaultEnableConcurrencyManagement = false
)

// LoadConfigFromFile loads http client configuration settings from a JSON file.
func LoadConfigFromFile(filepath string) (*ClientConfig, error) {
	absPath, err := validateFilePath(filepath)
	if err != nil {
		return nil, fmt.Errorf("invalid file path: %v", err)
	}

	file, err := os.Open(absPath)
	if err != nil {
		return nil, fmt.Errorf("could not open file: %v", err)
	}
	defer file.Close()

	byteValue, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("could not read file: %v", err)
	}

	var config ClientConfig
	err = json.Unmarshal(byteValue, &config)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal JSON: %v", err)
	}

	config.SetDefaultValuesClientConfig()

	return &config, nil
}

// LoadConfigFromEnv loads HTTP client configuration settings from environment variables.
// If any environment variables are not set, the default values defined in the constants are used instead.
func LoadConfigFromEnv() (*ClientConfig, error) {
	config := &ClientConfig{
		HideSensitiveData:           getEnvAsBool("HIDE_SENSITIVE_DATA", DefaultHideSensitiveData),
		MaxRetryAttempts:            getEnvAsInt("MAX_RETRY_ATTEMPTS", DefaultMaxRetryAttempts),
		MaxConcurrentRequests:       getEnvAsInt("MAX_CONCURRENT_REQUESTS", DefaultMaxConcurrentRequests),
		EnableDynamicRateLimiting:   getEnvAsBool("ENABLE_DYNAMIC_RATE_LIMITING", DefaultEnableDynamicRateLimiting),
		CustomTimeout:               getEnvAsDuration("CUSTOM_TIMEOUT", DefaultCustomTimeout),
		TokenRefreshBufferPeriod:    getEnvAsDuration("TOKEN_REFRESH_BUFFER_PERIOD", DefaultTokenRefreshBufferPeriod),
		TotalRetryDuration:          getEnvAsDuration("TOTAL_RETRY_DURATION", DefaultTotalRetryDuration),
		FollowRedirects:             getEnvAsBool("FOLLOW_REDIRECTS", DefaultFollowRedirects),
		MaxRedirects:                getEnvAsInt("MAX_REDIRECTS", DefaultMaxRedirects),
		EnableConcurrencyManagement: getEnvAsBool("ENABLE_CONCURRENCY_MANAGEMENT", DefaultEnableConcurrencyManagement),
	}

	// Load custom cookies from environment variables.
	customCookies := getEnvAsString("CUSTOM_COOKIES", "")
	if customCookies != "" {
		cookies := []*http.Cookie{}
		for _, cookie := range strings.Split(customCookies, ";") {
			parts := strings.SplitN(cookie, "=", 2)
			if len(parts) == 2 {
				cookies = append(cookies, &http.Cookie{
					Name:  parts[0],
					Value: parts[1],
				})
			} else {
				return nil, fmt.Errorf("invalid env cookie configuration: %s", parts)
			}
		}
		config.CustomCookies = cookies
	}

	return config, nil
}

// TODO Review validateClientConfig
func (c ClientConfig) validateClientConfig() error {

	if c.PopulateDefaultValues {
		c.SetDefaultValuesClientConfig()
	}

	// TODO adjust these strings to have links to documentation & centralise them
	if c.Integration == nil {
		return errors.New("no http client api integration supplied, please see repo documentation for this client and go-api-http-client-integration and provide an implementation")
	}

	if c.EnableConcurrencyManagement {
		if c.MaxConcurrentRequests < 1 {
			return errors.New("maximum concurrent requests cannot be less than 1")
		}
	}

	if c.CustomTimeout.Seconds() < 0 {
		return errors.New("timeout cannot be less than 0 seconds")
	}

	if c.TokenRefreshBufferPeriod.Seconds() < 0 {
		return errors.New("refresh buffer period cannot be less than 0 seconds")
	}

	if c.RetryEligiableRequests {
		if c.TotalRetryDuration.Seconds() < 0 {
			return errors.New("total retry duration cannot be less than 0 seconds")
		}

		if c.MaxRetryAttempts < 0 {
			return errors.New("max retry cannot be less than 0")
		}

	}

	if c.FollowRedirects {
		if DefaultMaxRedirects < 1 {
			return errors.New("max redirects cannot be less than 1")
		}
	}

	return nil
}

// SetDefaultValuesClientConfig sets default values for the client configuration. Ensuring that all fields have a valid or minimum value.
func (c *ClientConfig) SetDefaultValuesClientConfig() {
	setDefaultBool(&c.HideSensitiveData, DefaultHideSensitiveData)
	setDefaultInt(&c.MaxRetryAttempts, DefaultMaxRetryAttempts, 1)
	setDefaultInt(&c.MaxConcurrentRequests, DefaultMaxConcurrentRequests, 1)
	setDefaultBool(&c.EnableDynamicRateLimiting, DefaultEnableDynamicRateLimiting)
	setDefaultDuration(&c.CustomTimeout, DefaultCustomTimeout)
	setDefaultDuration(&c.TokenRefreshBufferPeriod, DefaultTokenRefreshBufferPeriod)
	setDefaultDuration(&c.TotalRetryDuration, DefaultTotalRetryDuration)
	setDefaultBool(&c.FollowRedirects, DefaultFollowRedirects)
	setDefaultInt(&c.MaxRedirects, DefaultMaxRedirects, 0)
	setDefaultBool(&c.EnableConcurrencyManagement, DefaultEnableConcurrencyManagement)
}
