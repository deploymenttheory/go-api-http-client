// httpclient/request.go
package httpclient

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/deploymenttheory/go-api-http-client/ratehandler"
	"github.com/deploymenttheory/go-api-http-client/response"
	"github.com/deploymenttheory/go-api-http-client/status"
	"go.uber.org/zap"
)

// DoRequest constructs and executes an HTTP request based on the provided method, endpoint, request body, and output variable.
// This function serves as a dispatcher, deciding whether to execute the request with or without retry logic based on the
// idempotency of the HTTP method. Idempotent methods (GET, PUT, DELETE) are executed with retries to handle transient errors
// and rate limits, while non-idempotent methods (POST, PATCH) are executed without retries to avoid potential side effects
// of duplicating non-idempotent operations. The function uses an instance of a logger implementing the logger.Logger interface,
// used to log informational messages, warnings, and errors encountered during the execution of the request.
// It also applies redirect handling to the client if configured, allowing the client to follow redirects up to a maximum
// number of times.

// Parameters:
// - method: A string representing the HTTP method to be used for the request. This method determines the execution path
//   and whether the request will be retried in case of failures.
// - endpoint: The target API endpoint for the request. This should be a relative path that will be appended to the base URL
//   configured for the HTTP client.
// - body: The payload for the request, which will be serialized into the request body. The serialization format (e.g., JSON, XML)
//   is determined by the content-type header and the specific implementation of the API handler used by the client.
// - out: A pointer to an output variable where the response will be deserialized. The function expects this to be a pointer to
//   a struct that matches the expected response schema.

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

func (c *Client) DoRequest(method, endpoint string, body, out interface{}) (*http.Response, error) {
	log := c.Logger

	if IsIdempotentHTTPMethod(method) {
		return c.executeRequestWithRetries(method, endpoint, body, out)
	} else if IsNonIdempotentHTTPMethod(method) {
		return c.executeRequest(method, endpoint, body, out)
	} else {
		return nil, log.Error("HTTP method not supported", zap.String("method", method))
	}
}

// executeRequestWithRetries executes an HTTP request using the specified method, endpoint, request body, and output variable.
// It is designed for idempotent HTTP methods (GET, PUT, DELETE), where the request can be safely retried in case of
// transient errors or rate limiting. The function implements a retry mechanism that respects the client's configuration
// for maximum retry attempts and total retry duration. Each retry attempt uses exponential backoff with jitter to avoid
// thundering herd problems. An instance of a logger (conforming to the logger.Logger interface) is used for logging the
// request, retry attempts, and any errors encountered.
//
// Parameters:
// - method: The HTTP method to be used for the request (e.g., "GET", "PUT", "DELETE").
// - endpoint: The API endpoint to which the request will be sent. This should be a relative path that will be appended
// to the base URL of the HTTP client.
// - body: The request payload, which will be marshaled into the request body based on the content type. Can be nil for
// methods that do not send a payload.
// - out: A pointer to the variable where the unmarshaled response will be stored. The function expects this to be a
// pointer to a struct that matches the expected response schema.
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
func (c *Client) executeRequestWithRetries(method, endpoint string, body, out interface{}) (*http.Response, error) {
	log := c.Logger
	ctx := context.Background()
	totalRetryDeadline := time.Now().Add(c.config.TotalRetryDuration)

	var resp *http.Response
	var err error
	var retryCount int

	log.Debug("Executing request with retries", zap.String("method", method), zap.String("endpoint", endpoint))

	for time.Now().Before(totalRetryDeadline) {
		res, requestErr := c.doRequest(ctx, method, endpoint, body)
		if requestErr != nil {
			return nil, requestErr
		}
		resp = res

		if resp.StatusCode >= 200 && resp.StatusCode < 400 {
			if resp.StatusCode >= 300 {
				log.Warn("Redirect response received", zap.Int("status_code", resp.StatusCode), zap.String("location", resp.Header.Get("Location")))
			}
			return resp, response.HandleAPISuccessResponse(resp, out, log)
		}

		statusMessage := status.TranslateStatusCode(resp)

		if resp != nil && status.IsNonRetryableStatusCode(resp) {
			log.Warn("Non-retryable error received", zap.Int("status_code", resp.StatusCode), zap.String("status_message", statusMessage))
			return resp, response.HandleAPIErrorResponse(resp, log)
		}

		if status.IsRateLimitError(resp) {
			waitDuration := ratehandler.ParseRateLimitHeaders(resp, log)
			if waitDuration > 0 {
				log.Warn("Rate limit encountered, waiting before retrying", zap.Duration("waitDuration", waitDuration))
				time.Sleep(waitDuration)
				continue
			}
		}

		if status.IsTransientError(resp) {
			retryCount++
			if retryCount > c.config.MaxRetryAttempts {
				log.Warn("Max retry attempts reached", zap.String("method", method), zap.String("endpoint", endpoint))
				break
			}
			waitDuration := ratehandler.CalculateBackoff(retryCount)
			log.Warn("Retrying request due to transient error", zap.String("method", method), zap.String("endpoint", endpoint), zap.Int("retryCount", retryCount), zap.Duration("waitDuration", waitDuration), zap.Error(err))
			time.Sleep(waitDuration)
			continue
		}

		if !status.IsRetryableStatusCode(resp.StatusCode) {
			if apiErr := response.HandleAPIErrorResponse(resp, log); apiErr != nil {
				err = apiErr
			}
			log.LogError("request_error", method, endpoint, resp.StatusCode, resp.Status, err, statusMessage)
			break
		}
	}

	if err != nil {
		return nil, err
	}

	return resp, response.HandleAPIErrorResponse(resp, log)
}

