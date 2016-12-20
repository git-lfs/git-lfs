package lfsapi

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/git-lfs/git-lfs/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type authRequest struct {
	Test string
}

func TestAuthenticateHeaderAccess(t *testing.T) {
	tests := map[string]Access{
		"":                BasicAccess,
		"basic 123":       BasicAccess,
		"basic":           BasicAccess,
		"unknown":         BasicAccess,
		"NTLM":            NTLMAccess,
		"ntlm":            NTLMAccess,
		"NTLM 1 2 3":      NTLMAccess,
		"ntlm 1 2 3":      NTLMAccess,
		"NEGOTIATE":       NTLMAccess,
		"negotiate":       NTLMAccess,
		"NEGOTIATE 1 2 3": NTLMAccess,
		"negotiate 1 2 3": NTLMAccess,
	}

	for _, key := range authenticateHeaders {
		for value, expected := range tests {
			res := &http.Response{Header: make(http.Header)}
			res.Header.Set(key, value)
			t.Logf("%s: %s", key, value)
			assert.Equal(t, expected, getAuthAccess(res))
		}
	}
}

func TestDoWithAuthApprove(t *testing.T) {
	var called uint32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		atomic.AddUint32(&called, 1)
		w.Header().Set("Lfs-Authenticate", "Basic")

		body := &authRequest{}
		err := json.NewDecoder(req.Body).Decode(body)
		assert.Nil(t, err)
		assert.Equal(t, "Approve", body.Test)

		actual := req.Header.Get("Authorization")
		if len(actual) == 0 {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		expected := "Basic " + strings.TrimSpace(
			base64.StdEncoding.EncodeToString([]byte("user:pass")),
		)
		assert.Equal(t, expected, actual)
	}))
	defer srv.Close()

	creds := newMockCredentialHelper()
	c := &Client{
		Credentials: creds,
		Endpoints: NewEndpointFinder(gitEnv(map[string]string{
			"lfs.url": srv.URL,
		})),
	}

	assert.Equal(t, NoneAccess, c.Endpoints.AccessFor(srv.URL))

	body, err := Marshal(&authRequest{Test: "Approve"})
	require.Nil(t, err)

	req, err := http.NewRequest("GET", srv.URL, body)
	require.Nil(t, err)

	res, err := c.DoWithAuth("", req)
	require.Nil(t, err)

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.True(t, creds.IsApproved(Creds(map[string]string{
		"username": "user",
		"password": "pass",
		"path":     "",
		"protocol": "http",
		"host":     srv.Listener.Addr().String(),
	})))
	assert.Equal(t, BasicAccess, c.Endpoints.AccessFor(srv.URL))
	assert.EqualValues(t, 2, called)
}

