package httpclient

import (
	"net/http/cookiejar"
	"net/url"
)

func (c *Client) loadCustomCookies() error {
	cookieJar, err := cookiejar.New(nil)
	if err != nil {
		return err
	}
	c.http.Jar = cookieJar

	cookieUrl, err := url.Parse((*c.Integration).GetFQDN())
	if err != nil {
		return err
	}

	c.http.Jar.SetCookies(cookieUrl, c.config.CustomCookies)
	c.Sugar.Debug("custom cookies set: %v", c.http.Jar.Cookies(cookieUrl))

	return nil
}
