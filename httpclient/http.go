package httpclient

import (
	"net/http"
	"net/url"
	"time"
)

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

type testClient struct {
}

func (m *testClient) Do(req *http.Request) (*http.Response, error) {
	// do some stuff which makes a response you like
	return nil, nil
}
