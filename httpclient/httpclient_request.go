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
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// executeHTTPRequest sends an HTTP request using the client's HTTP client. It logs the request and error details, if any, using structured logging with zap fields.
//
// Parameters:
// - req: The *http.Request object that contains all the details of the HTTP request to be sent.
// - log: An instance of a logger (conforming to the logger.Logger interface) used for logging the request details and any errors.
// - method: The HTTP method used for the request, used for logging.
// - endpoint: The API endpoint the request is being sent to, used for logging.
//
// Returns:
// - *http.Response: The HTTP response from the server.
// - error: An error object if an error occurred while sending the request or nil if no error occurred.
//
// Usage:
// This function should be used whenever the client needs to send an HTTP request. It abstracts away the common logic of request execution and error handling, providing detailed logs for debugging and monitoring.
func (c *Client) executeHTTPRequest(req *http.Request, log logger.Logger, method, endpoint string) (*http.Response, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		// Log the error with structured logging, including method, endpoint, and the error itself
		log.Error("Failed to send request",
			zap.String("method", method),
			zap.String("endpoint", endpoint),
			zap.Error(err),
		)
		return nil, err
	}

	// Log the response status code for successful requests
	log.Info("Request sent successfully",
		zap.String("method", method),
		zap.String("endpoint", endpoint),
		zap.Int("status_code", resp.StatusCode),
	)

	return resp, nil
}

// handleErrorResponse processes and logs errors from an HTTP response, allowing for a customizable error message.
//
// Parameters:
// - resp: The *http.Response received from the server.
// - log: An instance of a logger (conforming to the logger.Logger interface) for logging the error details.
// - errorMessage: A custom error message that provides context about the error.
// - method: The HTTP method used for the request, for logging purposes.
// - endpoint: The endpoint the request was sent to, for logging purposes.
//
// Returns:
// - An error object parsed from the HTTP response, indicating the nature of the failure.
func (c *Client) handleErrorResponse(resp *http.Response, log logger.Logger, errorMessage, method, endpoint string) error {
	apiErr := errors.HandleAPIError(resp, log)

	// Log the provided error message along with method, endpoint, and status code.
	log.Error(errorMessage,
		zap.String("method", method),
		zap.String("endpoint", endpoint),
		zap.Int("status_code", resp.StatusCode),
		zap.String("error", apiErr.Error()),
	)

	return apiErr
}

// handleSuccessResponse unmarshals a successful HTTP response into the provided output parameter and logs the
// success details. It's designed for use when the response indicates success (status code within 200-299).
// The function logs the request's success and, in case of unmarshalling errors, logs the failure and returns the error.
//
// Parameters:
// - resp: The *http.Response received from the server.
// - out: A pointer to the variable where the unmarshalled response will be stored.
// - log: An instance of a logger (conforming to the logger.Logger interface) for logging success or unmarshalling errors.
// - method: The HTTP method used for the request, for logging purposes.
// - endpoint: The endpoint the request was sent to, for logging purposes.
//
// Returns:
// - nil if the response was successfully unmarshalled into the 'out' parameter, or an error if unmarshalling failed.
func (c *Client) handleSuccessResponse(resp *http.Response, out interface{}, log logger.Logger, method, endpoint string) error {
	if err := c.APIHandler.UnmarshalResponse(resp, out, log); err != nil {
		log.Error("Failed to unmarshal HTTP response",
			zap.String("method", method),
			zap.String("endpoint", endpoint),
			zap.Error(err),
		)
		return err
	}
	log.Info("HTTP request succeeded",
		zap.String("method", method),
		zap.String("endpoint", endpoint),
		zap.Int("status_code", resp.StatusCode),
	)
	return nil
}

