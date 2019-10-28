package lfshttp

import (
	"fmt"
	"net/http"

	"github.com/git-lfs/git-lfs/tools"
	"github.com/ssgelm/cookiejarparser"
)

func isCookieJarEnabledForHost(c *Client, host string) bool {
	_, cookieFileOk := c.uc.Get("http", fmt.Sprintf("https://%v", host), "cookieFile")

	return cookieFileOk
}

func getCookieJarForHost(c *Client, host string) (http.CookieJar, error) {
	cookieFile, _ := c.uc.Get("http", fmt.Sprintf("https://%v", host), "cookieFile")

	cookieFilePath, err := tools.ExpandPath(cookieFile, false)
	if err != nil {
		return nil, err
	}

	return cookiejarparser.LoadCookieJarFile(cookieFilePath)
}
