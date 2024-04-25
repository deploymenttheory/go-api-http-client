// authenticationhandler/authenticationhandler.go

package authenticationhandler

import (
	"sync"
	"time"

	"github.com/deploymenttheory/go-api-http-client/logger"
)

// AuthTokenHandler manages authentication tokens.
type AuthTokenHandler struct {
	Credentials       ClientCredentials // Credentials holds the authentication credentials.
	Token             string            // Token holds the current authentication token.
	Expires           time.Time         // Expires indicates the expiry time of the current authentication token.
	Logger            logger.Logger     // Logger provides structured logging capabilities for logging information, warnings, and errors.
	AuthMethod        string            // AuthMethod specifies the method of authentication, e.g., "bearer" or "oauth".
	InstanceName      string            // InstanceName represents the name of the instance or environment the client is interacting with.
	tokenLock         sync.Mutex        // tokenLock ensures thread-safe access to the token and its expiry to prevent concurrent write/read issues.
	HideSensitiveData bool
}

// ClientCredentials holds the credentials necessary for authentication.
type ClientCredentials struct {
	Username     string
	Password     string
	ClientID     string
	ClientSecret string
}

// TokenResponse represents the structure of a token response from the API.
type TokenResponse struct {
	Token   string    `json:"token"`
	Expires time.Time `json:"expires"`
}

// NewAuthTokenHandler creates a new instance of AuthTokenHandler.
func NewAuthTokenHandler(logger logger.Logger, authMethod string, credentials ClientCredentials, instanceName string, hideSensitiveData bool) *AuthTokenHandler {
	return &AuthTokenHandler{
		Logger:            logger,
		AuthMethod:        authMethod,
		Credentials:       credentials,
		InstanceName:      instanceName,
		HideSensitiveData: hideSensitiveData,
	}
}
