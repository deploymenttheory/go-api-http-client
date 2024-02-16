package httpclient

import (
	"errors"
	"regexp"
)

// IsValidClientID checks if the provided client ID is a valid UUID.
// Returns true if valid, along with an empty error message; otherwise, returns false with an error message.
func IsValidClientID(clientID string) (bool, string) {
	uuidRegex := `^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`
	if regexp.MustCompile(uuidRegex).MatchString(clientID) {
		return true, ""
	}
	return false, "Client ID is not a valid UUID format."
}

// IsValidClientSecret checks if the provided client secret meets your application's validation criteria.
// Returns true if valid, along with an empty error message; otherwise, returns false with an error message.
func IsValidClientSecret(clientSecret string) (bool, string) {
	if len(clientSecret) < 16 {
		return false, "Client secret must be at least 16 characters long."
	}

	// Check if the client secret contains at least one lowercase letter, one uppercase letter, and one digit.
	// The requirement for a special character has been removed.
	complexityRegex := `(?=.*[a-z])(?=.*[A-Z])(?=.*\d)`
	if matched, _ := regexp.MatchString(complexityRegex, clientSecret); !matched {
		return false, "Client secret must contain at least one lowercase letter, one uppercase letter, and one digit."
	}
	return true, ""
}

// IsValidUsername checks if the provided username meets your application's validation criteria.
// Returns true if valid, along with an empty error message; otherwise, returns false with an error message.
func IsValidUsername(username string) (bool, string) {
	if regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString(username) {
		return true, ""
	}
	return false, "Username must contain only alphanumeric characters."
}

// IsValidPassword checks if the provided password meets your application's validation criteria.
// Returns true if valid, along with an empty error message; otherwise, returns false with an error message.
func IsValidPassword(password string) (bool, string) {
	if len(password) >= 8 {
		return true, ""
	}
	return false, "Password must be at least 8 characters long."
}

// DetermineAuthMethod determines the authentication method based on the provided credentials.
// It prefers strong authentication methods (e.g., OAuth) over weaker ones (e.g., bearer tokens).
// It logs an error and returns "unknown" if no valid credentials are provided.
func DetermineAuthMethod(authConfig AuthConfig) (string, error) {
	validClientID, clientIDErrMsg := IsValidClientID(authConfig.ClientID)
	validClientSecret, clientSecretErrMsg := IsValidClientSecret(authConfig.ClientSecret)
	validUsername, usernameErrMsg := IsValidUsername(authConfig.Username)
	validPassword, passwordErrMsg := IsValidPassword(authConfig.Password)

	if validClientID && validClientSecret {
		return "oauth", nil
	}

	if validUsername && validPassword {
		return "bearer", nil
	}

	// Construct error message
	errorMsg := "No valid credentials provided."
	if !validClientID {
		errorMsg += " " + clientIDErrMsg
	}
	if !validClientSecret {
		errorMsg += " " + clientSecretErrMsg
	}
	if !validUsername {
		errorMsg += " " + usernameErrMsg
	}
	if !validPassword {
		errorMsg += " " + passwordErrMsg
	}

	return "unknown", errors.New(errorMsg)
}
