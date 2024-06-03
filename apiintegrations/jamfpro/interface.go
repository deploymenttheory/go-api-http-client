// jamfpro_api_handler.go

package jamfpro

import (
	"net/http"
	"time"

	"github.com/deploymenttheory/go-api-http-client/logger"
)

// JamfAPIHandler implements the APIHandler interface for the Jamf Pro API.
type JamfAPIHandler struct {
	BaseDomain           string        // OverrideBaseDomain is used to override the base domain for URL construction.
	InstanceName         string        // InstanceName is the name of the Jamf instance.
	Logger               logger.Logger // Logger is the structured logger used for logging.
	AuthMethodDescriptor string
}

type TokenResponse struct {
	Token   string    `json:"token"`
	Expires time.Time `json:"expires"`
}

func (j *JamfAPIHandler) Token() string {
	return ""
}

func (j *JamfAPIHandler) Domain() string {
	return ""
}

func (j *JamfAPIHandler) AutheMethodDescriptor() string {
	return j.AuthMethodDescriptor
}

func (j *JamfAPIHandler) SetRequestHeaders(method string, req http.Request) http.Request {
	return req
}

func (j *JamfAPIHandler) MarshalRequest(body interface{}, method string, endpoint string) ([]byte, error) {
	return j.marshalRequest(body, method, endpoint, j.Logger)
}

func (j *JamfAPIHandler) MarshalMultipartRequest(fields map[string]string, files map[string]string) ([]byte, string, error) {
	return j.marshalMultipartRequest(fields, files, j.Logger)
}
