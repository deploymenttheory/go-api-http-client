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

// prepareRequest constructs an HTTP request for a given method, endpoint, and body payload. This method is a key part
// of the HTTP client's request preparation process, encapsulating several important steps:
//
//  1. Request Body Marshaling: The method parameter 'body' is marshaled into a byte slice according to the
//     specific requirements of the API being interacted with. This could involve JSON encoding, XML encoding,
//     or any other required format. The exact marshaling behavior is defined by the APIHandler implementation
//     associated with the client, allowing for flexibility and support for multiple API formats.
//
//  2. URL Construction: The full URL for the request is constructed using the base URL of the API, the specific
//     endpoint being accessed, and any necessary path or query parameters. This step ensures that requests are
//     directed to the correct resource on the API server.
//
//  3. HTTP Request Initialization: An *http.Request object is initialized with the marshaled body, the constructed
//     URL, and the specified HTTP method. This step prepares the request for sending to the API server.
//
//  4. Header Setting: The method sets necessary HTTP headers for the request, including Authorization (using a bearer token),
//     Content-Type (based on the payload format), and Accept (based on the expected response format). Additional headers,
//     like User-Agent, can also be set at this stage. The exact headers and their values are determined by the API's requirements
//     and the APIHandler implementation.
//
//  5. Sensitive Data Handling: If configured, the method redacts sensitive information (like the Authorization header) from the
//     request headers for security purposes. This step is important for logging and debugging, preventing accidental exposure
//     of sensitive credentials.
//
// Parameters:
// - method: The HTTP method to be used for the request (e.g., GET, POST, PUT, DELETE).
// - endpoint: The specific endpoint of the API to which the request will be sent.
// - body: The payload to be sent with the request. This could be an object that will be marshaled into the required format.
// - log: The logger instance to use for logging any errors encountered during request preparation.
//
// Returns:
//   - A pointer to the prepared *http.Request object, ready to be sent.
//   - An error if any issues are encountered during the request preparation process, such as marshaling errors, URL construction
//     issues, or problems initializing the HTTP request.
//
// Note:
// This method does not send the request; it only prepares it. The actual sending of the request is handled by other methods,
// potentially involving additional steps like retry logic or response handling.
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

// processResponse manages the handling of an HTTP response after an API call, utilizing the APIHandler interface to
// interpret and process the response. This method is integral to the client's response handling workflow, ensuring
// that both successful and error responses from the API are dealt with in a structured and consistent manner.
// It performs critical functions such as error checking, response body unmarshaling, and logging, guided by the
// rules and structures defined by the APIHandler in use.
//
// The processing sequence includes:
//  1. Deprecation Header Inspection: Searches for deprecation warnings in the response headers, logging any found
//     to alert developers or system administrators of potential future issues with the current API usage.
//  2. Success Response Processing: For HTTP status codes indicating success (2xx range), the method uses the APIHandler
//     to unmarshal the response body into the 'out' parameter, converting the raw response into a structured and
//     application-ready format.
//  3. Error Handling: Processes error responses (non-2xx status codes) by leveraging the APIHandler to distinguish between
//     known API errors with specific handling requirements and generic errors, ensuring appropriate logging and
//     error reporting.
//  4. Structured Logging: Employs structured logging throughout the process to clearly document significant events, such as
//     deprecation warnings, API errors, or unmarshaling issues, facilitating easier debugging and system monitoring.
//
// Parameters:
//   - resp: The *http.Response object received from the API call. This object contains the status code, headers, and body of the response.
//   - out: A pointer to a variable where the unmarshaled response body should be stored. The exact type of this variable should align
//     with the expected structure of the response body (e.g., a struct representing the JSON payload of the response).
//   - log: A logger instance for recording informational messages and errors encountered during the response processing.
//
// Returns:
//   - nil if the response is successfully processed without errors, indicating a successful API call and proper unmarshaling of the response body.
//   - An error object if any issues are encountered during the processing of the response. This could be an API-specific error (with structured
//     details) or a generic error (e.g., unmarshaling failures, unexpected status codes).
//
// Note:
//   - The method ensures that the response body is closed before returning, preventing resource leaks. This is aligned with best practices for
//     handling HTTP response bodies in Go.
//   - The method is designed to be called immediately after receiving an HTTP response, serving as a central point for handling all outcomes
//     of an API call.
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

