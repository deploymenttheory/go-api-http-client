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
func GetAuthMethod(config ClientConfig) (string, error) {
	var (
		validClientID      bool
		validClientSecret  bool
		validUsername      bool
		validPassword      bool
		errMsgClientID     string
		errMsgClientSecret string
		errMsgUsername     string
		errMesgPassword    string
	)

	if config.clientID != "" || config.clientSecret != "" {
		validClientID, errMsgClientID = authenticationhandler.IsValidClientID(config.clientID)
		validClientSecret, errMsgClientSecret = authenticationhandler.IsValidClientSecret(config.clientSecret)

		if validClientID && validClientSecret {
			return "oauth2", nil
		}
	}

	if config.basicAuthUsername != "" || config.basicAuthPassword != "" {
		validUsername, errMsgUsername = authenticationhandler.IsValidUsername(config.basicAuthUsername)
		validPassword, errMesgPassword = authenticationhandler.IsValidPassword(config.basicAuthPassword)

		if validUsername && validPassword {
			return "basicauth", nil
		}
	}

	errorMsg := "No valid credentials provided."
	if !validClientID && config.clientID != "" {
		errorMsg += " " + errMsgClientID
	}
	if !validClientSecret && config.clientSecret != "" {
		errorMsg += " " + errMsgClientSecret
	}
	if !validUsername && config.basicAuthUsername != "" {
		errorMsg += " " + errMsgUsername
	}
	if !validPassword && config.basicAuthPassword != "" {
		errorMsg += " " + errMesgPassword
	}

	return "unknown", errors.New(errorMsg)
}
