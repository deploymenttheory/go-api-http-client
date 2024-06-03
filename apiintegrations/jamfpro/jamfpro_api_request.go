// jamfpro_api_request.go
package jamfpro

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"

	"github.com/deploymenttheory/go-api-http-client/helpers"
	"github.com/deploymenttheory/go-api-http-client/logger"
	"go.uber.org/zap"
)

// MarshalRequest encodes the request body according to the endpoint for the API.
func (j *JamfAPIHandler) MarshalRequest(body interface{}, method string, endpoint string, log logger.Logger) ([]byte, error) {
	var (
		data []byte
		err  error
	)

	// Determine the format based on the endpoint
	format := "json"
	if strings.Contains(endpoint, "/JSSResource") {
		format = "xml"
	} else if strings.Contains(endpoint, "/api") {
		format = "json"
	}

	switch format {
	case "xml":
		data, err = xml.Marshal(body)
		if err != nil {
			return nil, err
		}

		if method == "POST" || method == "PUT" {
			j.Logger.Debug("XML Request Body", zap.String("Body", string(data)))
		}

	case "json":
		data, err = json.Marshal(body)
		if err != nil {
			j.Logger.Error("Failed marshaling JSON request", zap.Error(err))
			return nil, err
		}

		if method == "POST" || method == "PUT" || method == "PATCH" {
			j.Logger.Debug("JSON Request Body", zap.String("Body", string(data)))
		}
	}

	return data, nil
}

// MarshalMultipartRequest creates a multipart request body with the provided form fields and files.
func (j *JamfAPIHandler) MarshalMultipartRequest(fields map[string]string, files map[string]string, log logger.Logger) ([]byte, string, string, error) {
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
