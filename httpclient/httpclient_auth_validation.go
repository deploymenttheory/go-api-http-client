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

	// Check for at least one lowercase letter
	if matched, _ := regexp.MatchString(`[a-z]`, clientSecret); !matched {
		return false, "Client secret must contain at least one lowercase letter."
	}

	// Check for at least one uppercase letter
	if matched, _ := regexp.MatchString(`[A-Z]`, clientSecret); !matched {
		return false, "Client secret must contain at least one uppercase letter."
	}

	// Check for at least one digit
	if matched, _ := regexp.MatchString(`\d`, clientSecret); !matched {
		return false, "Client secret must contain at least one digit."
	}

	return true, ""
}

// IsValidUsername checks if the provided username meets password safe validation criteria.
// Returns true if valid, along with an empty error message; otherwise, returns false with an error message.
func IsValidUsername(username string) (bool, string) {
	// Extended regex to include a common set of password safe special characters
	usernameRegex := `^[a-zA-Z0-9!@#$%^&*()_\-\+=\[\]{\}\\|;:'",<.>/?]+$`
	if regexp.MustCompile(usernameRegex).MatchString(username) {
		return true, ""
	}
	return false, "Username must contain only alphanumeric characters and password safe special characters (!@#$%^&*()_-+=[{]}\\|;:'\",<.>/?)."
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
	// Initialize validation flags as true
	validClientID, validClientSecret, validUsername, validPassword := true, true, true, true
	clientIDErrMsg, clientSecretErrMsg, usernameErrMsg, passwordErrMsg := "", "", "", ""

	// Validate ClientID and ClientSecret for OAuth if provided
	if authConfig.ClientID != "" || authConfig.ClientSecret != "" {
		validClientID, clientIDErrMsg = IsValidClientID(authConfig.ClientID)
		validClientSecret, clientSecretErrMsg = IsValidClientSecret(authConfig.ClientSecret)
		// If both ClientID and ClientSecret are valid, use OAuth
		if validClientID && validClientSecret {
			return "oauth", nil
		}
	}

	// Validate Username and Password for Bearer if OAuth is not valid or not provided
	if authConfig.Username != "" || authConfig.Password != "" {
		validUsername, usernameErrMsg = IsValidUsername(authConfig.Username)
		validPassword, passwordErrMsg = IsValidPassword(authConfig.Password)
		// If both Username and Password are valid, use Bearer
		if validUsername && validPassword {
			return "bearer", nil
		}
	}

	// Construct an error message if any of the provided fields are invalid
	errorMsg := "No valid credentials provided."
	if !validClientID && authConfig.ClientID != "" {
		errorMsg += " " + clientIDErrMsg
	}
	if !validClientSecret && authConfig.ClientSecret != "" {
		errorMsg += " " + clientSecretErrMsg
	}
	if !validUsername && authConfig.Username != "" {
		errorMsg += " " + usernameErrMsg
	}
	if !validPassword && authConfig.Password != "" {
		errorMsg += " " + passwordErrMsg
	}

	return "unknown", errors.New(errorMsg)
}
