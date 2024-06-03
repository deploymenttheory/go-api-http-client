// httpmethod/httpmethod_test.go
package httpmethod

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestIsIdempotentHTTPMethod tests the IsIdempotentHTTPMethod function with various HTTP methods
func TestIsIdempotentHTTPMethod(t *testing.T) {
	tests := []struct {
		method   string
		expected bool
	}{
		{http.MethodGet, true},
		{http.MethodPut, true},
		{http.MethodDelete, true},
		{http.MethodHead, true},
		{http.MethodOptions, true},
		{http.MethodTrace, true},
		// Non-idempotent methods
		{http.MethodPost, false},
		{http.MethodPatch, false},
		{http.MethodConnect, false},
	}

	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			result := IsIdempotentHTTPMethod(tt.method)
			assert.Equal(t, tt.expected, result, "Idempotency status should match expected for method "+tt.method)
		})
	}
}

// TestIsNonIdempotentHTTPMethod tests the IsNonIdempotentHTTPMethod function with various HTTP methods
func TestIsNonIdempotentHTTPMethod(t *testing.T) {
	tests := []struct {
		method   string
		expected bool
	}{
		// Non-idempotent methods
		{http.MethodPost, true},
		{http.MethodPatch, true},
		{http.MethodConnect, true},
		// Idempotent methods
		{http.MethodGet, false},
		{http.MethodPut, false},
		{http.MethodDelete, false},
		{http.MethodHead, false},
		{http.MethodOptions, false},
		{http.MethodTrace, false},
	}

	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			result := IsNonIdempotentHTTPMethod(tt.method)
			assert.Equal(t, tt.expected, result, "Non-idempotency status should match expected for method "+tt.method)
		})
	}
}