// retryableHTTPMethods returns a map of HTTP methods that are considered suitable for retrying in case of errors.
// This typically includes idempotent HTTP methods which, when repeated, are expected to have the same effect
// on the server's state, thus making them safe for retry operations. The function identifies GET, DELETE, PUT,
// and PATCH methods as retryable, based on the assumption that these methods are implemented in an idempotent manner
// according to HTTP standards.
func retryableHTTPMethods() map[string]bool {
	return map[string]bool{
		http.MethodGet:    true,
		http.MethodDelete: true,
		http.MethodPut:    true,
		http.MethodPatch:  true,
	}
}

// executeRequest sends a single HTTP request using the client's HTTP client. This function is intended for
// executing a request without any retry logic, making it suitable for non-retryable HTTP methods or
// when a retryable method has reached its retry limit. The function injects the provided context into the request,
// allowing for timeout and cancellation control. The response and any errors encountered during the request
// execution are returned to the caller. It's important to note that the caller is responsible for closing
// the response body to avoid resource leaks. This is a standard practice in Go for managing HTTP response bodies.
//
// Parameters:
//   - req: The HTTP request to be sent. This request should be fully prepared, with the correct method, URL,
//     headers, and body as needed for the specific API call being made.
//   - ctx: The context to use for this request. This allows for controlling cancellations and timeouts.
//   - log: The logger instance to use for logging any errors encountered during the request execution.
//
// Returns:
// - A pointer to the http.Response received from the server if the request is successful.
// - An error if the request could not be sent or if the server responds with an error status code.
func (c *Client) executeRequest(req *http.Request, ctx context.Context, log logger.Logger) (*http.Response, error) {
	resp, err := c.httpClient.Do(req.WithContext(ctx))
	if err != nil {
		log.Error("HTTP request failed", zap.Error(err))
		return nil, err
	}
	return resp, nil // The caller is responsible for closing the response body.
}

// executeRequestWithRetry sends an HTTP request with a retry mechanism for handling transient errors and rate limits.
// The function first checks if the HTTP method of the request is considered retryable (GET, DELETE, PUT, PATCH).
// If the method is not retryable, it delegates the request execution to executeRequest without retrying.
// For retryable methods, the function enters a loop, attempting to send the request until it succeeds,
// reaches the maximum number of retry attempts, or encounters a non-retryable error.
//
// The retry logic includes handling for:
// - Transient errors, where the request may succeed if retried after a short delay.
// - Rate limit errors, where the request is retried after a delay specified by the server's rate limiting headers.
//
// The function uses exponential backoff with jitter for calculating the delay between retries,
// to avoid overwhelming the server and to mitigate the thundering herd problem.
//
// Parameters:
// - req: The HTTP request to be sent. The request should be fully prepared with the correct method, URL, headers, and body.
// - ctx: The context associated with the request. This context controls the request lifecycle and can be used to cancel the request or set a deadline.
// - log: The logger used to record informational messages and errors encountered during the request execution and retry process.
//
// Returns:
// - A pointer to the http.Response received from the server if the request eventually succeeds within the retry limits.
// - An error if the request fails to send after all retries, or if a non-retryable error is encountered.
//
// Note:
// - The caller is responsible for closing the response body of the returned http.Response to avoid resource leaks.
// - The function adheres to the context deadline or cancellation, terminating the retry loop if the context expires.
func (c *Client) executeRequestWithRetry(req *http.Request, ctx context.Context, log logger.Logger) (*http.Response, error) {
	if !retryableHTTPMethods()[req.Method] {
		// If the method is not retryable, execute the request once without retry logic.
		return c.executeRequest(req, ctx, log)
	}

	var lastErr error
	for i := 0; ; i++ {
		resp, err := c.executeRequest(req, ctx, log)
		if err == nil {
			// Check the response status code
			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				return resp, nil
			}

			if !errors.IsTransientError(resp) && !errors.IsRateLimitError(resp) {
				resp.Body.Close() // Ensure the response body is closed
				return resp, nil
			}

			if errors.IsRateLimitError(resp) {
				waitDuration := parseRateLimitHeaders(resp)
				log.Info("Encountered rate limit error, waiting before retrying", zap.Duration("waitDuration", waitDuration), zap.Int("attempt", i+1))
				time.Sleep(waitDuration)
				resp.Body.Close() // Ensure the response body is closed before the next attempt
				continue
			}

			resp.Body.Close() // Ensure the response body is closed before the next attempt
		} else {
			lastErr = err
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
		}

		if i >= c.clientConfig.ClientOptions.MaxRetryAttempts {
			log.Error("Max retry attempts reached, giving up", zap.Error(lastErr))
			break
		}

		backoffDuration := calculateBackoff(i)
		log.Info("Retrying HTTP request due to transient error", zap.Duration("backoff", backoffDuration), zap.Int("attempt", i+1))
		time.Sleep(backoffDuration)
	}

	return nil, lastErr
}

