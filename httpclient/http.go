package httpclient

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// ProdExecutor wraps http.Client and implements functions to adjust some of it's attrs
type ProdExecutor struct {
	*http.Client
}

// SetCookieJar func to wrap c.Jar = jar
func (c *ProdExecutor) SetCookieJar(jar http.CookieJar) {
	c.Jar = jar
}

// SetCookies wraps http.Client.Jar.SetCookies
func (c *ProdExecutor) SetCookies(url *url.URL, cookies []*http.Cookie) {
	c.Jar.SetCookies(url, cookies)
}

// SetCustomTimeout wraps http.Client.Timeout = timeout
func (c *ProdExecutor) SetCustomTimeout(timeout time.Duration) {
	c.Timeout = timeout
}

// Cookies wraps http.Client.Jar.Cookies()
func (c *ProdExecutor) Cookies(url *url.URL) []*http.Cookie {
	return c.Jar.Cookies(url)
}

// SetRedirectPolicy wraps http.Client.Jar.CheckRedirect =
func (c *ProdExecutor) SetRedirectPolicy(policy *func(req *http.Request, via []*http.Request) error) {
	c.CheckRedirect = *policy
}

// Mocking

// MockExecutor implements the same function pattern above but allows controllable responses for mocking/testing
type MockExecutor struct {
	LockedResponseCode int
	ResponseBody       string
}

// CloseIdleConnections does nothing.
func (m *MockExecutor) CloseIdleConnections() {
}

// Do returns a http.Response with controllable body and status code.
func (m *MockExecutor) Do(req *http.Request) (*http.Response, error) {
	statusString := http.StatusText(m.LockedResponseCode)

	if statusString == "" {
		return nil, fmt.Errorf("unknown response code requested: %d", m.LockedResponseCode)
	}

	response := &http.Response{
		StatusCode: m.LockedResponseCode,
		Body:       io.NopCloser(bytes.NewBufferString(m.ResponseBody)),
		Header:     make(http.Header),
	}

	return response, nil
}

// Get returns Do
func (m *MockExecutor) Get(_ string) (*http.Response, error) {
	return m.Do(nil)
}

// Head returns Do
func (m *MockExecutor) Head(_ string) (*http.Response, error) {
	return m.Do(nil)
}

// Post returns Do
func (m *MockExecutor) Post(_ string, _ string, _ io.Reader) (*http.Response, error) {
	return m.Do(nil)
}

// PostForm returns Do
func (m *MockExecutor) PostForm(_ string, _ url.Values) (*http.Response, error) {
	return m.Do(nil)
}

// SetCookieJar does nothing yet
func (m *MockExecutor) SetCookieJar(jar http.CookieJar) {}

// SetCookies does nothing yet
func (m *MockExecutor) SetCookies(url *url.URL, cookies []*http.Cookie) {}

// SetCustomTimeout does nothing yet
func (m *MockExecutor) SetCustomTimeout(time.Duration) {}

// Cookies does nothing yet
func (m *MockExecutor) Cookies(*url.URL) []*http.Cookie {
	return nil
}

// SetRedirectPolicy does nothing
func (m *MockExecutor) SetRedirectPolicy(*func(req *http.Request, via []*http.Request) error) {}
