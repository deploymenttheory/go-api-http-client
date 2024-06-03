// apiintegrations/msgraph/msgraph_api_request_test.go
package msgraph

import (
	"encoding/json"
	"testing"

	"github.com/deploymenttheory/go-api-http-client/mocklogger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// TestMarshalRequest tests the MarshalRequest function.
func TestMarshalRequest(t *testing.T) {
	body := map[string]interface{}{
		"name": "John Doe",
		"age":  30,
	}
	method := "POST"
	endpoint := "/users"
	mockLog := mocklogger.NewMockLogger()
	handler := GraphAPIHandler{Logger: mockLog}

	expectedData, _ := json.Marshal(body)

	// Correct the way we setup the logger mock
	mockLog.On("Debug", "JSON Request Body", mock.MatchedBy(func(fields []zap.Field) bool {
		if len(fields) != 1 {
			return false
		}
		return fields[0].Key == "Body" && fields[0].String == string(expectedData)
	})).Once()

	data, err := handler.MarshalRequest(body, method, endpoint, mockLog)

	assert.NoError(t, err)
	assert.Equal(t, expectedData, data)
	mockLog.AssertExpectations(t)
}

// func TestMarshalMultipartRequest(t *testing.T) {
// 	// Prepare the logger mock
// 	mockLog := mocklogger.NewMockLogger()

// 	// Setting up a temporary file to simulate a file upload
// 	tempDir := t.TempDir() // Create a temporary directory for test files
// 	tempFile, err := os.CreateTemp(tempDir, "upload-*.txt")
// 	assert.NoError(t, err)
// 	defer os.Remove(tempFile.Name()) // Ensure the file is removed after the test

// 	_, err = tempFile.WriteString("Test file content")
// 	assert.NoError(t, err)
// 	tempFile.Close()

// 	handler := GraphAPIHandler{Logger: mockLog}

// 	fields := map[string]string{"field1": "value1"}
// 	files := map[string]string{"fileField": tempFile.Name()}

// 	// Execute the function
// 	body, contentType, err := handler.MarshalMultipartRequest(fields, files, mockLog)
// 	assert.NoError(t, err)
// 	assert.Contains(t, contentType, "multipart/form-data; boundary=")

// 	// Check if the multipart form data contains the correct fields and file data
// 	reader := multipart.NewReader(bytes.NewReader(body), strings.TrimPrefix(contentType, "multipart/form-data; boundary="))
// 	var foundField, foundFile bool

// 	for {
// 		part, err := reader.NextPart()
// 		if err == io.EOF {
// 			break
// 		}
// 		assert.NoError(t, err)

// 		if part.FormName() == "field1" {
// 			buf := new(bytes.Buffer)
// 			_, err = buf.ReadFrom(part)
// 			assert.NoError(t, err)
// 			assert.Equal(t, "value1", buf.String())
// 			foundField = true
// 		} else if part.FileName() == filepath.Base(tempFile.Name()) {
// 			buf := new(bytes.Buffer)
// 			_, err = buf.ReadFrom(part)
// 			assert.NoError(t, err)
// 			assert.Equal(t, "Test file content", buf.String())
// 			foundFile = true
// 		}
// 	}

// 	// Ensure all expected parts were found
// 	assert.True(t, foundField, "Text field not found in the multipart form data")
// 	assert.True(t, foundFile, "File not found in the multipart form data")
// }
