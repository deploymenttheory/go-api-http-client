// authenticationhandler/tokenmanager.go
package httpclient

import (
	"fmt"
	"time"

	"go.uber.org/zap"
)

// CheckAndRefreshAuthToken checks the token's validity and refreshes it if necessary.
// It returns true if the token is valid post any required operations and false with an error otherwise.
func (c *Client) CheckAndRefreshAuthToken() (bool, error) {
	const maxConsecutiveRefreshAttempts = 10
	refreshAttempts := 0

	if c.isTokenValid() {
		c.Logger.Info("Authentication token is valid", zap.Bool("IsTokenValid", true))
		return true, nil
	}

	for !c.isTokenValid() {
		c.Logger.Debug("Token found to be invalid or close to expiry, handling token acquisition or refresh.")
		if err := c.obtainNewToken(); err != nil {
			c.Logger.Error("Failed to obtain new token", zap.Error(err))
			return false, err
		}

		refreshAttempts++
		if refreshAttempts >= maxConsecutiveRefreshAttempts {
			return false, fmt.Errorf(
				"exceeded maximum consecutive token refresh attempts (%d): access token lifetime (%s) is likely too short compared to the buffer period (%s) configured for token refresh",
				maxConsecutiveRefreshAttempts,
				c.AuthTokenExpiry.Sub(time.Now()).String(), // Access token lifetime
				c.config.TokenRefreshBufferPeriod.String(), // Configured buffer period
			)
		}
	}

	isValid := c.isTokenValid()
	c.Logger.Info("Authentication token status check completed", zap.Bool("IsTokenValid", isValid))
	return isValid, nil
}

// isTokenValid checks if the current token is non-empty and not about to expire.
// It considers a token valid if it exists and the time until its expiration is greater than the provided buffer period.
func (c *Client) isTokenValid() bool {
	isValid := c.AuthToken != "" && time.Until(c.AuthTokenExpiry) >= c.config.TokenRefreshBufferPeriod
	c.Logger.Debug("Checking token validity", zap.Bool("IsValid", isValid), zap.Duration("TimeUntilExpiry", time.Until(c.AuthTokenExpiry)))
	return isValid
}

// obtainNewToken acquires a new token using the credentials provided.
// It handles different authentication methods based on the AuthMethod setting.
func (c *Client) obtainNewToken() error {
	var err error
	backoff := time.Millisecond * 100

	for attempts := 0; attempts < 5; attempts++ {
		if c.config.AuthMethod == "basicauth" {
			err = c.GetBasicToken()
		} else if c.config.AuthMethod == "oauth2" {
			err = c.OAuth2TokenAcquisition()
		} else {
			err = fmt.Errorf("no valid credentials provided. Unable to obtain a token")
			c.Logger.Error("Authentication method not supported", zap.String("AuthMethod", c.config.AuthMethod))
			return err // Return the error immediately
		}

		if err == nil {
			break
		}

		c.Logger.Error("Failed to obtain new token, retrying...", zap.Error(err), zap.Int("attempt", attempts+1))
		time.Sleep(backoff)
		backoff *= 2
	}

	if err != nil {
		c.Logger.Error("Failed to obtain new token after all attempts", zap.Error(err))
		return err
	}

	return nil
}

// refreshTokenIfNeeded refreshes the token if it's close to expiration.
// This function decides on the method based on the credentials type available.
func (c *Client) refreshTokenIfNeeded() error {
	if time.Until(c.AuthTokenExpiry) < c.config.TokenRefreshBufferPeriod {
		c.Logger.Info("Token is close to expiry and will be refreshed", zap.Duration("TimeUntilExpiry", time.Until(c.AuthTokenExpiry)))
		var err error
		if c.config.BasicAuthUsername != "" && c.config.BasicAuthPassword != "" {
			err = c.GetBasicToken()
		} else if c.config.ClientID != "" && c.config.ClientSecret != "" {
			err = c.OAuth2TokenAcquisition()
		} else {
			err = fmt.Errorf("unknown auth method")
			c.Logger.Error("Failed to determine authentication method for token refresh", zap.String("AuthMethod", c.config.AuthMethod))
		}

		if err != nil {
			c.Logger.Error("Failed to refresh token", zap.Error(err))
			return err
		}
	}
	return nil
}
