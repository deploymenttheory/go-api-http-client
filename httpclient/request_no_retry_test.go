package httpclient

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"go.uber.org/zap"
)

// stubIntegration is a minimal APIIntegration that points every request at a
// test server and applies no authentication.
type stubIntegration struct{ baseURL string }

func (s *stubIntegration) GetFQDN() string                              { return s.baseURL }
func (s *stubIntegration) ConstructURL(endpoint string) string          { return s.baseURL + endpoint }
func (s *stubIntegration) GetAuthMethodDescriptor() string              { return "stub" }
func (s *stubIntegration) CheckRefreshToken() error                     { return nil }
func (s *stubIntegration) PrepRequestParamsAndAuth(*http.Request) error { return nil }
func (s *stubIntegration) GetSessionCookies() ([]*http.Cookie, error)   { return nil, nil }
func (s *stubIntegration) PrepRequestBody(body any, _, _ string) ([]byte, error) {
	if body == nil {
		return nil, nil
	}
	return json.Marshal(body)
}
func (s *stubIntegration) MarshalMultipartRequest(map[string]string, map[string]string) ([]byte, string, error) {
	return nil, "", nil
}

// newTestClient wires a Client to url with retries enabled and no artificial delays.
func newTestClient(url string) *Client {
	var integration APIIntegration = &stubIntegration{baseURL: url}
	sugar := zap.NewNop().Sugar()
	return &Client{
		config: &ClientConfig{
			Integration:            integration,
			Sugar:                  sugar,
			RetryEligiableRequests: true,
			MaxRetryAttempts:       3,
			TotalRetryDuration:     5 * time.Second,
			MandatoryRequestDelay:  0,
		},
		Integration: &integration,
		http:        &http.Client{Timeout: 5 * time.Second},
		Sugar:       sugar,
	}
}

// serverAlways500 counts requests and always replies 500, the transient status
// that drives the retry loop.
func serverAlways500(t *testing.T, calls *int32) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(calls, 1)
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"httpStatus":500,"errors":[]}`))
	}))
	t.Cleanup(srv.Close)
	return srv
}

// Establishes the behaviour DoRequestNoRetry exists to avoid: a PUT is classed
// idempotent and replayed, which for a version-locked write resubmits a token
// the server has already consumed.
func Test_DoRequest_RetriesPut(t *testing.T) {
	var calls int32
	srv := serverAlways500(t, &calls)

	c := newTestClient(srv.URL)
	resp, err := c.DoRequest(http.MethodPut, "/thing", map[string]string{"a": "b"}, nil)
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}

	if err == nil {
		t.Fatal("expected an error from a 500 response")
	}
	if got := atomic.LoadInt32(&calls); got < 2 {
		t.Fatalf("expected DoRequest to replay the PUT, but the server saw %d request(s)", got)
	}
}

// Exhausting the retry budget must return an error, not panic. The loop
// variable used to shadow the outer response, leaving it nil for the
// post-loop error path to dereference.
func Test_DoRequest_ExhaustedRetriesReturnsErrorNotPanic(t *testing.T) {
	var calls int32
	srv := serverAlways500(t, &calls)

	c := newTestClient(srv.URL)
	c.config.MaxRetryAttempts = 1

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("exhausting retries panicked: %v", r)
		}
	}()

	resp, err := c.DoRequest(http.MethodPut, "/thing", map[string]string{"a": "b"}, nil)
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}
	if err == nil {
		t.Fatal("expected an error once the retry budget is spent")
	}
}

func Test_DoRequestNoRetry_SendsExactlyOnce(t *testing.T) {
	var calls int32
	srv := serverAlways500(t, &calls)

	c := newTestClient(srv.URL)
	resp, err := c.DoRequestNoRetry(http.MethodPut, "/thing", map[string]string{"a": "b"}, nil)
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}

	if err == nil {
		t.Fatal("expected an error from a 500 response")
	}
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Fatalf("DoRequestNoRetry must send exactly one request, server saw %d", got)
	}
}

func Test_DoRequestNoRetry_SucceedsAndDecodesBody(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"versionLock":7}`))
	}))
	defer srv.Close()

	var out struct {
		VersionLock int `json:"versionLock"`
	}
	c := newTestClient(srv.URL)
	resp, err := c.DoRequestNoRetry(http.MethodPut, "/thing", map[string]string{"a": "b"}, &out)
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.VersionLock != 7 {
		t.Errorf("response body not decoded: got versionLock=%d, want 7", out.VersionLock)
	}
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Errorf("expected 1 request, server saw %d", got)
	}
}
