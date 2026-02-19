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
type contentHandler func(io.Reader, any, *zap.SugaredLogger, string) error

// responseUnmarshallers maps MIME types to the corresponding contentHandler functions.
var responseUnmarshallers = map[string]contentHandler{
	"application/json": handlerUnmarshalJSON,
	"application/xml":  handlerUnmarshalXML,
	"text/xml":         handlerUnmarshalXML,
}

// HandleAPISuccessResponse reads the response body, logs the raw response details, and unmarshals the response based on the content type.
func HandleAPISuccessResponse(resp *http.Response, out any, sugar *zap.SugaredLogger) error {
	if resp.Request.Method == "DELETE" {
		return successfulDeleteRequest(resp, sugar)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		sugar.Error("Failed to read response body", zap.Error(err))
		return err
	}

	// TODO do we need to redact some auth headers here? I think so.
	// sugar.Debugw("HTTP Response Headers", zap.Any("Headers", resp.Header))
	sugar.Debugw("Raw HTTP Response", zap.String("Body", string(bodyBytes)))

	bodyReader := bytes.NewReader(bodyBytes)
	contentType := resp.Header.Get("Content-Type")
	contentDisposition := resp.Header.Get("Content-Disposition")

	var handler contentHandler
	var ok bool

	contentTypeNoParams, _ := parseHeader(contentType)

	if handler, ok = responseUnmarshallers[contentTypeNoParams]; ok {
		return handler(bodyReader, out, sugar, contentType)
	}

	if isBinaryData(contentType, contentDisposition) {
		return handleBinaryData(bodyReader, sugar, out, contentDisposition)
	}

	errMsg := fmt.Sprintf("unexpected MIME type: %s", contentType)
	sugar.Errorw("Unmarshal error", zap.String("content type", contentType), zap.Error(errors.New(errMsg)))
	return errors.New(errMsg)

}

// handleDeleteRequest handles the special case for DELETE requests, where a successful response might not contain a body.
func successfulDeleteRequest(resp *http.Response, sugar *zap.SugaredLogger) error {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		sugar.Info("Successfully processed DELETE request", zap.String("URL", resp.Request.URL.String()), zap.Int("Status Code", resp.StatusCode))
		return nil
	}
	return fmt.Errorf("DELETE request failed, status code: %d", resp.StatusCode)
}

// unmarshalJSON unmarshals JSON content from an io.Reader into the provided output structure.
func handlerUnmarshalJSON(reader io.Reader, out any, sugar *zap.SugaredLogger, mimeType string) error {
	decoder := json.NewDecoder(reader)
	if err := decoder.Decode(out); err != nil {
		sugar.Error("JSON Unmarshal error", zap.Error(err))
		return err
	}
	sugar.Info("Successfully unmarshalled JSON response", zap.String("content type", mimeType))
	return nil
}

// unmarshalXML unmarshals XML content from an io.Reader into the provided output structure.
func handlerUnmarshalXML(reader io.Reader, out any, sugar *zap.SugaredLogger, mimeType string) error {
	decoder := xml.NewDecoder(reader)
	if err := decoder.Decode(out); err != nil {
		sugar.Error("XML Unmarshal error", zap.Error(err))
		return err
	}
	sugar.Info("Successfully unmarshalled XML response", zap.String("content type", mimeType))
	return nil
}

// isBinaryData checks if the MIME type or Content-Disposition indicates binary data.
func isBinaryData(contentType, contentDisposition string) bool {
	return strings.Contains(contentType, "application/octet-stream") || strings.HasPrefix(contentDisposition, "attachment")
}

// handleBinaryData reads binary data from an io.Reader and stores it in *[]byte or streams it to an io.Writer.
func handleBinaryData(reader io.Reader, sugar *zap.SugaredLogger, out any, contentDisposition string) error {
	switch out := out.(type) {
	case *[]byte:
		data, err := io.ReadAll(reader)
		if err != nil {
			sugar.Error("Failed to read binary data", zap.Error(err))
			return err
		}
		*out = data

	case io.Writer:
		_, err := io.Copy(out, reader)
		if err != nil {
			sugar.Error("Failed to stream binary data to io.Writer", zap.Error(err))
			return err
		}

	default:
		return errors.New("output parameter is not suitable for binary data (*[]byte or io.Writer)")
	}

	if contentDisposition != "" {
		_, params := parseHeader(contentDisposition)
		if filename, ok := params["filename"]; ok {
			sugar.Debug("Extracted filename from Content-Disposition", zap.String("filename", filename))
		}
	}

	return nil
}
