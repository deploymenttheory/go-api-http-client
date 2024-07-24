package httpclient

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Production

type ProdExecutor struct {
	*http.Client
}

func (c *ProdExecutor) SetCookieJar(jar http.CookieJar) {
	c.Jar = jar
}

func (c *ProdExecutor) SetCookies(url *url.URL, cookies []*http.Cookie) {
	c.Jar.SetCookies(url, cookies)
}

func (c *ProdExecutor) SetCustomTimeout(timeout time.Duration) {
	c.Timeout = timeout
}

func (c *ProdExecutor) Cookies(url *url.URL) []*http.Cookie {
	return c.Jar.Cookies(url)
}

func (c *ProdExecutor) SetRedirectPolicy(policy *func(req *http.Request, via []*http.Request) error) {
	c.CheckRedirect = *policy
}

// Mocking

type MockExecutor struct {
	LockedResponseCode int
	ResponseBody       string
}

func (m *MockExecutor) CloseIdleConnections() {
	panic("invalid function call")
}

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

func (m *MockExecutor) Get(_ string) (*http.Response, error) {
	return m.Do(nil)
}

func (m *MockExecutor) Head(_ string) (*http.Response, error) {
	return m.Do(nil)
}

func (m *MockExecutor) Post(_ string, _ string, _ io.Reader) (*http.Response, error) {
	return m.Do(nil)
}

func (m *MockExecutor) PostForm(_ string, _ url.Values) (*http.Response, error) {
	return m.Do(nil)
}

func (m *MockExecutor) SetCookieJar(jar http.CookieJar) {}

func (m *MockExecutor) SetCookies(url *url.URL, cookies []*http.Cookie) {}

func (m *MockExecutor) SetCustomTimeout(time.Duration) {}

func (m *MockExecutor) Cookies(*url.URL) []*http.Cookie {
	return nil
}

func (m *MockExecutor) SetRedirectPolicy(*func(req *http.Request, via []*http.Request) error) {}
