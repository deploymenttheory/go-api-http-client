// http_request.go
package httpclient

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/deploymenttheory/go-api-http-client/errors"
	"github.com/deploymenttheory/go-api-http-client/logger"
	"go.uber.org/zap"
)

// prepareRequest constructs an HTTP request including marshaling the body and setting necessary headers.
// It retains the detailed logic for handling different content types and dynamic header setting.
func (c *Client) prepareRequest(method, endpoint string, body interface{}, log logger.Logger) (*http.Request, error) {
	// Marshal the request body based on the API handler rules
	requestData, err := c.APIHandler.MarshalRequest(body, method, endpoint, log)
	if err != nil {
		log.Error("Error marshaling request body", zap.String("method", method), zap.String("endpoint", endpoint), zap.Error(err))
		return nil, err
	}

	// Construct the full URL for the request
	url := c.APIHandler.ConstructAPIResourceEndpoint(c.InstanceName, endpoint, log)

	// Initialize the HTTP request
	req, err := http.NewRequest(method, url, bytes.NewBuffer(requestData))
	if err != nil {
		log.Error("Error creating HTTP request", zap.String("method", method), zap.String("url", url), zap.Error(err))
		return nil, err
	}

	// Set dynamic headers based on API handler logic
	contentType := c.APIHandler.GetContentTypeHeader(endpoint, log)
	acceptHeader := c.APIHandler.GetAcceptHeader()

	req.Header.Add("Authorization", "Bearer "+c.Token)
	req.Header.Add("Content-Type", contentType)
	req.Header.Add("Accept", acceptHeader)
	req.Header.Set("User-Agent", GetUserAgentHeader())

	// Redact sensitive data if configured to do so
	if c.clientConfig.ClientOptions.HideSensitiveData {
		req.Header.Set("Authorization", "REDACTED")
	}

	return req, nil
}

// processResponse handles the HTTP response, including error checking, unmarshaling the response body, and logging.
// It maintains the structured error handling and response processing as per the original implementation.
func (c *Client) processResponse(resp *http.Response, out interface{}, log logger.Logger) error {
	defer resp.Body.Close()

	// Check for deprecation headers and log if present
	CheckDeprecationHeader(resp, log)

	// Handle API errors and unmarshal response body if successful
	if err := c.APIHandler.UnmarshalResponse(resp, out, log); err != nil {
		if apiErr, ok := err.(*errors.APIError); ok {
			// Log the structured API error
			log.Error("Received an API error", zap.Int("status_code", apiErr.StatusCode), zap.String("message", apiErr.Message))
			return apiErr // Return the structured API error
		}
		// Log the error encountered during unmarshaling
		log.Error("Failed to unmarshal HTTP response", zap.Error(err))
		return err // Return the original error
	}

	// Log successful response
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		log.Info("HTTP request succeeded", zap.Int("status_code", resp.StatusCode))
		return nil
	}

	// Handle non-success status codes
	statusDescription := errors.TranslateStatusCode(resp.StatusCode)
	errorMessage := fmt.Sprintf("HTTP request failed with status code %d - %s", resp.StatusCode, statusDescription)
	// Log the error message before returning it
	log.Error(errorMessage, zap.Int("status_code", resp.StatusCode), zap.String("description", statusDescription))
	return fmt.Errorf(errorMessage) // Construct and return the error with the logged message
}

// executeRequestWithRetry handles the sending of an HTTP request with retry logic for transient errors and rate limits.
// It preserves the detailed retry mechanism including logging and backoff calculations.
func (c *Client) executeRequestWithRetry(req *http.Request, ctx context.Context, log logger.Logger) (*http.Response, error) {
	var lastErr error

	for i := 0; ; i++ {
		resp, err := c.httpClient.Do(req.WithContext(ctx))

		if err != nil {
			// Network error or other request failure
			log.Error("HTTP request failed", zap.Error(err))
			lastErr = err
			if ctx.Err() != nil {
				// Context cancellation or deadline exceeded
				return nil, ctx.Err()
			}
		} else {
			defer resp.Body.Close() // Ensure response body is closed

			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				// Successful response
				return resp, nil
			}

			// Log non-successful HTTP status
			log.Error("HTTP request failed", zap.Int("status_code", resp.StatusCode), zap.String("status_text", http.StatusText(resp.StatusCode)))
			lastErr = errors.HandleAPIError(resp, log) // Use the structured error handling

			if !errors.IsTransientError(resp) {
				// Non-retryable error, break the loop
				break
			}

			if errors.IsRateLimitError(resp) {
				// Rate limit error detected, parse headers to get wait duration
				waitDuration := parseRateLimitHeaders(resp)
				log.Info("Encountered rate limit error, waiting before retrying", zap.Duration("waitDuration", waitDuration), zap.Int("attempt", i+1))
				time.Sleep(waitDuration) // Wait according to the rate limit before retrying
				continue                 // Proceed to the next iteration for retry
			}
		}

		// Retry logic for transient errors
		if i >= c.clientConfig.ClientOptions.MaxRetryAttempts {
			// Maximum retry attempts reached
			log.Error("Max retry attempts reached, giving up", zap.Error(lastErr))
			break
		}

		backoffDuration := calculateBackoff(i) // Calculate backoff duration
		log.Info("Retrying HTTP request due to transient error", zap.Duration("backoff", backoffDuration), zap.Int("attempt", i+1))
		time.Sleep(backoffDuration) // Wait before retrying
	}

	return nil, lastErr // Return the last encountered error
}