// DoRequest constructs and executes a standard HTTP request with support for retry logic.
// It is intended for operations that can be encoded in a single JSON or XML body such as
// creating or updating resources. This method includes token validation, concurrency control,
// performance metrics, dynamic header setting, and structured error handling.
//
// Parameters:
// - method: The HTTP method to use (e.g., GET, POST, PUT, DELETE, PATCH).
// - endpoint: The API endpoint to which the request will be sent.
// - body: The payload to send in the request, which will be marshaled based on the API handler rules.
// - out: A pointer to a variable where the unmarshaled response will be stored.
//
// Returns:
// - A pointer to the http.Response received from the server.
// - An error if the request could not be sent, the response could not be processed, or if retry attempts fail.
//
// The function starts by validating the client's authentication token and managing concurrency using
// a token system. It then determines the appropriate API handler for marshaling the request body and
// setting headers. The request is sent to the constructed URL with all necessary headers including
// authorization, content type, and user agent.
//
// If configured for debug logging, the function logs all request headers before sending. The function then
// enters a loop to handle retryable HTTP methods, implementing a retry mechanism for transient errors,
// rate limits, and other retryable conditions based on response status codes.
//
// The function also updates performance metrics to track total request count and cumulative response time.
// After processing the response, it handles any API errors and unmarshals the response body into the provided
// 'out' parameter if the response is successful.
//
// Note:
// The function assumes that retryable HTTP methods have been properly defined in the retryableHTTPMethods map.
// It is the caller's responsibility to close the response body when the request is successful to avoid resource leaks.
func (c *Client) DoRequest(method, endpoint string, body, out interface{}, log logger.Logger) (*http.Response, error) {
	// Auth Token validation check
	valid, err := c.ValidAuthTokenCheck(log)
	if err != nil || !valid {
		return nil, fmt.Errorf("validity of the authentication token failed with error: %w", err)
	}

	// Acquire a token for concurrency management
	ctx, err := c.AcquireConcurrencyToken(context.Background(), log)
	if err != nil {
		return nil, err
	}
	defer func() {
		// Extract the requestID from the context and release the concurrency token
		if requestID, ok := ctx.Value(requestIDKey{}).(uuid.UUID); ok {
			c.ConcurrencyMgr.Release(requestID)
		}
	}()

	// Determine which set of encoding and content-type request rules to use
	apiHandler := c.APIHandler

	// Marshal Request with correct encoding
	requestData, err := apiHandler.MarshalRequest(body, method, endpoint, log)
	if err != nil {
		return nil, err
	}

	// Construct URL using the ConstructAPIResourceEndpoint function
	url := apiHandler.ConstructAPIResourceEndpoint(c.InstanceName, endpoint, log)

	// Initialize total request counter
	c.PerfMetrics.lock.Lock()
	c.PerfMetrics.TotalRequests++
	c.PerfMetrics.lock.Unlock()

	// Perform Request
	req, err := http.NewRequest(method, url, bytes.NewBuffer(requestData))
	if err != nil {
		return nil, err
	}

	// Initialize HeaderManager with the request, logger, APIHandler, and token from the Client
	headerManager := NewHeaderManager(req, log, c.APIHandler, c.Token)

	// Set and log the HTTP request headers using the HeaderManager
	headerManager.SetRequestHeaders(endpoint)
	headerManager.LogHeaders(c)

	if IsIdempotentHTTPMethod(method) {
		//if retryableHTTPMethods[method] {
		// Define a deadline for total retries based on http client TotalRetryDuration config
		totalRetryDeadline := time.Now().Add(c.clientConfig.ClientOptions.TotalRetryDuration)
		i := 0
		for {
			// Check if we've reached the maximum number of retries or if our total retry time has exceeded
			if i > c.clientConfig.ClientOptions.MaxRetryAttempts || time.Now().After(totalRetryDeadline) {
				return nil, fmt.Errorf("max retry attempts reached or total retry duration exceeded")
			}

			// This context is used to propagate cancellations and timeouts for the request.
			// For example, if a request's context gets canceled or times out, the request will be terminated early.
			req = req.WithContext(ctx)

			// Start response time measurement
			responseTimeStart := time.Now()

			// Execute the request
			resp, err := c.executeHTTPRequest(req, log, method, endpoint)
			if err != nil {
				return nil, err
			}

			// After each request, compute and update response time
			responseDuration := time.Since(responseTimeStart)
			c.updatePerformanceMetrics(responseDuration)

			// Checks for the presence of a deprecation header in the HTTP response and logs if found.
			if i == 0 {
				CheckDeprecationHeader(resp, log)
			}

			// Handle (unmarshal) response with API Handler
			if err := apiHandler.UnmarshalResponse(resp, out, log); err != nil {
				// Use type assertion to check if the error is of type *errors.APIError
				if apiErr, ok := err.(*errors.APIError); ok {
					// Log the API error with structured logging for specific APIError handling
					log.Error("Received an API error",
						zap.String("method", method),
						zap.String("endpoint", endpoint),
						zap.Int("status_code", apiErr.StatusCode),
						zap.String("message", apiErr.Message),
					)
					return resp, apiErr // Return the typed error for further handling if needed
				} else {
					// Log other errors with structured logging for general error handling
					log.Error("Failed to unmarshal HTTP response",
						zap.String("method", method),
						zap.String("endpoint", endpoint),
						zap.Error(err), // Use zap.Error to log the error object
					)
					return resp, err
				}
			}

			// Successful response
			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				log.Info("HTTP request succeeded",
					zap.String("method", method),
					zap.String("endpoint", endpoint),
					zap.Int("status_code", resp.StatusCode),
				)
				return resp, nil
			} else if
			// Resource not found
			resp.StatusCode == http.StatusNotFound {
				log.Warn("Resource not found",
					zap.String("method", method),
					zap.String("endpoint", endpoint),
					zap.Int("status_code", resp.StatusCode),
				)
				// Use a centralized method for handling not found error
				return resp, err
			}

			// Retry Logic
			// Non-retryable error
			if errors.IsNonRetryableError(resp) {
				log.Warn("Encountered a non-retryable error",
					zap.String("method", method),
					zap.String("endpoint", endpoint),
					zap.Int("status_code", resp.StatusCode),
					zap.String("description", errors.TranslateStatusCode(resp.StatusCode)),
				)
				return resp, errors.HandleAPIError(resp, log) // Assume this method logs the error internally
			} else if errors.IsRateLimitError(resp) {
				waitDuration := parseRateLimitHeaders(resp) // Parses headers to determine wait duration
				log.Warn("Encountered a rate limit error. Retrying after wait duration.",
					zap.String("method", method),
					zap.String("endpoint", endpoint),
					zap.Int("status_code", resp.StatusCode),
					zap.Duration("wait_duration", waitDuration),
				)
				time.Sleep(waitDuration)
				i++
				continue // This will restart the loop, effectively "retrying" the request
			} else if errors.IsTransientError(resp) {
				waitDuration := calculateBackoff(i) // Calculates backoff duration
				log.Warn("Encountered a transient error. Retrying after backoff.",
					zap.String("method", method),
					zap.String("endpoint", endpoint),
					zap.Int("status_code", resp.StatusCode),
					zap.Duration("wait_duration", waitDuration),
					zap.Int("attempt", i),
				)
				time.Sleep(waitDuration)
				i++
				continue // This will restart the loop, effectively "retrying" the request
			} else {
				log.Error("Received unexpected error status from HTTP request",
					zap.String("method", method),
					zap.String("endpoint", endpoint),
					zap.Int("status_code", resp.StatusCode),
					zap.String("description", errors.TranslateStatusCode(resp.StatusCode)),
				)
				return resp, errors.HandleAPIError(resp, log)
			}
		}
	} else if IsNonIdempotentHTTPMethod(method) {
		// Start response time measurement
		responseTimeStart := time.Now()
		// For non-retryable HTTP Methods (POST - Create)
		req = req.WithContext(ctx)

		// Execute the request
		resp, err := c.executeHTTPRequest(req, log, method, endpoint)
		if err != nil {
			return nil, err
		}

		// After each request, compute and update response time
		responseDuration := time.Since(responseTimeStart)
		c.updatePerformanceMetrics(responseDuration)

		// Checks for the presence of a deprecation header in the HTTP response and logs if found.
		CheckDeprecationHeader(resp, log)

		// Handle (unmarshal) response with API Handler
		if err := apiHandler.UnmarshalResponse(resp, out, log); err != nil {
			// Use type assertion to check if the error is of type *errors.APIError
			if apiErr, ok := err.(*errors.APIError); ok {
				// Log the API error with structured logging for specific APIError handling
				log.Error("Received an API error",
					zap.String("method", method),
					zap.String("endpoint", endpoint),
					zap.Int("status_code", apiErr.StatusCode),
					zap.String("message", apiErr.Message),
				)
				return resp, apiErr // Return the typed error for further handling if needed
			} else {
				// Log other errors with structured logging for general error handling
				log.Error("Failed to unmarshal HTTP response",
					zap.String("method", method),
					zap.String("endpoint", endpoint),
					zap.Error(err), // Use zap.Error to log the error object
				)
				return resp, err // Return the original error
			}
		}

		// Check if the response status code is within the success range
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			// Success, no need for logging
			return resp, nil
		} else {
			// Translate the status code to a human-readable description
			statusDescription := errors.TranslateStatusCode(resp.StatusCode)
			if apiErr, ok := err.(*errors.APIError); ok {
				// Log the API error with structured logging for specific APIError handling
				log.Error("Received an API error",
					zap.String("method", method),
					zap.String("endpoint", endpoint),
					zap.Int("status_code", apiErr.StatusCode),
					zap.String("message", apiErr.Message),
				)
			}

			// Return an error with the status code and its description
			return resp, fmt.Errorf("error status code: %d - %s", resp.StatusCode, statusDescription)
		}
	}
	// TODO refactor to remove repition and to streamline error handling.
	return nil, fmt.Errorf("an unexpected error occurred")
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

	// Get Request Headers dynamically based on api handler
	acceptHeader := apiHandler.GetAcceptHeader()

	// Set Request Headers
	c.SetRequestHeaders(req, contentType, acceptHeader, log)

	// Execute the request
	resp, err := c.executeHTTPRequest(req, log, method, endpoint)
	if err != nil {
		return nil, err
	}

	// Check for successful status code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// Handle error responses
		return nil, c.handleErrorResponse(resp, log, "Failed to process the HTTP request", method, endpoint)
	} else {
		// Handle successful responses
		return resp, c.handleSuccessResponse(resp, out, log, method, endpoint)
	}
	// TODO refactor to remove dependancy on func (c *Client) SetRequestHeaders
}
