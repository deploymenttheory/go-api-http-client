// sdk_version.go
package httpclient

import "fmt"

const (
	SDKVersion    = "0.0.90"
	UserAgentBase = "go-api-http-client"
)

func GetUserAgentHeader() string {
	return fmt.Sprintf("%s/%s", UserAgentBase, SDKVersion)
}
