// apiintegrations/msgraph/msgraph_api_request.go
package msgraph

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"path/filepath"

	"github.com/deploymenttheory/go-api-http-client/helpers"
	"github.com/deploymenttheory/go-api-http-client/logger"
	"go.uber.org/zap"
)

// MarshalRequest encodes the request body as JSON for the Microsoft Graph API.
func (g *GraphAPIHandler) MarshalRequest(body interface{}, method string, endpoint string, log logger.Logger) ([]byte, error) {
	// Marshal the body as JSON
	data, err := json.Marshal(body)
	if err != nil {
		g.Logger.Error("Failed marshaling JSON request", zap.Error(err))
		return nil, err
	}

	// Log the JSON request body for POST, PUT, or PATCH methods
	if method == "POST" || method == "PUT" || method == "PATCH" {
		g.Logger.Debug("JSON Request Body", zap.String("Body", string(data)))
	}

	return data, nil
}

// MarshalMultipartRequest creates a multipart request body with the provided form fields and files.
func (g *GraphAPIHandler) MarshalMultipartRequest(fields map[string]string, files map[string]string, log logger.Logger) ([]byte, string, string, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add the simple fields to the form data
	for field, value := range fields {
		if err := writer.WriteField(field, value); err != nil {
			return nil, "", "", err
		}
	}

	// Add the files to the form data, using safeOpenFile to ensure secure file access
	for formField, filePath := range files {
		file, err := helpers.SafeOpenFile(filePath)
		if err != nil {
			log.Error("Failed to open file securely", zap.String("file", filePath), zap.Error(err))
			return nil, "", "", err
		}
		defer file.Close()

		part, err := writer.CreateFormFile(formField, filepath.Base(filePath))
		if err != nil {
			return nil, "", "", err
		}
		if _, err := io.Copy(part, file); err != nil {
			return nil, "", "", err
		}
	}

	// Close the writer to finish writing the multipart message
	if err := writer.Close(); err != nil {
		return nil, "", "", err
	}

	contentType := writer.FormDataContentType()
	bodyBytes := body.Bytes()

	// Extract the first and last parts of the body
	const logSegmentSize = 1024 // 1 KB
	bodyLen := len(bodyBytes)
	var logBody string
	if bodyLen <= 2*logSegmentSize {
		logBody = string(bodyBytes)
	} else {
		logBody = string(bodyBytes[:logSegmentSize]) + "..." + string(bodyBytes[bodyLen-logSegmentSize:])
	}

	return bodyBytes, contentType, logBody, nil
}
