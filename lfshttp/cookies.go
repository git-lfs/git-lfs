package lfshttp

import (
	"fmt"
	"net/http"

	"github.com/google/slothfs/cookie"
)

func isCookieJarEnabledForHost(c *Client, host string) bool {
	_, cookieFileOk := c.uc.Get("http", fmt.Sprintf("https://%v", host), "cookieFile")

	return cookieFileOk
}

func getCookieJarForHost(c *Client, host string) (http.CookieJar, error) {
cookieFile, _ := c.uc.Get("http", fmt.Sprintf("https://%v", host), "cookieFile")

	return cookie.NewJar(cookieFile)
}
