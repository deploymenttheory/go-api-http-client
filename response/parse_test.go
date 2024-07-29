// response/parse.go
package response

import (
	"reflect"
	"testing"
)

func Test_parseHeader(t *testing.T) {
	type args struct {
		header string
	}
	tests := []struct {
		name  string
		args  args
		want  string
		want1 map[string]string
	}{
		{
			name: "testing content type",
			args: args{
				header: "content-type:application/json;something",
			},
			want: "application/json",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := parseHeader(tt.args.header)
			if got != tt.want {
				t.Errorf("parseHeader() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("parseHeader() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
