package httpclient

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
)

func (c *Client) loadCustomCookies(cookiesList []*http.Cookie) error {
	cookieJar, err := cookiejar.New(nil)
	if err != nil {
		return err
	}

	cookieUrl, err := url.Parse((*c.Integration).Domain())

	if err != nil {
		return err
	}

	c.http.Jar = cookieJar
	c.http.Jar.SetCookies(cookieUrl, cookiesList)
	c.Logger.Debug(fmt.Sprintf("%+v", c.http.Jar))

	return nil
}
