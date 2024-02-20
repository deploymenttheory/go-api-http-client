// http_request.go
package httpclient

import (
	"bytes"
	"context"
	"net/http"
	"time"

	"github.com/deploymenttheory/go-api-http-client/errors"
	"github.com/deploymenttheory/go-api-http-client/logger"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// DoRequest constructs and executes an HTTP request based on the provided method, endpoint, request body, and output variable.
// This function serves as a dispatcher, deciding whether to execute the request with or without retry logic based on the
// idempotency of the HTTP method. Idempotent methods (GET, PUT, DELETE) are executed with retries to handle transient errors
// and rate limits, while non-idempotent methods (POST, PATCH) are executed without retries to avoid potential side effects
// of duplicating non-idempotent operations.

// Parameters:
// - method: A string representing the HTTP method to be used for the request. This method determines the execution path
//   and whether the request will be retried in case of failures.
// - endpoint: The target API endpoint for the request. This should be a relative path that will be appended to the base URL
//   configured for the HTTP client.
// - body: The payload for the request, which will be serialized into the request body. The serialization format (e.g., JSON, XML)
//   is determined by the content-type header and the specific implementation of the API handler used by the client.
// - out: A pointer to an output variable where the response will be deserialized. The function expects this to be a pointer to
//   a struct that matches the expected response schema.
// - log: An instance of a logger implementing the logger.Logger interface, used to log informational messages, warnings, and
//   errors encountered during the execution of the request.

// Returns:
// - *http.Response: The HTTP response received from the server. In case of successful execution, this response contains
//   the status code, headers, and body of the response. In case of errors, particularly after exhausting retries for
//   idempotent methods, this response may contain the last received HTTP response that led to the failure.
// - error: An error object indicating failure during request execution. This could be due to network issues, server errors,
//   or a failure in request serialization/deserialization. For idempotent methods, an error is returned if all retries are
//   exhausted without success.

// Usage:
// This function is the primary entry point for executing HTTP requests using the client. It abstracts away the details of
// request retries, serialization, and response handling, providing a simplified interface for making HTTP requests. It is
// suitable for a wide range of HTTP operations, from fetching data with GET requests to submitting data with POST requests.

// Example:
// var result MyResponseType
// resp, err := client.DoRequest("GET", "/api/resource", nil, &result, logger)
// if err != nil {
//     // Handle error
// }
// // Use `result` or `resp` as needed

// Note:
// - The caller is responsible for closing the response body when not nil to avoid resource leaks.
// - The function ensures concurrency control by managing concurrency tokens internally, providing safe concurrent operations
//   within the client's concurrency model.
// - The decision to retry requests is based on the idempotency of the HTTP method and the client's retry configuration,
//   including maximum retry attempts and total retry duration.

func (c *Client) DoRequest(method, endpoint string, body, out interface{}, log logger.Logger) (*http.Response, error) {
	if IsIdempotentHTTPMethod(method) {
		return c.executeRequestWithRetries(method, endpoint, body, out, log)
	} else if IsNonIdempotentHTTPMethod(method) {
		return c.executeRequest(method, endpoint, body, out, log)
	} else {
		return nil, log.Error("HTTP method not supported", zap.String("method", method))
	}
}

// executeRequestWithRetries executes an HTTP request using the specified method, endpoint, request body, and output variable.
// It is designed for idempotent HTTP methods (GET, PUT, DELETE), where the request can be safely retried in case of
// transient errors or rate limiting. The function implements a retry mechanism that respects the client's configuration
// for maximum retry attempts and total retry duration. Each retry attempt uses exponential backoff with jitter to avoid
// thundering herd problems.
//
// Parameters:
// - method: The HTTP method to be used for the request (e.g., "GET", "PUT", "DELETE").
// - endpoint: The API endpoint to which the request will be sent. This should be a relative path that will be appended
// to the base URL of the HTTP client.
// - body: The request payload, which will be marshaled into the request body based on the content type. Can be nil for
// methods that do not send a payload.
// - out: A pointer to the variable where the unmarshaled response will be stored. The function expects this to be a
// pointer to a struct that matches the expected response schema.
// - log: An instance of a logger (conforming to the logger.Logger interface) used for logging the request, retry
// attempts, and any errors encountered.
//
// Returns:
// - *http.Response: The HTTP response from the server, which may be the response from a successful request or the last
// failed attempt if all retries are exhausted.
//   - error: An error object if an error occurred during the request execution or if all retry attempts failed. The error
//     may be a structured API error parsed from the response or a generic error indicating the failure reason.
//
// Usage:
// This function should be used for operations that are safe to retry and where the client can tolerate the additional
// latency introduced by the retry mechanism. It is particularly useful for handling transient errors and rate limiting
// responses from the server.
//
// Note:
// - The caller is responsible for closing the response body to prevent resource leaks.
// - The function respects the client's concurrency token, acquiring and releasing it as needed to ensure safe concurrent
// operations.
// - The retry mechanism employs exponential backoff with jitter to mitigate the impact of retries on the server.
func (c *Client) executeRequestWithRetries(method, endpoint string, body, out interface{}, log logger.Logger) (*http.Response, error) {
	// Include the core logic for handling non-idempotent requests with retries here.
	log.Debug("Executing request with retries", zap.String("method", method), zap.String("endpoint", endpoint))

	// Auth Token validation check
	valid, err := c.ValidAuthTokenCheck(log)
	if err != nil || !valid {
		return nil, err
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

	// Marshal Request with correct encoding defined in api handler
	requestData, err := c.APIHandler.MarshalRequest(body, method, endpoint, log)
	if err != nil {
		return nil, err
	}

	// Construct URL with correct structure defined in api handler
	url := c.APIHandler.ConstructAPIResourceEndpoint(c.InstanceName, endpoint, log)

	// Initialize total request counter
	c.PerfMetrics.lock.Lock()
	c.PerfMetrics.TotalRequests++
	c.PerfMetrics.lock.Unlock()

	// Perform Request
	req, err := http.NewRequest(method, url, bytes.NewBuffer(requestData))
	if err != nil {
		return nil, err
	}

	// Set request headers
	headerManager := NewHeaderManager(req, log, c.APIHandler, c.Token)
	headerManager.SetRequestHeaders(endpoint)
	headerManager.LogHeaders(c)

	// Define a retry deadline based on the client's total retry duration configuration
	totalRetryDeadline := time.Now().Add(c.clientConfig.ClientOptions.TotalRetryDuration)

	var resp *http.Response
	retryCount := 0
	for time.Now().Before(totalRetryDeadline) { // Check if the current time is before the total retry deadline
		req = req.WithContext(ctx)
		resp, err = c.executeHTTPRequest(req, log, method, endpoint)
		// Check for successful status code
		if err == nil && resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return resp, c.handleSuccessResponse(resp, out, log, method, endpoint)
		}

		// Leverage TranslateStatusCode for more descriptive error logging
		statusMessage := errors.TranslateStatusCode(resp)

		// Check for non-retryable errors
		if resp != nil && errors.IsNonRetryableStatusCode(resp) {
			log.Warn("Non-retryable error received", zap.Int("status_code", resp.StatusCode), zap.String("status_message", statusMessage))
			return resp, errors.HandleAPIError(resp, log)
		}

		// Check for retryable errors
		if errors.IsRateLimitError(resp) || errors.IsTransientError(resp) {
			retryCount++
			if retryCount > c.clientConfig.ClientOptions.MaxRetryAttempts {
				log.Warn("Max retry attempts reached", zap.String("method", method), zap.String("endpoint", endpoint))
				break
			}
			waitDuration := calculateBackoff(retryCount)
			log.Warn("Retrying request due to error", zap.String("method", method), zap.String("endpoint", endpoint), zap.Int("retryCount", retryCount), zap.Duration("waitDuration", waitDuration), zap.Error(err), zap.String("status_message", statusMessage))
			time.Sleep(waitDuration)
			continue
		}

		// Handle error responses
		if err != nil || !errors.IsRetryableStatusCode(resp.StatusCode) {
			if apiErr := errors.HandleAPIError(resp, log); apiErr != nil {
				err = apiErr
			}
			log.Error("API error", zap.String("status_message", statusMessage), zap.Error(err))
			break
		}
	}
	// Handles final non-API error.
	if err != nil {
		return nil, err
	}

	return resp, errors.HandleAPIError(resp, log)
}

// executeRequest executes an HTTP request using the specified method, endpoint, and request body without implementing
// retry logic. It is primarily designed for non idempotent HTTP methods like POST and PATCH, where the request should
// not be automatically retried within this function due to the potential side effects of re-submitting the same data.
//
// Parameters:
// - method: The HTTP method to be used for the request, typically "POST" or "PATCH".
// - endpoint: The API endpoint to which the request will be sent. This should be a relative path that will be appended
// to the base URL of the HTTP client.
//   - body: The request payload, which will be marshaled into the request body based on the content type. This can be any
//     data structure that can be marshaled into the expected request format (e.g., JSON, XML).
//   - out: A pointer to the variable where the unmarshaled response will be stored. This should be a pointer to a struct
//
// that matches the expected response schema.
// - log: An instance of a logger (conforming to the logger.Logger interface) used for logging the request and any errors
// encountered.
//
// Returns:
// - *http.Response: The HTTP response from the server. This includes the status code, headers, and body of the response.
// - error: An error object if an error occurred during the request execution. This could be due to network issues,
// server errors, or issues with marshaling/unmarshaling the request/response.
//
// Usage:
// This function is suitable for operations where the request should not be retried automatically, such as data submission
// operations where retrying could result in duplicate data processing. It ensures that the request is executed exactly
// once and provides detailed logging for debugging purposes.
//
// Note:
// - The caller is responsible for closing the response body to prevent resource leaks.
// - The function ensures concurrency control by acquiring and releasing a concurrency token before and after the request
// execution.
// - The function logs detailed information about the request execution, including the method, endpoint, status code, and
// any errors encountered.
func (c *Client) executeRequest(method, endpoint string, body, out interface{}, log logger.Logger) (*http.Response, error) {
	// Include the core logic for handling idempotent requests here.
	log.Debug("Executing request without retries", zap.String("method", method), zap.String("endpoint", endpoint))

	// Auth Token validation check
	valid, err := c.ValidAuthTokenCheck(log)
	if err != nil || !valid {
		return nil, err
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
	url := c.APIHandler.ConstructAPIResourceEndpoint(c.InstanceName, endpoint, log)

	// Perform Request
	req, err := http.NewRequest(method, url, bytes.NewBuffer(requestData))
	if err != nil {
		return nil, err
	}

	// Set request headers
	headerManager := NewHeaderManager(req, log, c.APIHandler, c.Token)
	headerManager.SetRequestHeaders(endpoint)
	headerManager.LogHeaders(c)

	req = req.WithContext(ctx)

	// Execute the HTTP request
	resp, err := c.executeHTTPRequest(req, log, method, endpoint)
	if err != nil {
		return nil, err
	}

	// Checks for the presence of a deprecation header in the HTTP response and logs if found.
	CheckDeprecationHeader(resp, log)

	// Check for successful status code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// Handle error responses
		return nil, c.handleErrorResponse(resp, log, "Failed to process the HTTP request", method, endpoint)
	} else {
		// Handle successful responses
		return resp, c.handleSuccessResponse(resp, out, log, method, endpoint)
	}
}

// executeHTTPRequest sends an HTTP request using the client's HTTP client. It logs the request and error details, if any,
// using structured logging with zap fields.
//
// Parameters:
// - req: The *http.Request object that contains all the details of the HTTP request to be sent.
// - log: An instance of a logger (conforming to the logger.Logger interface) used for logging the request details and any
// errors.
// - method: The HTTP method used for the request, used for logging.
// - endpoint: The API endpoint the request is being sent to, used for logging.
//
// Returns:
// - *http.Response: The HTTP response from the server.
// - error: An error object if an error occurred while sending the request or nil if no error occurred.
//
// Usage:
// This function should be used whenever the client needs to send an HTTP request. It abstracts away the common logic of
// request execution and error handling, providing detailed logs for debugging and monitoring.
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
	if err := c.APIHandler.HandleResponse(resp, out, log); err != nil {
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
		return nil, err
	}

	// Determine which set of encoding and content-type request rules to use
	//apiHandler := c.APIHandler

	// Marshal the multipart form data
	requestData, contentType, err := c.APIHandler.MarshalMultipartRequest(fields, files, log)
	if err != nil {
		return nil, err
	}

	// Construct URL using the ConstructAPIResourceEndpoint function
	url := c.APIHandler.ConstructAPIResourceEndpoint(c.InstanceName, endpoint, log)

	// Create the request
	req, err := http.NewRequest(method, url, bytes.NewBuffer(requestData))
	if err != nil {
		return nil, err
	}

	// Initialize HeaderManager
	headerManager := NewHeaderManager(req, log, c.APIHandler, c.Token)

	// Use HeaderManager to set headers
	headerManager.SetContentType(contentType)
	headerManager.SetRequestHeaders(endpoint)
	headerManager.LogHeaders(c)

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
}
