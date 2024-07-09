// response/parse_test.go
package response

import (
	"reflect"
	"testing"
)

func TestParseContentTypeHeader(t *testing.T) {
	tests := []struct {
		name       string
		header     string
		wantType   string
		wantParams map[string]string
	}{
		{
			name:       "Basic",
			header:     "text/html; charset=UTF-8",
			wantType:   "text/html",
			wantParams: map[string]string{"charset": "UTF-8"},
		},
		{
			name:       "No Params",
			header:     "application/json",
			wantType:   "application/json",
			wantParams: map[string]string{},
		},
		{
			name:       "Multiple Params",
			header:     "multipart/form-data; boundary=something; charset=utf-8",
			wantType:   "multipart/form-data",
			wantParams: map[string]string{"boundary": "something", "charset": "utf-8"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotType, gotParams := parseDispositionHeader(tt.header)
			if gotType != tt.wantType {
				t.Errorf("ParseContentTypeHeader() gotType = %v, want %v", gotType, tt.wantType)
			}
			if !reflect.DeepEqual(gotParams, tt.wantParams) {
				t.Errorf("ParseContentTypeHeader() gotParams = %v, want %v", gotParams, tt.wantParams)
			}
		})
	}
}

func TestParseContentDisposition(t *testing.T) {
	tests := []struct {
		name       string
		header     string
		wantType   string
		wantParams map[string]string
	}{
		{
			name:       "Attachment with Filename",
			header:     "attachment; filename=\"filename.jpg\"",
			wantType:   "attachment",
			wantParams: map[string]string{"filename": "filename.jpg"},
		},
		{
			name:       "Inline",
			header:     "inline",
			wantType:   "inline",
			wantParams: map[string]string{},
		},
		// Add more test cases as needed
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotType, gotParams := parseDispositionHeader(tt.header)
			if gotType != tt.wantType {
				t.Errorf("ParseContentDisposition() gotType = %v, want %v", gotType, tt.wantType)
			}
			if !reflect.DeepEqual(gotParams, tt.wantParams) {
				t.Errorf("ParseContentDisposition() gotParams = %v, want %v", gotParams, tt.wantParams)
			}
		})
	}
}
