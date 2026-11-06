package httpclient

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestRequestHeadersBootstrapAndRefreshedTokens(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		call := calls.Add(1)
		if req.Header.Get("Authorization") != "Basic proxy-secret" || req.Header.Get("X-Company") != "company" {
			t.Error("proxy headers missing")
		}
		if call == 1 {
			if req.URL.Path != "/proxypath/api/v1/oauth/token" || req.Header.Get("X-Jamf-Token") != "" {
				t.Error("bootstrap path or headers changed")
			}
		} else if req.Header.Get("X-Jamf-Token") != "Bearer token-"+string(rune('0'+call-1)) {
			t.Error("current bearer token was not moved to the configured header")
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()
	config := RequestHeaderConfig{
		CustomHeaders:     map[string]string{"Authorization": "Basic proxy-secret", "X-Company": "company"},
		BearerTokenHeader: "X-Jamf-Token",
	}
	client, err := config.HTTPClient(http.Client{Timeout: time.Second}, server.URL+"/proxypath")
	if err != nil {
		t.Fatal(err)
	}
	// Config maps are copied; callers cannot change an already-built policy.
	config.CustomHeaders["Authorization"] = "changed after build"
	for _, token := range []string{"", "token-1", "token-2"} {
		endpoint := "/proxypath/api/v1/jamf-pro-version"
		if token == "" {
			endpoint = "/proxypath/api/v1/oauth/token"
		}
		req, err := http.NewRequest(http.MethodGet, server.URL+endpoint, nil)
		if err != nil {
			t.Fatal(err)
		}
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		if req.Header.Get("X-Jamf-Token") != "" || (token != "" && req.Header.Get("Authorization") != "Bearer "+token) {
			t.Fatal("request header policy modified the caller's request")
		}
	}
	if calls.Load() != 3 {
		t.Fatal("unexpected request count")
	}
}

func TestRequestHeadersRejectInvalidConfiguration(t *testing.T) {
	for name, config := range map[string]RequestHeaderConfig{
		"invalid token header":   {BearerTokenHeader: "X-Token\r\nInjected"},
		"invalid name":           {CustomHeaders: map[string]string{"Bad Header": "secret"}},
		"invalid value":          {CustomHeaders: map[string]string{"X-Key": "secret\r\nInjected: value"}},
		"managed auth collision": {CustomHeaders: map[string]string{"authorization": "secret"}},
		"custom token collision": {CustomHeaders: map[string]string{"x-token": "secret"}, BearerTokenHeader: "X-Token"},
		"duplicate case":         {CustomHeaders: map[string]string{"X-Key": "first-secret", "x-key": "second-secret"}},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := config.HTTPClient(http.Client{}, "https://example.com")
			if err == nil || strings.Contains(err.Error(), "secret") {
				t.Fatal("invalid headers must fail without exposing their values")
			}
		})
	}
}

func TestRequestHeadersPreserveCustomTLSTransport(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Header.Get("X-Company") != "company" {
			t.Error("custom header missing")
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()
	client, err := (RequestHeaderConfig{CustomHeaders: map[string]string{"X-Company": "company"}}).HTTPClient(*server.Client(), server.URL)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := client.Get(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	if err := resp.Body.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestRequestHeadersBlockCrossOriginRedirect(t *testing.T) {
	var foreignCalls atomic.Int32
	foreign := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		foreignCalls.Add(1)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer foreign.Close()
	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		http.Redirect(w, req, foreign.URL, http.StatusFound)
	}))
	defer origin.Close()
	client, err := (RequestHeaderConfig{
		CustomHeaders: map[string]string{"Authorization": "Basic proxy-secret"}, BearerTokenHeader: "X-Jamf-Token",
	}).HTTPClient(http.Client{}, origin.URL)
	if err != nil {
		t.Fatal(err)
	}
	req, err := http.NewRequest(http.MethodGet, origin.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer jamf-secret")
	resp, err := client.Do(req)
	if resp != nil {
		resp.Body.Close()
	}
	if !errors.Is(err, errCrossOriginHeaderRedirect) || foreignCalls.Load() != 0 {
		t.Fatal("configured credentials must not follow a redirect to another origin")
	}
}

func TestRequestHeadersWithURLProxyCredentials(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		username, password, ok := req.BasicAuth()
		if !ok || username != "proxy-user" || password != "proxy-password" || req.Header.Get("X-Jamf-Token") != "Bearer fresh-token" {
			t.Error("URL proxy credentials or bearer token were lost")
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()
	client, err := (RequestHeaderConfig{BearerTokenHeader: "X-Jamf-Token"}).HTTPClient(http.Client{}, server.URL)
	if err != nil {
		t.Fatal(err)
	}
	req, err := http.NewRequest(http.MethodGet, strings.Replace(server.URL, "://", "://proxy-user:proxy-password@", 1), nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer fresh-token")
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
}

func TestRequestHeadersBuildAndSensitiveLogging(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Header.Get("Authorization") != "Basic proxy-secret" || req.Header.Get("X-Company-Key") != "company-secret" {
			t.Error("ClientConfig did not apply custom headers")
		}
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{}`)
	}))
	defer server.Close()
	core, logs := observer.New(zapcore.DebugLevel)
	client, err := (&ClientConfig{
		Integration: &stubIntegration{baseURL: server.URL}, Sugar: zap.New(core).Sugar(), HideSensitiveData: true,
		RequestHeaders: RequestHeaderConfig{
			CustomHeaders: map[string]string{"Authorization": "Basic proxy-secret", "X-Company-Key": "company-secret"}, BearerTokenHeader: "X-Jamf-Token",
		},
	}).Build()
	if err != nil {
		t.Fatal(err)
	}
	resp, err := client.DoRequest(http.MethodGet, "/api/v1/jamf-pro-version", nil, nil)
	if resp != nil {
		resp.Body.Close()
	}
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range logs.All() {
		if strings.Contains(entry.Message, "secret") {
			t.Fatal("custom header values leaked into a log message")
		}
		if _, raw := entry.ContextMap()["raw_response"]; raw {
			t.Fatal("raw response contains request credentials when sensitive logging is hidden")
		}
	}
}
