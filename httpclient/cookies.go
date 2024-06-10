package httpclient

import (
	"fmt"
	"net/http"
	"net/url"
)

func (c *Client) parseCustomCookies(cookiesList []*http.Cookie) error {
	c.Logger.Debug("FLAG-1")
	cookieUrl, err := url.Parse((*c.Integration).Domain())
	c.Logger.Debug(cookieUrl.Host)
	c.Logger.Debug("FLAG-2")
	if err != nil {
		return err
	}
	c.Logger.Debug("FLAG-3")
	c.Logger.Debug(fmt.Sprintf("%+v", cookiesList))
	c.Logger.Debug(fmt.Sprintf("%+v", c.http.Jar))
	c.http.Jar.SetCookies(cookieUrl, cookiesList)
	c.Logger.Debug("FLAG-4")

	return nil
}
