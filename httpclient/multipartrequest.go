// httpclient/multipartrequest.go
package httpclient

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/deploymenttheory/go-api-http-client/response"
	"go.uber.org/zap"
)

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
func (c *Client) DoMultipartRequest(method, endpoint string, fields map[string]string, files map[string]string, out interface{}) (*http.Response, error) {
	log := c.Logger

	requestData, _, err := (*c.Integration).MarshalMultipartRequest(fields, files)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s%s", (*c.Integration).Domain(), endpoint)

	req, err := http.NewRequest(method, url, bytes.NewBuffer(requestData))
	if err != nil {
		return nil, err
	}

	err = (*c.Integration).PrepRequestParamsForIntegration(req, c.config.TokenRefreshBufferPeriod)
	if err != nil {
		return nil, err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		log.Error("Failed to send multipart request", zap.String("method", method), zap.String("endpoint", endpoint), zap.Error(err))
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, response.HandleAPIErrorResponse(resp, log)
	} else {
		return resp, response.HandleAPISuccessResponse(resp, out, log)
	}
}
