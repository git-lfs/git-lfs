package lfs

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"testing"
)

func TestGetCredentialsForApi(t *testing.T) {
	checkGetCredentials(t, getCredsForAPI, []*getCredentialCheck{
		{
			Desc:     "simple",
			Config:   map[string]string{"lfs.url": "https://git-server.com"},
			Method:   "GET",
			Href:     "https://git-server.com/foo",
			Protocol: "https",
			Host:     "git-server.com",
			Username: "git-server.com",
			Password: "monkey",
		},
		{
			Desc:     "username in url",
			Config:   map[string]string{"lfs.url": "https://user@git-server.com"},
			Method:   "GET",
			Href:     "https://git-server.com/foo",
			Protocol: "https",
			Host:     "git-server.com",
			Username: "user",
			Password: "monkey",
		},
		{
			Desc:          "auth header",
			Config:        map[string]string{"lfs.url": "https://git-server.com"},
			Header:        map[string]string{"Authorization": "Test monkey"},
			Method:        "GET",
			Href:          "https://git-server.com/foo",
			Authorization: "Test monkey",
		},
		{
			Desc:     "scheme mismatch",
			Config:   map[string]string{"lfs.url": "https://git-server.com"},
			Method:   "GET",
			Href:     "http://git-server.com/foo",
			Protocol: "http",
			Host:     "git-server.com",
			Path:     "foo",
			Username: "git-server.com",
			Password: "monkey",
		},
		{
			Desc:     "host mismatch",
			Config:   map[string]string{"lfs.url": "https://git-server.com"},
			Method:   "GET",
			Href:     "https://git-server2.com/foo",
			Protocol: "https",
			Host:     "git-server2.com",
			Path:     "foo",
			Username: "git-server2.com",
			Password: "monkey",
		},
		{
			Desc:     "port mismatch",
			Config:   map[string]string{"lfs.url": "https://git-server.com"},
			Method:   "GET",
			Href:     "https://git-server.com:8080/foo",
			Protocol: "https",
			Host:     "git-server.com:8080",
			Path:     "foo",
			Username: "git-server.com:8080",
			Password: "monkey",
		},
		{
			Desc:          "api url auth",
			Config:        map[string]string{"lfs.url": "https://testuser:testpass@git-server.com"},
			Method:        "GET",
			Href:          "https://git-server.com/foo",
			Authorization: "Basic " + base64.URLEncoding.EncodeToString([]byte("testuser:testpass")),
		},
		{
			Desc:          "git url auth",
			CurrentRemote: "origin",
			Config: map[string]string{
				"lfs.url":           "https://git-server.com",
				"remote.origin.url": "https://gituser:gitpass@git-server.com",
			},
			Method:        "GET",
			Href:          "https://git-server.com/foo",
			Authorization: "Basic " + base64.URLEncoding.EncodeToString([]byte("gituser:gitpass")),
		},
		{
			Desc:     "username in url",
			Config:   map[string]string{"lfs.url": "https://user@git-server.com"},
			Method:   "GET",
			Href:     "https://git-server.com/foo",
			Protocol: "https",
			Host:     "git-server.com",
			Username: "user",
			Password: "monkey",
		},
		{
			Desc:     "?token query",
			Config:   map[string]string{"lfs.url": "https://git-server.com"},
			Method:   "GET",
			Href:     "https://git-server.com/foo?token=abc",
			SkipAuth: true,
		},
	})
}

func TestGetCredentials(t *testing.T) {
	checks := []*getCredentialCheck{
		{
			Desc:     "git server",
			Method:   "GET",
			Href:     "https://git-server.com/foo",
			Protocol: "https",
			Host:     "git-server.com",
			Username: "git-server.com",
			Password: "monkey",
		},
		{
			Desc:     "separate lfs server",
			Method:   "GET",
			Href:     "https://lfs-server.com/foo",
			Protocol: "https",
			Host:     "lfs-server.com",
			Username: "lfs-server.com",
			Password: "monkey",
		},
		{
			Desc:     "?token query",
			Config:   map[string]string{"lfs.url": "https://git-server.com"},
			Method:   "GET",
			Href:     "https://git-server.com/foo?token=abc",
			SkipAuth: true,
		},
	}

	// these properties should not change the outcome
	for _, check := range checks {
		check.CurrentRemote = "origin"
		check.Config = map[string]string{
			"lfs.url":           "https://testuser:testuser@git-server.com",
			"remote.origin.url": "https://gituser:gitpass@git-server.com",
		}
	}

	checkGetCredentials(t, getCreds, checks)
}

