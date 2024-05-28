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
func GetAuthMethod(authConfig AuthConfig) (string, error) {
	var validClientID, validClientSecret, validUsername, validPassword bool
	var clientIDErrMsg, clientSecretErrMsg, usernameErrMsg, passwordErrMsg string

	if authConfig.ClientID != "" || authConfig.ClientSecret != "" {
		validClientID, clientIDErrMsg = authenticationhandler.IsValidClientID(authConfig.ClientID)
		validClientSecret, clientSecretErrMsg = authenticationhandler.IsValidClientSecret(authConfig.ClientSecret)

		if validClientID && validClientSecret {
			return "oauth2", nil
		}
	}

	if authConfig.Username != "" || authConfig.Password != "" {
		validUsername, usernameErrMsg = authenticationhandler.IsValidUsername(authConfig.Username)
		validPassword, passwordErrMsg = authenticationhandler.IsValidPassword(authConfig.Password)

		if validUsername && validPassword {
			return "basicauth", nil
		}
	}

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
