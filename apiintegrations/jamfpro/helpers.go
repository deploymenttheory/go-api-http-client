package jamfpro

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// TODO this duplicated across both integrations. Improve it another time.

// ParseISO8601Date attempts to parse a string date in ISO 8601 format.
func ParseISO8601_Date(dateStr string) (time.Time, error) {
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
