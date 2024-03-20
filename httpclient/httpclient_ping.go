// httpclient_ping.go
package httpclient

import (
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// DoPing performs an HTTP "ping" to the specified endpoint using the given HTTP method, body,
// and output variable. It attempts the request until a 200 OK response is received or the
// maximum number of retry attempts is reached. The function uses a backoff strategy for retries
// to manage load on the server and network. This function is useful for checking the availability
// or health of an endpoint, particularly in environments where network reliability might be an issue.

// Parameters:
// - method: The HTTP method to be used for the request. This should typically be "GET" for a ping operation, but the function is designed to accommodate any HTTP method.
// - endpoint: The target API endpoint for the ping request. This should be a relative path that will be appended to the base URL configured for the HTTP client.
// - body: The payload for the request, if any. For a typical ping operation, this would be nil, but the function is designed to accommodate requests that might require a body.
// - out: A pointer to an output variable where the response will be deserialized. This is included to maintain consistency with the signature of other request functions, but for a ping operation, it might not be used.

// Returns:
// - *http.Response: The HTTP response from the server. In case of a successful ping (200 OK),
// this response contains the status code, headers, and body of the response. In case of errors
// or if the maximum number of retries is reached without a successful response, this will be the
// last received HTTP response.
//
// - error: An error object indicating failure during the execution of the ping operation. This
// could be due to network issues, server errors, or reaching the maximum number of retry attempts
// without receiving a 200 OK response.

// Usage:
// This function is intended for use in scenarios where it's necessary to confirm the availability
// or health of an endpoint, with built-in retry logic to handle transient network or server issues.
// The caller is responsible for handling the response and error according to their needs, including
// closing the response body when applicable to avoid resource leaks.

// Example:
// var result MyResponseType
// resp, err := client.DoPing("GET", "/api/health", nil, &result)
//
//	if err != nil {
//	    // Handle error
//	}
//
// // Process response
func (c *Client) DoPing(method, endpoint string, body, out interface{}) (*http.Response, error) {
	log := c.Logger
	log.Debug("Starting Ping", zap.String("method", method), zap.String("endpoint", endpoint))

	// Initialize retry count and define maximum retries
	var retryCount int
	maxRetries := c.clientConfig.ClientOptions.MaxRetryAttempts

	// Loop until a successful response is received or maximum retries are reached
	for retryCount <= maxRetries {
		// Use the existing 'do' function for sending the request
		resp, err := c.executeRequestWithRetries(method, endpoint, body, out)

		// If request is successful and returns 200 status code, return the response
		if err == nil && resp.StatusCode == http.StatusOK {
			log.Debug("Ping successful", zap.String("method", method), zap.String("endpoint", endpoint))
			return resp, nil
		}

		// Increment retry count and log the attempt
		retryCount++
		log.Warn("Ping failed, retrying...", zap.String("method", method), zap.String("endpoint", endpoint), zap.Int("retryCount", retryCount))

		// Calculate backoff duration and wait before retrying
		backoffDuration := calculateBackoff(retryCount)
		time.Sleep(backoffDuration)
	}

	// If maximum retries are reached without a successful response, return an error
	log.Error("Ping failed after maximum retries", zap.String("method", method), zap.String("endpoint", endpoint))
	return nil, fmt.Errorf("ping failed after %d retries", maxRetries)
}
