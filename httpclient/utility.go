package httpclient

import (
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

// TODO all func comments in here

const ConfigFileExtension = ".json"

func validateFilePath(path string) (string, error) {
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

func validateValidClientID(clientID string) error {
	uuidRegex := `^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`
	if regexp.MustCompile(uuidRegex).MatchString(clientID) {
		return nil
	}
	return errors.New("clientID failed regex check")
}

func validateClientSecret(clientSecret string) error {
	if len(clientSecret) < 16 {
		return errors.New("client secret must be at least 16 characters long.")
	}

	if matched, _ := regexp.MatchString(`[a-z]`, clientSecret); !matched {
		return errors.New("client secret must contain at least one lowercase letter.")
	}

	if matched, _ := regexp.MatchString(`[A-Z]`, clientSecret); !matched {
		return errors.New("client secret must contain at least one uppercase letter.")
	}

	if matched, _ := regexp.MatchString(`\d`, clientSecret); !matched {
		return errors.New("client secret must contain at least one digit.")
	}

	return nil
}

func validateUsername(username string) error {
	usernameRegex := `^[a-zA-Z0-9!@#$%^&*()_\-\+=\[\]{\}\\|;:'",<.>/?]+$`
	if !regexp.MustCompile(usernameRegex).MatchString(username) {
		return errors.New("username failed regex test")
	}
	return nil
}

func validatePassword(password string) error {
	if len(password) < 8 {
		return errors.New("password not long enough")
	}
	return nil
}
