package auth

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/bgentry/go-netrc/netrc"
	"github.com/github/git-lfs/config"
)

func TestGetCredentialsForApi(t *testing.T) {
	SetupTestCredentialsFunc()
	defer RestoreCredentialsFunc()

	checkGetCredentials(t, GetCreds, []*getCredentialCheck{
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
			Authorization: "Basic " + strings.TrimSpace(base64.StdEncoding.EncodeToString([]byte("testuser:testpass"))),
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
			Authorization: "Basic " + strings.TrimSpace(base64.StdEncoding.EncodeToString([]byte("gituser:gitpass"))),
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

type fakeNetrc struct{}

func (n *fakeNetrc) FindMachine(host string) *netrc.Machine {
	if host == "some-host" {
		return &netrc.Machine{Login: "abc", Password: "def"}
	}
	return nil
}

func TestNetrcWithHostAndPort(t *testing.T) {
	SetupTestCredentialsFunc()
	defer RestoreCredentialsFunc()

	cfg := config.New()
	cfg.SetNetrc(&fakeNetrc{})
	u, err := url.Parse("http://some-host:123/foo/bar")
	if err != nil {
		t.Fatal(err)
	}

	req := &http.Request{
		URL:    u,
		Header: http.Header{},
	}

	if !setCredURLFromNetrc(cfg, req) {
		t.Fatal("no netrc match")
	}

	auth := req.Header.Get("Authorization")
	if auth != "Basic YWJjOmRlZg==" {
		t.Fatalf("bad basic auth: %q", auth)
	}
}

func TestNetrcWithHost(t *testing.T) {
	SetupTestCredentialsFunc()
	defer RestoreCredentialsFunc()

	cfg := config.New()
	cfg.SetNetrc(&fakeNetrc{})
	u, err := url.Parse("http://some-host/foo/bar")
	if err != nil {
		t.Fatal(err)
	}

	req := &http.Request{
		URL:    u,
		Header: http.Header{},
	}

	if !setCredURLFromNetrc(cfg, req) {
		t.Fatalf("no netrc match")
	}

	auth := req.Header.Get("Authorization")
	if auth != "Basic YWJjOmRlZg==" {
		t.Fatalf("bad basic auth: %q", auth)
	}
}

func TestNetrcWithBadHost(t *testing.T) {
	SetupTestCredentialsFunc()
	defer RestoreCredentialsFunc()

	cfg := config.New()
	cfg.SetNetrc(&fakeNetrc{})
	u, err := url.Parse("http://other-host/foo/bar")
	if err != nil {
		t.Fatal(err)
	}

	req := &http.Request{
		URL:    u,
		Header: http.Header{},
	}

	if setCredURLFromNetrc(cfg, req) {
		t.Fatalf("unexpected netrc match")
	}

	auth := req.Header.Get("Authorization")
	if auth != "" {
		t.Fatalf("bad basic auth: %q", auth)
	}
}

func checkGetCredentials(t *testing.T, getCredsFunc func(*config.Configuration, *http.Request) (Creds, error), checks []*getCredentialCheck) {
	for _, check := range checks {
		t.Logf("Checking %q", check.Desc)
		cfg := config.NewFrom(config.Values{
			Git: check.Config,
		})
		cfg.CurrentRemote = check.CurrentRemote

		req, err := http.NewRequest(check.Method, check.Href, nil)
		if err != nil {
			t.Errorf("[%s] %s", check.Desc, err)
			continue
		}

		for key, value := range check.Header {
			req.Header.Set(key, value)
		}

		creds, err := getCredsFunc(cfg, req)
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
			expected := "Basic " + strings.TrimSpace(base64.StdEncoding.EncodeToString([]byte(rawtoken)))
			if reqAuth != expected {
				t.Errorf("[%s] Bad Authorization. Expected '%s', got '%s'", check.Desc, expected, reqAuth)
			}
		}
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

var (
	TestCredentialsFunc CredentialFunc
	origCredentialsFunc CredentialFunc
)

func init() {
	TestCredentialsFunc = func(cfg *config.Configuration, input Creds, subCommand string) (Creds, error) {
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

// Override the credentials func for testing
func SetupTestCredentialsFunc() {
	origCredentialsFunc = SetCredentialsFunc(TestCredentialsFunc)
}

// Put the original credentials func back
func RestoreCredentialsFunc() {
	SetCredentialsFunc(origCredentialsFunc)
}
