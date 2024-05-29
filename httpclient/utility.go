package httpclient

import (
	"fmt"
	"path/filepath"
	"strings"
)

func validateFilePath(path string) (string, error) {
	cleanPath := filepath.Clean(path)

	// Resolve the cleanPath to an absolute path to ensure it resolves any symbolic links
	absPath, err := filepath.EvalSymlinks(cleanPath)
	if err != nil {
		return "", fmt.Errorf("unable to resolve the absolute path of the configuration file: %s, error: %w", path, err)
	}

	// Check for suspicious patterns in the resolved path
	if strings.Contains(absPath, "..") {
		return "", fmt.Errorf("invalid path, path traversal patterns detected: %s", path)
	}

	// Ensure the file has the correct extension
	if filepath.Ext(absPath) != ConfigFileExtension {
		return "", fmt.Errorf("invalid file extension for configuration file: %s, expected .json", path)
	}

}
