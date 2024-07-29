// httpclient/utility.go
package httpclient

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// TODO all func comments in here

const ConfigFileExtension = ".json"

// validateFilePath checks if a file path is valid.
func validateConfigFilePath(path string) (string, error) {
	cleanPath := filepath.Clean(path)

	absPath, err := filepath.EvalSymlinks(cleanPath)
	if err != nil {
		return "", fmt.Errorf("unable to resolve the absolute path of the configuration file: %s, error: %w", path, err)
	}

	if strings.Contains(absPath, "..") {
		return "", fmt.Errorf("invalid path, path traversal patterns detected: %s", path)
	}

	if filepath.Ext(absPath) != ConfigFileExtension {
		return "", fmt.Errorf("invalid file extension for configuration file: %s, expected .json", path)
	}

	return path, nil

}

// getEnvAsString reads an environment variable as a string, with a fallback default value.
func getEnvAsString(name string, defaultVal string) string {
	if value, exists := os.LookupEnv(name); exists {
		return value
	}
	return defaultVal
}

// getEnvAsBool reads an environment variable as a boolean, with a fallback default value.
func getEnvAsBool(name string, defaultVal bool) bool {
	if value, exists := os.LookupEnv(name); exists {
		boolValue, err := strconv.ParseBool(value)
		if err == nil {
			return boolValue
		}
	}
	return defaultVal
}

// getEnvAsInt reads an environment variable as an integer, with a fallback default value.
func getEnvAsInt(name string, defaultVal int) int {
	if value, exists := os.LookupEnv(name); exists {
		intValue, err := strconv.Atoi(value)
		if err == nil {
			return intValue
		}
	}
	return defaultVal
}

// getEnvAsDuration reads an environment variable as a duration, with a fallback default value.
func getEnvAsDuration(name string, defaultVal time.Duration) time.Duration {
	if value, exists := os.LookupEnv(name); exists {
		durationValue, err := time.ParseDuration(value)
		if err == nil {
			return durationValue
		}
	}
	return defaultVal
}

// http field validation functions

// setDefaultBool sets a boolean field to a default value if it is not already set during http client config field validation.
func setDefaultBool(field *bool, defaultValue bool) {
	if !*field {
		*field = defaultValue
	}
}

// setDefaultInt sets an integer field to a default value if it is not already set during http client config field validation.
func setDefaultInt(field *int, defaultValue, minValue int) {
	if *field == 0 {
		*field = defaultValue
	} else if *field < minValue {
		*field = minValue
	}
}

// setDefaultDuration sets a duration field to a default value if it is not already set during http client config field validation.
func setDefaultDuration(field *time.Duration, defaultValue time.Duration) {
	if *field == 0 {
		*field = defaultValue
	} else if *field < 0 {
		*field = defaultValue
	}
}

// CheckDeprecationHeader checks the response headers for the Deprecation header and logs a warning if present.
func (c *Client) CheckDeprecationHeader(resp *http.Response) {
	deprecationHeader := resp.Header.Get("Deprecation")
	if deprecationHeader != "" {
		c.Sugar.Warn("API endpoint is deprecated", deprecationHeader, resp.Request.URL.String())
	}
}
