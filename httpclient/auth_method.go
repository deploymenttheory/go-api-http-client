// authenticationhandler/httpclient_auth_method.go

/* The authenticationhandler package is dedicated to managing authentication
for HTTP clients, with support for multiple authentication strategies,
including OAuth and Bearer Token mechanisms. It encapsulates the logic for
determining the appropriate authentication method based on provided credentials,
validating those credentials, and managing authentication tokens. This package
aims to provide a flexible and extendable framework for handling authentication
in a secure and efficient manner, ensuring that HTTP clients can seamlessly
authenticate against various services with minimal configuration. */

package httpclient

import (
	"errors"

	"github.com/deploymenttheory/go-api-http-client/authenticationhandler"
)

// DetermineAuthMethod determines the authentication method based on the provided credentials.
// It prefers strong authentication methods (e.g., OAuth) over weaker ones (e.g., bearer tokens).
// It logs an error and returns "unknown" if no valid credentials are provided.
func DetermineAuthMethod(authConfig AuthConfig) (string, error) {
	// Initialize validation flags as true
	validClientID, validClientSecret, validUsername, validPassword := true, true, true, true
	clientIDErrMsg, clientSecretErrMsg, usernameErrMsg, passwordErrMsg := "", "", "", ""

	// Validate ClientID and ClientSecret for OAuth if provided
	if authConfig.ClientID != "" || authConfig.ClientSecret != "" {
		validClientID, clientIDErrMsg = authenticationhandler.IsValidClientID(authConfig.ClientID)
		validClientSecret, clientSecretErrMsg = authenticationhandler.IsValidClientSecret(authConfig.ClientSecret)
		// If both ClientID and ClientSecret are valid, use OAuth
		if validClientID && validClientSecret {
			return "oauth2", nil
		}
	}

	// Validate Username and Password for Bearer if OAuth is not valid or not provided
	if authConfig.Username != "" || authConfig.Password != "" {
		validUsername, usernameErrMsg = authenticationhandler.IsValidUsername(authConfig.Username)
		validPassword, passwordErrMsg = authenticationhandler.IsValidPassword(authConfig.Password)
		// If both Username and Password are valid, use Bearer
		if validUsername && validPassword {
			return "basicauth", nil
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
