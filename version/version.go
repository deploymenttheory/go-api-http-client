// version.go
package version

import "fmt"

const (
	SDKVersion    = "0.1.21"
	UserAgentBase = "go-api-http-client"
)

func GetUserAgentHeader() string {
	return fmt.Sprintf("%s/%s", UserAgentBase, SDKVersion)
}
