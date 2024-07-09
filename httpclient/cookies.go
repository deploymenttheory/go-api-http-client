package httpclient

import (
	"net/http/cookiejar"
	"net/url"
)

// loadCustomCookies applies the custom cookies supplied in the config and applies them to the http session.
func (c *Client) loadCustomCookies() error {
	c.Sugar.Debug("initilizing cookie jar")

	cookieJar, err := cookiejar.New(nil)
	if err != nil {
		return err
	}

	c.http.Jar = cookieJar

	cookieUrl, err := url.Parse((*c.Integration).GetFQDN())
	c.Sugar.Debug("cookie URL set globally to: %s", cookieUrl)
	if err != nil {
		return err
	}

	c.http.Jar.SetCookies(cookieUrl, c.config.CustomCookies)

	if c.config.HideSensitiveData {
		c.Sugar.Debug("[REDACTED] cookies set successfully")
	} else {
		c.Sugar.Debug("custom cookies set: %v", c.http.Jar.Cookies(cookieUrl))
	}

	return nil
}
