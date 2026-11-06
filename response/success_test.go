package response

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"go.uber.org/zap"
)

func TestSuccessResponseWithoutRepresentation(t *testing.T) {
	for _, tc := range []struct {
		name, method string
		status       int
	}{
		{"patch no content", http.MethodPatch, http.StatusNoContent},
		{"put no content", http.MethodPut, http.StatusNoContent},
		{"head without MIME type", http.MethodHead, http.StatusOK},
		{"ignored response body", http.MethodPost, http.StatusCreated},
	} {
		t.Run(tc.name, func(t *testing.T) {
			resp := &http.Response{
				StatusCode: tc.status, Header: make(http.Header),
				Request: &http.Request{Method: tc.method}, Body: io.NopCloser(strings.NewReader("")),
			}
			defer resp.Body.Close()
			if err := HandleAPISuccessResponse(resp, nil, zap.NewNop().Sugar()); err != nil {
				t.Fatalf("successful bodyless response must not need a MIME type: %v", err)
			}
		})
	}
}