func TestDoWithAuthReject(t *testing.T) {
	var called uint32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		atomic.AddUint32(&called, 1)
		w.Header().Set("Lfs-Authenticate", "Basic")

		body := &authRequest{}
		err := json.NewDecoder(req.Body).Decode(body)
		assert.Nil(t, err)
		assert.Equal(t, "Reject", body.Test)

		actual := req.Header.Get("Authorization")
		expected := "Basic " + strings.TrimSpace(
			base64.StdEncoding.EncodeToString([]byte("user:pass")),
		)

		if actual != expected {
			// Write http.StatuUnauthorized to force the credential
			// helper to reject the credentials
			w.WriteHeader(http.StatusUnauthorized)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer srv.Close()

	invalidCreds := Creds(map[string]string{
		"username": "user",
		"password": "wrong_pass",
		"path":     "",
		"protocol": "http",
		"host":     srv.Listener.Addr().String(),
	})

	creds := newMockCredentialHelper()

	creds.Approve(invalidCreds)
	assert.True(t, creds.IsApproved(invalidCreds))

	c := &Client{
		Credentials: creds,
		Endpoints: NewEndpointFinder(gitEnv(map[string]string{
			"lfs.url": srv.URL,
		})),
	}

	body, err := Marshal(&authRequest{Test: "Reject"})
	require.Nil(t, err)

	req, err := http.NewRequest("GET", srv.URL, body)
	require.Nil(t, err)

	res, err := c.DoWithAuth("", req)
	require.Nil(t, err)

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.False(t, creds.IsApproved(invalidCreds))
	assert.True(t, creds.IsApproved(Creds(map[string]string{
		"username": "user",
		"password": "pass",
		"path":     "",
		"protocol": "http",
		"host":     srv.Listener.Addr().String(),
	})))
	assert.EqualValues(t, 3, called)
}

type mockCredentialHelper struct {
	Approved map[string]Creds
}

func newMockCredentialHelper() *mockCredentialHelper {
	return &mockCredentialHelper{
		Approved: make(map[string]Creds),
	}
}

func (m *mockCredentialHelper) Fill(input Creds) (Creds, error) {
	if found, ok := m.Approved[credsToKey(input)]; ok {
		return found, nil
	}

	output := make(Creds)
	for key, value := range input {
		output[key] = value
	}
	if _, ok := output["username"]; !ok {
		output["username"] = "user"
	}
	output["password"] = "pass"
	return output, nil
}

func (m *mockCredentialHelper) Approve(creds Creds) error {
	m.Approved[credsToKey(creds)] = creds
	return nil
}

func (m *mockCredentialHelper) Reject(creds Creds) error {
	delete(m.Approved, credsToKey(creds))
	return nil
}

func (m *mockCredentialHelper) IsApproved(creds Creds) bool {
	if found, ok := m.Approved[credsToKey(creds)]; ok {
		return found["password"] == creds["password"]
	}
	return false
}

func credsToKey(creds Creds) string {
	var kvs []string
	for _, k := range []string{"protocol", "host", "path"} {
		kvs = append(kvs, fmt.Sprintf("%s:%s", k, creds[k]))
	}

	return strings.Join(kvs, " ")
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
	Remote        string
	SkipAuth      bool
}

func (c *getCredentialCheck) ExpectCreds() bool {
	return len(c.Protocol) > 0 || len(c.Host) > 0 || len(c.Username) > 0 ||
		len(c.Password) > 0 || len(c.Path) > 0
}

func TestGetCredentials(t *testing.T) {
	checks := []*getCredentialCheck{
		{
			Desc: "simple",
			Config: map[string]string{
				"lfs.url":                           "https://git-server.com",
				"lfs.https://git-server.com.access": "basic",
			},
			Method:   "GET",
			Href:     "https://git-server.com/foo",
			Protocol: "https",
			Host:     "git-server.com",
			Username: "git-server.com",
			Password: "monkey",
		},
		{
			Desc: "username in url",
			Config: map[string]string{
				"lfs.url": "https://user@git-server.com",
				"lfs.https://user@git-server.com.access": "basic",
			},
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
			Desc: "scheme mismatch",
			Config: map[string]string{
				"lfs.url": "https://git-server.com",
				"lfs.http://git-server.com/foo.access": "basic",
			},
			Method:   "GET",
			Href:     "http://git-server.com/foo",
			Protocol: "http",
			Host:     "git-server.com",
			Path:     "foo",
			Username: "git-server.com",
			Password: "monkey",
		},
		{
			Desc: "host mismatch",
			Config: map[string]string{
				"lfs.url": "https://git-server.com",
				"lfs.https://git-server2.com/foo.access": "basic",
			},
			Method:   "GET",
			Href:     "https://git-server2.com/foo",
			Protocol: "https",
			Host:     "git-server2.com",
			Path:     "foo",
			Username: "git-server2.com",
			Password: "monkey",
		},
		{
			Desc: "port mismatch",
			Config: map[string]string{
				"lfs.url": "https://git-server.com",
				"lfs.https://git-server.com:8080/foo.access": "basic",
			},
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
			Desc:   "git url auth",
			Remote: "origin",
			Config: map[string]string{
				"lfs.url":           "https://git-server.com",
				"remote.origin.url": "https://gituser:gitpass@git-server.com",
			},
			Method:        "GET",
			Href:          "https://git-server.com/foo",
			Authorization: "Basic " + strings.TrimSpace(base64.StdEncoding.EncodeToString([]byte("gituser:gitpass"))),
		},
		{
			Desc: "username in url",
			Config: map[string]string{
				"lfs.url": "https://user@git-server.com",
				"lfs.https://user@git-server.com.access": "basic",
			},
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
	}

	credHelper := &fakeCredentialFiller{}

	for _, check := range checks {
		t.Logf("Checking %q", check.Desc)
		ef := NewEndpointFinder(gitEnv(check.Config))

		req, err := http.NewRequest(check.Method, check.Href, nil)
		if err != nil {
			t.Errorf("[%s] %s", check.Desc, err)
			continue
		}

		for key, value := range check.Header {
			req.Header.Set(key, value)
		}

		creds, _, err := getCreds(credHelper, &noFinder{}, ef, check.Remote, req)
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

type fakeCredentialFiller struct{}

func (f *fakeCredentialFiller) Fill(input Creds) (Creds, error) {
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

func (f *fakeCredentialFiller) Approve(creds Creds) error {
	return errors.New("Not implemented")
}

func (f *fakeCredentialFiller) Reject(creds Creds) error {
	return errors.New("Not implemented")
}
