package httpclient

import (
	"net/http"
	"net/url"
)

func (c *Client) parseCustomCookies(cookies []*http.Cookie) error {
	c.Logger.Debug("FLAG-1")
	cookieUrl, err := url.Parse((*c.Integration).Domain())
	c.Logger.Debug("FLAG-2")
	if err != nil {
		return err
	}
	c.Logger.Debug("FLAG-3")
	c.http.Jar.SetCookies(cookieUrl, cookies)
	c.Logger.Debug("FLAG-4")

	return nil
}