// executeRequest executes an HTTP request using the specified method, endpoint, and request body without implementing
// retry logic. It is primarily designed for non-idempotent HTTP methods like POST and PATCH, where the request should
// not be automatically retried within this function due to the potential side effects of re-submitting the same data.
//
// Parameters:
//   - method: The HTTP method to be used for the request, typically "POST" or "PATCH".
//   - endpoint: The API endpoint to which the request will be sent. This should be a relative path that will be appended
//     to the base URL of the HTTP client.
//   - body: The request payload, which will be marshaled into the request body based on the content type. This can be any
//     data structure that can be marshaled into the expected request format (e.g., JSON, XML).
//   - out: A pointer to the variable where the unmarshaled response will be stored. This should be a pointer to a struct
//     that matches the expected response schema.
//
// Returns:
//   - *http.Response: The HTTP response from the server. This includes the status code, headers, and body of the response.
//   - error: An error object if an error occurred during the request execution. This could be due to network issues,
//     server errors, or issues with marshaling/unmarshaling the request/response.
//
// Usage:
// This function is suitable for operations where the request should not be retried automatically, such as data submission
// operations where retrying could result in duplicate data processing. It ensures that the request is executed exactly
// once and provides detailed logging for debugging purposes.
//
// Note:
//   - The caller is responsible for closing the response body to prevent resource leaks.
//   - The function ensures concurrency control by acquiring and releasing a concurrency token before and after the request
//     execution.
//   - The function logs detailed information about the request execution, including the method, endpoint, status code, and
//     any errors encountered.
func (c *Client) executeRequest(method, endpoint string, body, out interface{}) (*http.Response, error) {
	log := c.Logger
	ctx := context.Background()

	log.Debug("Executing request without retries", zap.String("method", method), zap.String("endpoint", endpoint))

	res, err := c.doRequest(ctx, method, endpoint, body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode >= 200 && res.StatusCode < 400 {
		if res.StatusCode >= 300 {
			log.Warn("Redirect response received", zap.Int("status_code", res.StatusCode), zap.String("location", res.Header.Get("Location")))
		}
		return res, response.HandleAPISuccessResponse(res, out, log)
	}

	return nil, response.HandleAPIErrorResponse(res, log)
}

func (c *Client) doRequest(ctx context.Context, method, endpoint string, body interface{}) (*http.Response, error) {
	log := c.Logger
	log.Debug("start of doRequest")

	_, err := (*c.Integration).Token()
	if err != nil {
		return nil, err
	}

	// region concurrency
	ctx, requestID, err := c.Concurrency.AcquireConcurrencyPermit(ctx)
	if err != nil {
		return nil, c.Logger.Error("Failed to acquire concurrency permit", zap.Error(err))

	}

	defer c.Concurrency.ReleaseConcurrencyPermit(requestID)

	c.Concurrency.Metrics.Lock.Lock()
	c.Concurrency.Metrics.TotalRequests++
	c.Concurrency.Metrics.Lock.Unlock()

	// Marshal the request data based on the provided api handler
	requestData, err := (*c.Integration).MarshalRequest(body, method, endpoint)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf((*c.Integration).Domain(), endpoint)
	req, err := http.NewRequest(method, url, bytes.NewBuffer(requestData))
	if err != nil {
		return nil, err
	}

	(*c.Integration).SetRequestHeaders(req)
	req = req.WithContext(ctx)

	startTime := time.Now()
	resp, err := c.http.Do(req)
	if err != nil {
		log.Error("Failed to send request", zap.String("method", method), zap.String("endpoint", endpoint), zap.Error(err))
		return nil, err
	}

	duration := time.Since(startTime)
	c.Concurrency.EvaluateAndAdjustConcurrency(resp, duration)
	log.LogCookies("incoming", req, method, endpoint)
	CheckDeprecationHeader(resp, log)

	log.Debug("Request sent successfully", zap.String("method", method), zap.String("endpoint", endpoint), zap.Int("status_code", resp.StatusCode))

	return resp, nil
}
