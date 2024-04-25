package httpclient

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfigFromFile(t *testing.T) {
	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "config-*.json")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name()) // Clean up the file after test

	// Write updated JSON configuration to the temp file
	configJSON := `{
		"Auth": {
			"ClientID": "787xxxxd-98bb-xxxx-8d17-xxx0f8cbfb7b",
			"ClientSecret": "xxxxxxxxxxxxx"
		},
		"Environment": {
			"InstanceName": "lbgsandbox",
			"OverrideBaseDomain": "",
			"APIType": "jamfpro"
		},
		"ClientOptions": {
			"LogLevel": "LogLevelDebug",
			"LogOutputFormat": "console",
			"LogConsoleSeparator": "  ",
			"HideSensitiveData": true,
			"EnableDynamicRateLimiting": true,
			"MaxRetryAttempts": 5,
			"MaxConcurrentRequests": 3,
			"EnableCookieJar": true,
			"FollowRedirects": true,
			"MaxRedirects": 5
		}
	}`
	_, err = tmpFile.WriteString(configJSON)
	assert.NoError(t, err)
	assert.NoError(t, tmpFile.Close())

	// Test loading from the temp file
	config, err := LoadConfigFromFile(tmpFile.Name())
	assert.NoError(t, err)
	assert.Equal(t, "787xxxxd-98bb-xxxx-8d17-xxx0f8cbfb7b", config.Auth.ClientID)
	assert.Equal(t, "xxxxxxxxxxxxx", config.Auth.ClientSecret)
	assert.Equal(t, "lbgsandbox", config.Environment.InstanceName)
	assert.Equal(t, "jamfpro", config.Environment.APIType)
	assert.Equal(t, "LogLevelDebug", config.ClientOptions.Logging.LogLevel)
	assert.Equal(t, "console", config.ClientOptions.Logging.LogOutputFormat)
	assert.Equal(t, "  ", config.ClientOptions.Logging.LogConsoleSeparator)
	assert.True(t, config.ClientOptions.Logging.HideSensitiveData)
	assert.True(t, config.ClientOptions.Retry.EnableDynamicRateLimiting)
	assert.Equal(t, 5, config.ClientOptions.Retry.MaxRetryAttempts)
	assert.Equal(t, 3, config.ClientOptions.Concurrency.MaxConcurrentRequests)
	assert.True(t, config.ClientOptions.Cookies.EnableCookieJar)
	assert.True(t, config.ClientOptions.Redirect.FollowRedirects)
	assert.Equal(t, 5, config.ClientOptions.Redirect.MaxRedirects)
}

func TestGetEnvOrDefault(t *testing.T) {
	const envKey = "TEST_ENV_VAR"
	defer os.Unsetenv(envKey)

	// Scenario 1: Environment variable is set
	expectedValue := "test_value"
	os.Setenv(envKey, expectedValue)
	assert.Equal(t, expectedValue, getEnvOrDefault(envKey, "default_value"))

	// Scenario 2: Environment variable is not set
	assert.Equal(t, "default_value", getEnvOrDefault("NON_EXISTENT_ENV_VAR", "default_value"))
}

func TestParseBool(t *testing.T) {
	assert.True(t, parseBool("true"))
	assert.False(t, parseBool("false"))
	assert.False(t, parseBool("invalid_value"))
}

func TestParseInt(t *testing.T) {
	assert.Equal(t, 42, parseInt("42", 10))
	assert.Equal(t, 10, parseInt("invalid_value", 10))
}

func TestParseDuration(t *testing.T) {
	assert.Equal(t, 5*time.Minute, parseDuration("5m", 1*time.Minute))
	assert.Equal(t, 1*time.Minute, parseDuration("invalid_value", 1*time.Minute))
}

func TestSetLoggerDefaultValues(t *testing.T) {
	config := &ClientConfig{ClientOptions: ClientOptions{}}
	setLoggerDefaultValues(config)
	assert.Equal(t, ",", config.ClientOptions.Logging.LogConsoleSeparator)
}