// DoRequest constructs and executes an HTTP request with optional retry logic based on the request method.
// This function serves as a comprehensive solution for making HTTP calls, encompassing various features:
// authentication token validation, concurrency control, dynamic header setting, structured error handling,
// and conditional retry logic. It is designed to support operations that can be encoded in a single JSON
// or XML body, such as creating or updating resources.
//
// The function workflow includes:
//  1. Authentication Token Validation: Ensures that the client's authentication token is valid before proceeding with the request.
//  2. Concurrency Control: Manages concurrency using a token system, ensuring that no more than a predefined number of requests
//     are made concurrently, to prevent overwhelming the server.
//  3. Request Preparation: Constructs the HTTP request, including marshaling the request body based on the API handler rules
//     and setting necessary headers like Authorization, Content-Type, and User-Agent.
//  4. Request Execution: Depending on the HTTP method, the request may be executed with or without retry logic. Retryable methods
//     (GET, DELETE, PUT, PATCH) are subject to retry logic in case of transient errors or rate limits, using an exponential backoff
//     strategy. Non-retryable methods are executed once without retries.
//  5. Response Processing: Handles the server response, unmarshals the response body into the provided output parameter if the
//     response is successful, and manages errors using structured error handling.
//
// Parameters:
// - method: The HTTP method to use (e.g., GET, POST, PUT, DELETE, PATCH). Determines whether retry logic should be applied.
// - endpoint: The API endpoint to which the request will be sent. Used to construct the full request URL.
// - body: The payload to send in the request, which will be marshaled according to the specified content type.
// - out: A pointer to a variable where the unmarshaled response will be stored. This is where the result of a successful request will be placed.
// - log: A logger instance for recording informational messages and errors encountered during the request execution.
//
// Returns:
// - A pointer to the http.Response received from the server if the request is successful.
// - An error if the request could not be sent, the response could not be processed, or if retry attempts for retryable methods fail.
//
// Note:
// - It is the caller's responsibility to close the response body when the request is successful to avoid resource leaks.
// - The function adheres to best practices for HTTP communication in Go, ensuring robust error handling and efficient resource management.
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

	// Add the requestID to the context
	ctx = context.WithValue(ctx, requestIDKey{}, requestID)

	// Prepare the HTTP request
	req, err := c.prepareRequest(method, endpoint, body, log)
	if err != nil {
		errMsg := "Failed to prepare HTTP request"
		log.Error(errMsg, zap.Error(err))
		return nil, fmt.Errorf("%s: %w", errMsg, err)
	}

	// Determine if the request should use retry logic based on the http method
	var resp *http.Response
	if retryableHTTPMethods()[method] {
		// Execute the request with retry logic for retryable methods
		resp, err = c.executeRequestWithRetry(req, ctx, log)
	} else {
		// Execute the request once without retry logic for non-retryable methods
		resp, err = c.executeRequest(req, ctx, log)
	}

	if err != nil {
		errMsg := "Failed to execute HTTP request"
		log.Error(errMsg, zap.Error(err))
		return nil, fmt.Errorf("%s: %w", errMsg, err)
	}

	// Process the HTTP response
	if err := c.processResponse(resp, out, log); err != nil {
		return resp, err
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
