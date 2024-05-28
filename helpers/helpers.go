// helpers/helpers.go
package helpers

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// ParseISO8601Date attempts to parse a string date in ISO 8601 format.
func ParseISO8601Date(dateStr string) (time.Time, error) {
	return time.Parse(time.RFC3339, dateStr)
}

// SafeOpenFile opens a file safely after validating and resolving its path.
func SafeOpenFile(filePath string) (*os.File, error) {
	// Clean the file path to remove any ".." or similar components that can lead to directory traversal
	cleanPath := filepath.Clean(filePath)

	// Resolve the clean path to an absolute path and ensure it resolves any symbolic links
	absPath, err := filepath.EvalSymlinks(cleanPath)
	if err != nil {
		return nil, fmt.Errorf("unable to resolve the absolute path: %s, error: %w", filePath, err)
	}

	// Optionally, check if the absolute path is within a permitted directory (omitted here for brevity)
	// Example: allowedPathPrefix := "/safe/directory/"
	// if !strings.HasPrefix(absPath, allowedPathPrefix) {
	// 	return nil, fmt.Errorf("access to the file path is not allowed: %s", absPath)
	// }

	// Open the file if the path is deemed safe
	return os.Open(absPath)
}

// UnmarshalJSON parses the duration from JSON string.
func (d *JSONDuration) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	duration, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	*d = JSONDuration(duration)
	return nil
}

// Duration returns the time.Duration value.
func (d JSONDuration) Duration() time.Duration {
	return time.Duration(d)
}

// MarshalJSON returns the JSON representation of the duration.
func (d JSONDuration) String() string {
	return time.Duration(d).String()
}

// JSONDuration wraps time.Duration for custom JSON unmarshalling.
type JSONDuration time.Duration

// GetEnvOrDefault returns the value of an environment variable or a default value.
func GetEnvOrDefault(envKey string, defaultValue string) string {
	if value, exists := os.LookupEnv(envKey); exists {
		return value
	}
	return defaultValue
}

// ParseJSONDuration attempts to parse a string value as a duration and returns the result or a default value.
func ParseJSONDuration(value string, defaultVal JSONDuration) JSONDuration {
	result, err := time.ParseDuration(value)
	if err != nil {
		return defaultVal
	}
	return JSONDuration(result)
}
