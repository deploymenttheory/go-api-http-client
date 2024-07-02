// response/success.go
/* Responsible for handling successful API responses. It reads the response body, logs the raw response details,
and unmarshals the response based on the content type (JSON or XML). */
package response

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"go.uber.org/zap"
)

// contentHandler defines the signature for unmarshaling content from an io.Reader.
type contentHandler func(io.Reader, interface{}, *zap.SugaredLogger, string) error

// responseUnmarshallers maps MIME types to the corresponding contentHandler functions.
var responseUnmarshallers = map[string]contentHandler{
	"application/json": unmarshalJSON,
	"application/xml":  unmarshalXML,
	"text/xml":         unmarshalXML,
}

// HandleAPISuccessResponse reads the response body, logs the raw response details, and unmarshals the response based on the content type.
func HandleAPISuccessResponse(resp *http.Response, out interface{}, log *zap.SugaredLogger) error {
	if resp.Request.Method == "DELETE" {
		return handleDeleteRequest(resp, log)
	}

	// Read the response body into a buffer
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error("Failed to read response body", zap.Error(err))
		return err
	}

	// After reading, reset resp.Body so it can be read again.
	log.Debug("HTTP Response Headers", zap.Any("Headers", resp.Header))
	log.Debug("Raw HTTP Response", zap.String("Body", string(bodyBytes)))

	// Use the buffer to create a new io.Reader for unmarshalling
	bodyReader := bytes.NewReader(bodyBytes)

	mimeType, _ := ParseContentTypeHeader(resp.Header.Get("Content-Type"))
	contentDisposition := resp.Header.Get("Content-Disposition")

	if handler, ok := responseUnmarshallers[mimeType]; ok {
		return handler(bodyReader, out, log, mimeType)
	} else if isBinaryData(mimeType, contentDisposition) {
		return handleBinaryData(bodyReader, log, out, mimeType, contentDisposition)
	} else {
		errMsg := fmt.Sprintf("unexpected MIME type: %s", mimeType)
		log.Error("Unmarshal error", zap.String("content type", mimeType), zap.Error(errors.New(errMsg)))
		return errors.New(errMsg)
	}
}

// handleDeleteRequest handles the special case for DELETE requests, where a successful response might not contain a body.
func handleDeleteRequest(resp *http.Response, log *zap.SugaredLogger) error {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		log.Info("Successfully processed DELETE request", zap.String("URL", resp.Request.URL.String()), zap.Int("Status Code", resp.StatusCode))
		return nil
	}
	return fmt.Errorf("DELETE request failed, status code: %d", resp.StatusCode)
}

// unmarshalJSON unmarshals JSON content from an io.Reader into the provided output structure.
func unmarshalJSON(reader io.Reader, out interface{}, log *zap.SugaredLogger, mimeType string) error {
	decoder := json.NewDecoder(reader)
	if err := decoder.Decode(out); err != nil {
		log.Error("JSON Unmarshal error", zap.Error(err))
		return err
	}
	log.Info("Successfully unmarshalled JSON response", zap.String("content type", mimeType))
	return nil
}

// unmarshalXML unmarshals XML content from an io.Reader into the provided output structure.
func unmarshalXML(reader io.Reader, out interface{}, log *zap.SugaredLogger, mimeType string) error {
	decoder := xml.NewDecoder(reader)
	if err := decoder.Decode(out); err != nil {
		log.Error("XML Unmarshal error", zap.Error(err))
		return err
	}
	log.Info("Successfully unmarshalled XML response", zap.String("content type", mimeType))
	return nil
}

// isBinaryData checks if the MIME type or Content-Disposition indicates binary data.
func isBinaryData(contentType, contentDisposition string) bool {
	return strings.Contains(contentType, "application/octet-stream") || strings.HasPrefix(contentDisposition, "attachment")
}

// handleBinaryData reads binary data from an io.Reader and stores it in *[]byte or streams it to an io.Writer.
func handleBinaryData(reader io.Reader, log *zap.SugaredLogger, out interface{}, mimeType, contentDisposition string) error {
	// Check if the output interface is either *[]byte or io.Writer
	switch out := out.(type) {
	case *[]byte:
		// Read all data from reader and store it in *[]byte
		data, err := io.ReadAll(reader)
		if err != nil {
			log.Error("Failed to read binary data", zap.Error(err))
			return err
		}
		*out = data

	case io.Writer:
		// Stream data directly to the io.Writer
		_, err := io.Copy(out, reader)
		if err != nil {
			log.Error("Failed to stream binary data to io.Writer", zap.Error(err))
			return err
		}

	default:
		errMsg := "output parameter is not suitable for binary data (*[]byte or io.Writer)"
		log.Error(errMsg, zap.String("Content-Type", mimeType))
		return errors.New(errMsg)
	}

	// Handle Content-Disposition if present
	if contentDisposition != "" {
		_, params := ParseContentDisposition(contentDisposition)
		if filename, ok := params["filename"]; ok {
			log.Debug("Extracted filename from Content-Disposition", zap.String("filename", filename))
			// Additional processing for the filename can be done here if needed
		}
	}

	return nil
}