// DoRequest constructs and executes a standard HTTP request with support for retry logic.
// This refactored version uses helper functions to handle request preparation, response processing, and retry logic.
func (c *Client) DoRequest(method, endpoint string, body, out interface{}, log logger.Logger) (*http.Response, error) {
	// Validate the authentication token
	valid, err := c.ValidAuthTokenCheck(log)
	if err != nil || !valid {
		log.Error("Authentication token validation failed", zap.Error(err))
		return nil, err
	}

	// Acquire a concurrency token with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	requestID, err := c.ConcurrencyMgr.Acquire(ctx)
	if err != nil {
		log.Error("Failed to acquire concurrency token", zap.Error(err))
		return nil, err
	}
	defer c.ConcurrencyMgr.Release(requestID)

	// Prepare the HTTP request
	req, err := c.prepareRequest(method, endpoint, body, log)
	if err != nil {
		errMsg := "Failed to prepare HTTP request"
		log.Error(errMsg, zap.Error(err))
		return nil, fmt.Errorf("%s: %w", errMsg, err)
	}

	// Execute the request with retry logic
	resp, err := c.executeRequestWithRetry(req, ctx, log)
	if err != nil {
		errMsg := "Failed to execute HTTP request"
		log.Error(errMsg, zap.Error(err))
		return nil, fmt.Errorf("%s: %w", errMsg, err)
	}

	// Process the HTTP response
	if err := c.processResponse(resp, out, log); err != nil {
		return resp, err // processResponse already logs and formats the error appropriately
	}

	return resp, nil
}

// DoMultipartRequest creates and executes a multipart HTTP request. It is used for sending files
// and form fields in a single request. This method handles the construction of the multipart
// message body, setting the appropriate headers, and sending the request to the given endpoint.
//
// Parameters:
// - method: The HTTP method to use (e.g., POST, PUT).
// - endpoint: The API endpoint to which the request will be sent.
// - fields: A map of form fields and their values to include in the multipart message.
// - files: A map of file field names to file paths that will be included as file attachments.
// - out: A pointer to a variable where the unmarshaled response will be stored.
//
// Returns:
// - A pointer to the http.Response received from the server.
// - An error if the request could not be sent or the response could not be processed.
//
// The function first validates the authentication token, then constructs the multipart
// request body based on the provided fields and files. It then constructs the full URL for
// the request, sets the required headers (including Authorization and Content-Type), and
// sends the request.
//
// If debug mode is enabled, the function logs all the request headers before sending the request.
// After the request is sent, the function checks the response status code. If the response is
// not within the success range (200-299), it logs an error and returns the response and an error.
// If the response is successful, it attempts to unmarshal the response body into the 'out' parameter.
//
// Note:
// The caller should handle closing the response body when successful.
func (c *Client) DoMultipartRequest(method, endpoint string, fields map[string]string, files map[string]string, out interface{}, log logger.Logger) (*http.Response, error) {
	// Auth Token validation check
	valid, err := c.ValidAuthTokenCheck(log)
	if err != nil || !valid {
		return nil, fmt.Errorf("validity of the authentication token failed with error: %w", err)
	}

	// Determine which set of encoding and content-type request rules to use
	apiHandler := c.APIHandler

	// Marshal the multipart form data
	requestData, contentType, err := apiHandler.MarshalMultipartRequest(fields, files, log)
	if err != nil {
		return nil, err
	}

	// Construct URL using the ConstructAPIResourceEndpoint function
	url := apiHandler.ConstructAPIResourceEndpoint(c.InstanceName, endpoint, log)

	// Create the request
	req, err := http.NewRequest(method, url, bytes.NewBuffer(requestData))
	if err != nil {
		return nil, err
	}

	// Set Request Headers
	req.Header.Add("Authorization", "Bearer "+c.Token)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("User-Agent", GetUserAgentHeader())

	// Debug: Print request headers
	redactedAuthorization := RedactSensitiveData(c, "Authorization", req.Header.Get("Authorization"))
	c.Logger.Debug("HTTP Request Headers",
		zap.String("Authorization", redactedAuthorization),
		zap.String("Content-Type", req.Header.Get("Content-Type")),
		zap.String("Accept", req.Header.Get("Accept")),
		zap.String("User-Agent", req.Header.Get("User-Agent")),
	)

	// Execute the request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.Error("Failed to send multipart request",
			zap.String("method", method),
			zap.String("endpoint", endpoint),
			zap.Int("status_code", resp.StatusCode),
			zap.String("status_text", http.StatusText(resp.StatusCode)),
		)
		return nil, err
	}

	// Check for successful status code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// Use HandleAPIError to process the error response and log it accordingly
		apiErr := errors.HandleAPIError(resp, log)

		// Log additional context about the request that led to the error
		log.Error("Received non-success status code from multipart request",
			zap.String("method", method),
			zap.String("endpoint", endpoint),
			zap.Int("status_code", resp.StatusCode),
			zap.String("status_text", http.StatusText(resp.StatusCode)),
		)

		// Return the original HTTP response and the API error
		return resp, apiErr
	}

	// Unmarshal the response
	if err := apiHandler.UnmarshalResponse(resp, out, log); err != nil {
		log.Error("Failed to unmarshal HTTP response",
			zap.String("method", method),
			zap.String("endpoint", endpoint),
			zap.String("error", err.Error()),
		)
		return resp, err
	}

	return resp, nil
}
