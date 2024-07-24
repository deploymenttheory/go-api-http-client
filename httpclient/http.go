package httpclient

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Production

type ProdClient struct {
	*http.Client
}

func (c *ProdClient) SetCookieJar(jar http.CookieJar) {
	c.Jar = jar
}

func (c *ProdClient) SetCookies(url *url.URL, cookies []*http.Cookie) {
	c.Jar.SetCookies(url, cookies)
}

func (c *ProdClient) SetCustomTimeout(timeout time.Duration) {
	c.Timeout = timeout
}

func (c *ProdClient) Cookies(url *url.URL) []*http.Cookie {
	return c.Jar.Cookies(url)
}

func (c *ProdClient) SetRedirectPolicy(policy *func(req *http.Request, via []*http.Request) error) {
	c.CheckRedirect = *policy
}

// Mocking

type mockClient struct {
	lockedResponseCode int
}

func (m *mockClient) CloseIdleConnections() {
	panic("invalid function call")
}

func (m *mockClient) Do(req *http.Request) (*http.Response, error) {
	statusString := http.StatusText(m.lockedResponseCode)

	if statusString == "" {
		return nil, fmt.Errorf("unknown response code requested: %d", m.lockedResponseCode)
	}

	response := &http.Response{StatusCode: m.lockedResponseCode}

	return response, nil
}

func (m *mockClient) Get(_ string) (*http.Response, error) {
	return m.Do(nil)
}

func (m *mockClient) Head(_ string) (*http.Response, error) {
	return m.Do(nil)
}

func (m *mockClient) Post(_ string, _ string, _ io.Reader) (*http.Response, error) {
	return m.Do(nil)
}

func (m *mockClient) PostForm(_ string, _ url.Values) (*http.Response, error) {
	return m.Do(nil)
}

func (m *mockClient) SetCookieJar(jar http.CookieJar) {}

func (m *mockClient) SetCookies(url *url.URL, cookies []*http.Cookie) {}

func (m *mockClient) SetCustomTimeout(time.Duration) {}

func (m *mockClient) Cookies(*url.URL) []*http.Cookie {
	return nil
}

func (m *mockClient) SetRedirectPolicy(*func(req *http.Request, via []*http.Request) error) {}
