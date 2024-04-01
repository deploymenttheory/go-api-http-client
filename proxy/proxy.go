// proxy.go

package proxy

import (
	"net/http"
	"net/url"

	"github.com/deploymenttheory/go-api-http-client/logger"
	"go.uber.org/zap"
)

// InitializeProxy initializes the proxy configuration based on the provided options.
// It supports proxy authentication using username/password or an authentication token (e.g., for SSO).
func InitializeProxy(httpClient *http.Client, proxyURL, proxyUsername, proxyPassword, authToken string, log logger.Logger) error {
	if proxyURL == "" {
		return nil // No proxy configuration provided, nothing to do
	}

	parsedProxyURL, err := url.Parse(proxyURL)
	if err != nil {
		log.Error("Failed to parse proxy URL", zap.Error(err))
		return err
	}

	// Initialize proxyAuth variable for username/password authentication
	var proxyAuth *url.Userinfo
	if proxyUsername != "" && proxyPassword != "" {
		proxyAuth = url.UserPassword(proxyUsername, proxyPassword)
	}

	// Set up proxy with authentication if necessary
	if proxyAuth != nil {
		parsedProxyURL.User = proxyAuth
		httpClient.Transport = &http.Transport{
			Proxy: http.ProxyURL(parsedProxyURL),
			ProxyConnectHeader: http.Header{
				"Proxy-Authorization": []string{proxyAuth.String()},
			},
		}
	} else if authToken != "" {
		// SSO authentication
		// Assuming authToken is passed in a configurable way (e.g., as a header)
		httpClient.Transport = &http.Transport{
			Proxy: http.ProxyURL(parsedProxyURL),
			ProxyConnectHeader: http.Header{
				"Authorization": []string{"Bearer " + authToken},
			},
		}
	} else {
		// Proxy without authentication
		httpClient.Transport = &http.Transport{
			Proxy: http.ProxyURL(parsedProxyURL),
		}
	}

	log.Info("Proxy configured", zap.String("ProxyURL", proxyURL))
	return nil
}
