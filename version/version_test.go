// version_test.go
package version

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestGetUserAgentHeader verifies that the GetUserAgentHeader function returns the expected user agent string
func TestGetUserAgentHeader(t *testing.T) {
	expectedUserAgent := fmt.Sprintf("%s/%s", UserAgentBase, SDKVersion)
	userAgent := GetUserAgentHeader()

	assert.Equal(t, expectedUserAgent, userAgent, "User agent string should match expected format")
}
