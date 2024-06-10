package httpclient

import (
	"net/http"
	"net/url"
)

func (c *Client) parseCustomCookies(cookies []*http.Cookie) error {
	cookieUrl, err := url.Parse((*c.Integration).Domain())
	if err != nil {
		return err
	}

	c.http.Jar.SetCookies(cookieUrl, cookies)

	return nil
}
