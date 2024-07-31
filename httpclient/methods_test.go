// httpmethod/httpmethod.go
package httpclient

import (
	"net/http"
	"testing"
)

func Test_isIdempotentHTTPMethod(t *testing.T) {
	type args struct {
		method string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "testing an indempotent method",
			args: args{
				method: http.MethodGet,
			},
			want: true,
		},
		{
			name: "testing a nonindempotent method",
			args: args{
				method: http.MethodPost,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isIdempotentHTTPMethod(tt.args.method); got != tt.want {
				t.Errorf("isIdempotentHTTPMethod() = %v, want %v", got, tt.want)
			}
		})
	}
}
