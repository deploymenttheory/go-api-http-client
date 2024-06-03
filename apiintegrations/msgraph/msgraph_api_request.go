// apiintegrations/msgraph/msgraph_api_request.go
package msgraph

import (
	"bytes"
	"encoding/json"
	"mime/multipart"

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

// MarshalMultipartRequest handles multipart form data encoding with secure file handling and returns the encoded body and content type.
func (g *GraphAPIHandler) MarshalMultipartRequest(formFields map[string]string, fileContents map[string][]byte, log *zap.Logger) ([]byte, string, string, error) {
	const snippetLength = 20
	var b bytes.Buffer
	writer := multipart.NewWriter(&b)

	// Log form fields
	for key, val := range formFields {
		err := writer.WriteField(key, val)
		if err != nil {
			log.Error("Failed to add form field to multipart request", zap.String("key", key), zap.Error(err))
			return nil, "", "", err
		}
		log.Debug("Added form field", zap.String("key", key), zap.String("value", val))
	}

	// Log file contents snippets
	for key, val := range fileContents {
		contentSnippet := string(val)
		if len(contentSnippet) > snippetLength {
			contentSnippet = contentSnippet[:snippetLength] + "..."
		}
		log.Debug("File content snippet", zap.String("key", key), zap.String("snippet", contentSnippet))

		part, err := writer.CreateFormFile(key, key)
		if err != nil {
			log.Error("Failed to create form file in multipart request", zap.String("key", key), zap.Error(err))
			return nil, "", "", err
		}
		_, err = part.Write(val)
		if err != nil {
			log.Error("Failed to write file to multipart request", zap.String("key", key), zap.Error(err))
			return nil, "", "", err
		}
	}

	// Close the writer
	err := writer.Close()
	if err != nil {
		log.Error("Failed to close multipart writer", zap.Error(err))
		return nil, "", "", err
	}

	log.Debug("Multipart request constructed", zap.Any("formFields", formFields))

	return b.Bytes(), writer.FormDataContentType(), b.String()[:snippetLength], nil
}
