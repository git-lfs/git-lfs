package lfs

import (
	"encoding/base64"
	"net/http"
	"testing"
)

func TestGetCredentials(t *testing.T) {
	req, err := http.NewRequest("GET", "http://example.com", nil)
	if err != nil {
		t.Fatal(err)
	}

	creds, err := setRequestHeaders(req)
	if err != nil {
		t.Fatal(err)
	}

	if value := creds["username"]; value != "example.com" {
		t.Errorf("bad username: %s", value)
	}

	if value := creds["password"]; value != "monkey" {
		t.Errorf("bad username: %s", value)
	}

	expected := "Basic " + base64.URLEncoding.EncodeToString([]byte("example.com:monkey"))
	if value := req.Header.Get("Authorization"); value != expected {
		t.Errorf("Bad Authorization. Expected '%s', got '%s'", expected, value)
	}
}

func TestGetCredentialsWithPort(t *testing.T) {
	req, err := http.NewRequest("GET", "http://example.com:123", nil)
	if err != nil {
		t.Fatal(err)
	}

	creds, err := setRequestHeaders(req)
	if err != nil {
		t.Fatal(err)
	}

	if value := creds["username"]; value != "example.com:123" {
		t.Errorf("bad username: %s", value)
	}

	if value := creds["password"]; value != "monkey" {
		t.Errorf("bad username: %s", value)
	}

	expected := "Basic " + base64.URLEncoding.EncodeToString([]byte("example.com:123:monkey"))
	if value := req.Header.Get("Authorization"); value != expected {
		t.Errorf("Bad Authorization. Expected '%s', got '%s'", expected, value)
	}
}

func TestGetCredentialsWithAuthorization(t *testing.T) {
	req, err := http.NewRequest("GET", "http://example.com", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "")

	creds, err := setRequestHeaders(req)
	if err != nil {
		t.Fatal(err)
	}

	if creds != nil {
		t.Errorf("Unexpected credentials: %v", creds)
	}

	if value := req.Header.Get("Authorization"); value != "" {
		t.Errorf("Unexpected authorization: %s", value)
	}
}

func init() {
	execCreds = func(input Creds, subCommand string) (credentialFetcher, error) {
		return &testCredentialFetcher{input}, nil
	}
}

type testCredentialFetcher struct {
	Creds Creds
}

func (c *testCredentialFetcher) Credentials() Creds {
	c.Creds["username"] = c.Creds["host"]
	c.Creds["password"] = "monkey"
	return c.Creds
}
