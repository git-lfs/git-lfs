package cookiejarparser

import (
	"bufio"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/publicsuffix"
)

const httpOnlyPrefix = "#HttpOnly_"

func parseCookieLine(cookieLine string, lineNum int) (*http.Cookie, error) {
	var err error
	cookieLineHttpOnly := false
	if strings.HasPrefix(cookieLine, httpOnlyPrefix) {
		cookieLineHttpOnly = true
		cookieLine = strings.TrimPrefix(cookieLine, httpOnlyPrefix)
	}

	if strings.HasPrefix(cookieLine, "#") || cookieLine == "" {
		return nil, nil
	}

	cookieFields := strings.Split(cookieLine, "\t")

	if len(cookieFields) < 6 || len(cookieFields) > 7 {
		return nil, fmt.Errorf("incorrect number of fields in line %d.  Expected 6 or 7, got %d.", lineNum, len(cookieFields))
	}

	for i, v := range cookieFields {
		cookieFields[i] = strings.TrimSpace(v)
	}

	cookie := &http.Cookie{
		Domain:   cookieFields[0],
		Path:     cookieFields[2],
		Name:     cookieFields[5],
		HttpOnly: cookieLineHttpOnly,
	}
	cookie.Secure, err = strconv.ParseBool(cookieFields[3])
	if err != nil {
		return nil, err
	}
	expiresInt, err := strconv.ParseInt(cookieFields[4], 10, 64)
	if err != nil {
		return nil, err
	}
	if expiresInt > 0 {
		cookie.Expires = time.Unix(expiresInt, 0)
	}

	if len(cookieFields) == 7 {
		cookie.Value = cookieFields[6]
	}

	return cookie, nil
}

// LoadCookieJarFile takes a path to a curl (netscape) cookie jar file and crates a go http.CookieJar with the contents
func LoadCookieJarFile(path string) (http.CookieJar, error) {
	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		return nil, err
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	lineNum := 1
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		cookieLine := scanner.Text()
		cookie, err := parseCookieLine(cookieLine, lineNum)
		if cookie == nil {
			continue
		}
		if err != nil {
			return nil, err
		}

		var cookieScheme string
		if cookie.Secure {
			cookieScheme = "https"
		} else {
			cookieScheme = "http"
		}
		cookieUrl := &url.URL{
			Scheme: cookieScheme,
			Host:   cookie.Domain,
		}

		cookies := jar.Cookies(cookieUrl)
		cookies = append(cookies, cookie)
		jar.SetCookies(cookieUrl, cookies)

		lineNum++
	}

	return jar, nil
}