func checkGetCredentials(t *testing.T, getCredsFunc func(*http.Request) (Creds, error), checks []*getCredentialCheck) {
	existingRemote := Config.CurrentRemote
	for _, check := range checks {
		t.Logf("Checking %q", check.Desc)
		Config.CurrentRemote = check.CurrentRemote

		for key, value := range check.Config {
			Config.SetConfig(key, value)
		}

		req, err := http.NewRequest(check.Method, check.Href, nil)
		if err != nil {
			t.Errorf("[%s] %s", check.Desc, err)
			continue
		}

		for key, value := range check.Header {
			req.Header.Set(key, value)
		}

		creds, err := getCredsFunc(req)
		if err != nil {
			t.Errorf("[%s] %s", check.Desc, err)
			continue
		}

		if check.ExpectCreds() {
			if creds == nil {
				t.Errorf("[%s], no credentials returned", check.Desc)
				continue
			}

			if value := creds["protocol"]; len(check.Protocol) > 0 && value != check.Protocol {
				t.Errorf("[%s] bad protocol: %q, expected: %q", check.Desc, value, check.Protocol)
			}

			if value := creds["host"]; len(check.Host) > 0 && value != check.Host {
				t.Errorf("[%s] bad host: %q, expected: %q", check.Desc, value, check.Host)
			}

			if value := creds["username"]; len(check.Username) > 0 && value != check.Username {
				t.Errorf("[%s] bad username: %q, expected: %q", check.Desc, value, check.Username)
			}

			if value := creds["password"]; len(check.Password) > 0 && value != check.Password {
				t.Errorf("[%s] bad password: %q, expected: %q", check.Desc, value, check.Password)
			}

			if value := creds["path"]; len(check.Path) > 0 && value != check.Path {
				t.Errorf("[%s] bad path: %q, expected: %q", check.Desc, value, check.Path)
			}
		} else {
			if creds != nil {
				t.Errorf("[%s], unexpected credentials: %v // %v", check.Desc, creds, check)
				continue
			}
		}

		reqAuth := req.Header.Get("Authorization")
		if check.SkipAuth {
		} else if len(check.Authorization) > 0 {
			if reqAuth != check.Authorization {
				t.Errorf("[%s] Unexpected Authorization header: %s", check.Desc, reqAuth)
			}
		} else {
			rawtoken := fmt.Sprintf("%s:%s", check.Username, check.Password)
			expected := "Basic " + base64.URLEncoding.EncodeToString([]byte(rawtoken))
			if reqAuth != expected {
				t.Errorf("[%s] Bad Authorization. Expected '%s', got '%s'", check.Desc, expected, reqAuth)
			}
		}

		Config.ResetConfig()
		Config.CurrentRemote = existingRemote
	}
}

type getCredentialCheck struct {
	Desc          string
	Config        map[string]string
	Header        map[string]string
	Method        string
	Href          string
	Protocol      string
	Host          string
	Username      string
	Password      string
	Path          string
	Authorization string
	CurrentRemote string
	SkipAuth      bool
}

func (c *getCredentialCheck) ExpectCreds() bool {
	return len(c.Protocol) > 0 || len(c.Host) > 0 || len(c.Username) > 0 ||
		len(c.Password) > 0 || len(c.Path) > 0
}

func init() {
	execCreds = func(input Creds, subCommand string) (Creds, error) {
		output := make(Creds)
		for key, value := range input {
			output[key] = value
		}
		if _, ok := output["username"]; !ok {
			output["username"] = input["host"]
		}
		output["password"] = "monkey"
		return output, nil
	}
}
