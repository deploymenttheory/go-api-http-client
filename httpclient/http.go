package httpclient

import (
	"net/http"
	"net/url"
	"time"
)

type prodClient struct {
	*http.Client
}

func (c *prodClient) SetCookieJar(jar http.CookieJar) {
	c.Jar = jar
}

func (c *prodClient) SetCookies(url *url.URL, cookies []*http.Cookie) {
	c.Jar.SetCookies(url, cookies)
}

func (c *prodClient) SetCustomTimeout(timeout time.Duration) {
	c.Timeout = timeout
}

func (c *prodClient) Cookies(url *url.URL) []*http.Cookie {
	return c.Jar.Cookies(url)
}

func (c *prodClient) SetRedirectPolicy(policy *func(req *http.Request, via []*http.Request) error) {
	c.CheckRedirect = *policy
}
